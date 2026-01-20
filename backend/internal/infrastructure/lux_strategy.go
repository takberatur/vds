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
	// Use a special scheme to indicate lux should be used for download
	info.DownloadURL = "lux://" + url

	// Add a default format if none found (lux -i usually lists streams)
	if len(info.Formats) == 0 {
		info.Formats = []FormatInfo{
			{
				URL: info.DownloadURL,
				Ext: "mp4", // Assume mp4
			},
		}
	}

	return info, nil
}

func parseLuxOutput(output string) *VideoInfo {
	lines := strings.Split(output, "\n")
	info := &VideoInfo{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Title:") {
			info.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		}
		// Parsing streams is complex and fragile from text output.
		// For now, we mainly need the Title to verify it works.
		// If we need streams, we'd need a more robust parser.
	}

	if info.Title == "" {
		return nil
	}
	return info
}
