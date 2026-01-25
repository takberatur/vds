package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type LuxStrategy struct {
}

func NewLuxStrategy() *LuxStrategy {
	return &LuxStrategy{}
}

func (s *LuxStrategy) Name() string {
	return "lux"
}

func (s *LuxStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	// lux -i URL
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// -i displays info
	cmd := exec.CommandContext(ctx, "lux", "-i", url)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Info().Str("url", url).Msg("Fetching video info with lux")
	if err := cmd.Run(); err != nil {
		// Lux sometimes prints errors to stdout, but might have printed the info/url before failing
		output := stdout.String()

		// Try to recover info from stdout even if it failed (common for TikTok 403)
		if recoveredInfo := s.tryRecoverInfo(output, url); recoveredInfo != nil {
			log.Info().Msg("Recovered video info from failed lux execution")
			return recoveredInfo, nil
		}

		// Lux sometimes prints errors to stdout
		return nil, fmt.Errorf("lux failed: %w, stderr: %s, stdout: %s", err, stderr.String(), stdout.String())
	}
	output := stdout.String()
	info := parseLuxOutput(output)
	if info == nil {
		log.Error().Str("output", output).Msg("Failed to parse lux output")
		return nil, fmt.Errorf("failed to parse lux output")
	}

	info.WebpageURL = url
	info.Extractor = "lux"
	
	// If parseLuxOutput didn't find a direct URL, we can't use it for direct download.
	// We used to set "lux://" here, but that causes the worker to fail because it expects a real URL.
	// If we return nil/error here, the factory will fallback to Chromedp, which is what we want.
	if info.DownloadURL == "" || strings.HasPrefix(info.DownloadURL, "lux://") {
		// Try to see if we can extract it from the output lines manually if parser missed it
		// Reuse tryRecoverInfo logic which does a good job finding URLs
		if recovered := s.tryRecoverInfo(output, url); recovered != nil && recovered.DownloadURL != "" {
			info.DownloadURL = recovered.DownloadURL
			info.Formats = recovered.Formats
		} else {
			// Fail so we fallback to Chromedp
			log.Warn().Str("url", url).Msg("Lux strategy failed to extract direct URL, falling back")
			return nil, fmt.Errorf("lux failed to extract direct URL")
		}
	}

	return info, nil
}

func (s *LuxStrategy) tryRecoverInfo(output string, originalURL string) *VideoInfo {
	lines := strings.Split(output, "\n")
	var title string
	var downloadURL string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		}

		// Check for URL in error context OR normal output
		// Example: "https://v16-webapp-prime.tiktok.com/... request error: HTTP 403"
		// Or sometimes just the URL on a line if it's listing streams
		if strings.HasPrefix(line, "https://") && (strings.Contains(line, "tiktok.com") || strings.Contains(line, "googlevideo.com") || strings.Contains(line, "akamaized.net") || strings.Contains(line, "instagram.com") || strings.Contains(line, "fbcdn.net")) {
			// Basic check to avoid capturing the input URL if it's just repeating "Downloading ..."
			if !strings.Contains(line, "Downloading") && line != originalURL {
				parts := strings.Split(line, " ")
				if len(parts) > 0 && strings.HasPrefix(parts[0], "https://") {
					candidate := parts[0]
					// Filter out some common non-video URLs if needed, but for now accept it
					downloadURL = candidate
				}
			}
		}
	}

	if downloadURL != "" {
		if title == "" {
			title = "Video (Recovered)" // Fallback title
		}

		return &VideoInfo{
			Title:       title,
			WebpageURL:  originalURL,
			Extractor:   "lux",
			DownloadURL: downloadURL,
			Formats:     []FormatInfo{{URL: downloadURL, Ext: "mp4"}},
		}
	}
	return nil
}

func parseLuxOutput(output string) *VideoInfo {
	lines := strings.Split(output, "\n")
	info := &VideoInfo{}
	
	// We reuse tryRecoverInfo logic effectively, but here we parse formally
	// Since tryRecoverInfo is robust, we can actually just rely on it or similar logic.
	// But let's keep basic title parsing here.
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Title:") {
			info.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		}
	}

	if info.Title == "" {
		return nil
	}
	return info
}
