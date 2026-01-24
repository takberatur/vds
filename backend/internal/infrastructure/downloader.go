package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
)

type VideoInfo struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Duration    *float64     `json:"duration"` // seconds
	Thumbnail   string       `json:"thumbnail"`
	WebpageURL  string       `json:"webpage_url"`
	Extractor   string       `json:"extractor"` // youtube, tiktok, etc.
	Filename    string       `json:"filename,omitempty"`
	Filesize    *int64       `json:"filesize,omitempty"`
	DownloadURL string            `json:"url,omitempty"` // Direct link if available
	Cookies     map[string]string `json:"cookies,omitempty"` // Cookies required for download
	Formats     []FormatInfo      `json:"formats,omitempty"`
}

type FormatInfo struct {
	URL      string   `json:"url"`
	Filesize *int64   `json:"filesize,omitempty"`
	FormatID string   `json:"format_id,omitempty"`
	Acodec   string   `json:"acodec,omitempty"`
	Vcodec   string   `json:"vcodec,omitempty"`
	Ext      string   `json:"ext,omitempty"`
	Height   *int     `json:"height,omitempty"`
	Width    *int     `json:"width,omitempty"`
	Tbr      *float64 `json:"tbr,omitempty"`
}

type DownloaderClient interface {
	GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error)
	DownloadVideo(ctx context.Context, url string) (*VideoInfo, error)
	DownloadToPath(ctx context.Context, url string, formatID string, outputPath string) error
}

type ytDlpClient struct {
	executablePath string
}

func NewDownloaderClient() DownloaderClient {
	return &ytDlpClient{
		executablePath: "yt-dlp",
	}
}

func (c *ytDlpClient) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 25*time.Second)
	defer cancel()

	args := []string{
		"--js-runtimes", "node",
		"--dump-json",
		"--no-playlist",
		"--no-check-certificate",
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
	}

	if strings.Contains(url, "rumble.com") {
		args = append(args, "--referer", "https://rumble.com/")
		args = append(args, "--add-header", "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		args = append(args, "--add-header", "Accept-Language: en-US,en;q=0.9")
	}

	if strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly") {
		args = append(args, "--referer", "https://www.dailymotion.com/")
	}

	args = append(args, url)

	cmd := exec.CommandContext(subCtx, c.executablePath, args...)
	output, err := cmd.Output()
	if err != nil {
		log.Error().Str("url", url).Err(err).Msg("yt-dlp failed")
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	if info.DownloadURL == "" && len(info.Formats) > 0 {
		if best := pickBestFormat(info.Formats); best != nil {
			info.DownloadURL = best.URL
			if info.Filesize == nil && best.Filesize != nil {
				info.Filesize = best.Filesize
			}
		}
	}

	return &info, nil
}

func isHLSFormat(f FormatInfo) bool {
	return strings.Contains(f.URL, ".m3u8")
}

func pickBestFormat(formats []FormatInfo) *FormatInfo {
	if len(formats) == 0 {
		return nil
	}

	var bestNonHLS *FormatInfo
	for i := range formats {
		f := &formats[i]
		if f.URL == "" || isHLSFormat(*f) {
			continue
		}
		if bestNonHLS == nil {
			bestNonHLS = f
			continue
		}

		fHeight := 0
		if f.Height != nil {
			fHeight = *f.Height
		}
		bestHeight := 0
		if bestNonHLS.Height != nil {
			bestHeight = *bestNonHLS.Height
		}

		if fHeight > bestHeight {
			bestNonHLS = f
			continue
		}

		fTbr := 0.0
		if f.Tbr != nil {
			fTbr = *f.Tbr
		}
		bestTbr := 0.0
		if bestNonHLS.Tbr != nil {
			bestTbr = *bestNonHLS.Tbr
		}

		if fHeight == bestHeight && fTbr > bestTbr {
			bestNonHLS = f
			continue
		}

		fSize := int64(0)
		if f.Filesize != nil {
			fSize = *f.Filesize
		}
		bestSize := int64(0)
		if bestNonHLS.Filesize != nil {
			bestSize = *bestNonHLS.Filesize
		}

		if fHeight == bestHeight && fTbr == bestTbr && fSize > bestSize {
			bestNonHLS = f
			continue
		}
	}

	if bestNonHLS != nil {
		return bestNonHLS
	}

	var best *FormatInfo
	for i := range formats {
		f := &formats[i]
		if f.URL == "" {
			continue
		}
		if best == nil {
			best = f
			continue
		}

		fHeight := 0
		if f.Height != nil {
			fHeight = *f.Height
		}
		bestHeight := 0
		if best.Height != nil {
			bestHeight = *best.Height
		}

		if fHeight > bestHeight {
			best = f
			continue
		}

		fTbr := 0.0
		if f.Tbr != nil {
			fTbr = *f.Tbr
		}
		bestTbr := 0.0
		if best.Tbr != nil {
			bestTbr = *best.Tbr
		}

		if fHeight == bestHeight && fTbr > bestTbr {
			best = f
			continue
		}

		fSize := int64(0)
		if f.Filesize != nil {
			fSize = *f.Filesize
		}
		bestSize := int64(0)
		if best.Filesize != nil {
			bestSize = *best.Filesize
		}

		if fHeight == bestHeight && fTbr == bestTbr && fSize > bestSize {
			best = f
			continue
		}
	}
	return best
}

func (c *ytDlpClient) DownloadVideo(ctx context.Context, url string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 25*time.Second)
	defer cancel()

	info, err := c.GetVideoInfo(subCtx, url)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (c *ytDlpClient) DownloadToPath(ctx context.Context, url string, formatID string, outputPath string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 10*time.Minute)
	defer cancel()

	args := []string{
		"--no-playlist",
		"--no-check-certificate",
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"-o", outputPath,
	}

	if formatID != "" {
		args = append(args, "-f", formatID)
	}

	if strings.Contains(url, "rumble.com") {
		args = append(args, "--referer", "https://rumble.com/")
	}

	if strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly") {
		args = append(args, "--referer", "https://www.dailymotion.com/")
	}

	args = append(args, url)

	cmd := exec.CommandContext(subCtx, c.executablePath, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp download failed: %w", err)
	}

	return nil
}
