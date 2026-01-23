package infrastructure

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type VimeoStrategy struct {
	Client *http.Client
}

func NewVimeoStrategy() *VimeoStrategy {
	return &VimeoStrategy{
		Client: &http.Client{
			Timeout: 15 * time.Minute, // Reduced timeout to prevent indefinite hangs
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Skip certificate verification
				},
			},
		},
	}
}

// VideoInfo struct is already defined in downloader.go

// ExtractVideoID extracts the Vimeo video ID from various URL formats
func (v *VimeoStrategy) ExtractVideoID(url string) (string, error) {
	patterns := []string{
		`vimeo\.com/(\d+)`,
		`vimeo\.com/video/(\d+)`,
		`player\.vimeo\.com/video/(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("invalid Vimeo URL: unable to extract video ID")
}

// fetchVideoInfo fetches video metadata from Vimeo's config endpoint
func (v *VimeoStrategy) fetchVideoInfo(videoID string) (*VideoInfo, error) {
	configURL := fmt.Sprintf("https://player.vimeo.com/video/%s/config", videoID)

	req, err := http.NewRequest("GET", configURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Mimic a modern browser to avoid 403 Forbidden
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://vimeo.com/")
	req.Header.Set("Origin", "https://vimeo.com")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := v.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch video config: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	info := &VideoInfo{}

	if video, ok := config["video"].(map[string]interface{}); ok {
		if title, ok := video["title"].(string); ok {
			info.Title = title
		}
		if duration, ok := video["duration"].(float64); ok {
			info.Duration = &duration
		}
	}

	if request, ok := config["request"].(map[string]interface{}); ok {
		if files, ok := request["files"].(map[string]interface{}); ok {
			// Prioritize progressive (direct MP4)
			if progressive, ok := files["progressive"].([]interface{}); ok && len(progressive) > 0 {
				bestQuality := progressive[len(progressive)-1].(map[string]interface{})
				if url, ok := bestQuality["url"].(string); ok {
					info.DownloadURL = url
				}
			}

			// If no progressive, check for HLS (which might be handled by player but not directly downloadable as single file)
			// But at least we know the video exists.
			if info.DownloadURL == "" {
				if hls, ok := files["hls"].(map[string]interface{}); ok {
					// We can't use HLS URL as direct download URL for the client as it's a playlist.
					// But we can log it or maybe use it if we had a proper HLS downloader.
					// For now, if only HLS exists, we return error to force fallback to yt-dlp which handles HLS.
					if _, ok := hls["url"].(string); ok {
						return nil, fmt.Errorf("video is HLS-only, falling back to yt-dlp")
					}
				}
			}
		}
	}

	if info.DownloadURL == "" {
		return nil, fmt.Errorf("no download URL found - video may be private or restricted")
	}

	return info, nil
}

// GetVideoInfo implements the DownloaderStrategy interface
func (v *VimeoStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	videoID, err := v.ExtractVideoID(url)
	if err != nil {
		return nil, err
	}

	return v.fetchVideoInfo(videoID)
}

// Name implements the DownloaderStrategy interface
func (v *VimeoStrategy) Name() string {
	return "vimeo-custom"
}

// GetDirectURL is the adapter for download_handler.go
// It uses the new ExtractVideoID and fetchVideoInfo logic
func (v *VimeoStrategy) GetDirectURL(ctx context.Context, url string) (string, error) {
	log.Info().Str("url", url).Msg("Fetching Vimeo direct URL via config endpoint")

	videoID, err := v.ExtractVideoID(url)
	if err != nil {
		return "", err
	}

	info, err := v.fetchVideoInfo(videoID)
	if err != nil {
		return "", err
	}

	log.Info().Str("direct_url", info.DownloadURL).Msg("Successfully extracted Vimeo direct URL")
	return info.DownloadURL, nil
}

// Download saves the video to a file using the extracted direct URL
func (v *VimeoStrategy) Download(ctx context.Context, url string, outputFile string) error {
	// Use GetDirectURL to resolve the final MP4 URL
	directURL, err := v.GetDirectURL(ctx, url)
	if err != nil {
		return err
	}

	return v.VimeoDownload(ctx, directURL, outputFile)
}

// VimeoDownload performs the actual HTTP download with progress tracking support via Context
func (v *VimeoStrategy) VimeoDownload(ctx context.Context, url, outputPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := v.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	downloaded := int64(0)
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
					return fmt.Errorf("failed to write to file: %w", writeErr)
				}
				downloaded += int64(n)
			}
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("download error: %w", err)
			}
		}
	}
}

func (v *VimeoStrategy) SanitizeFilename(name string) string {
	invalid := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	name = invalid.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
