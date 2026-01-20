package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

type VimeoStrategy struct {
	Client *http.Client
}

func NewVimeoStrategy() *VimeoStrategy {
	return &VimeoStrategy{
		Client: &http.Client{
			Timeout: 60 * time.Minute, // Long timeout for large downloads
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

// VimeoConfig represents the JSON structure found in Vimeo player page
type VimeoConfig struct {
	Request struct {
		Files struct {
			Progressive []struct {
				Profile interface{} `json:"profile"` // Can be int or string
				Width   int         `json:"width"`
				Height  int         `json:"height"`
				Mime    string      `json:"mime"`
				URL     string      `json:"url"`
				Quality string      `json:"quality"`
			} `json:"progressive"`
			HLS struct {
				Cdn        string `json:"cdn"`
				Url        string `json:"url"`
				HdUrl      string `json:"hd_url"`
				DefaultCdn string `json:"default_cdn"`
			} `json:"hls"`
		} `json:"files"`
	} `json:"request"`
}

// Download attempts to download a Vimeo video by scraping the player page for the config
// and extracting direct progressive MP4 links.
func (v *VimeoStrategy) Download(ctx context.Context, url string, outputFile string) error {
	log.Info().Str("url", url).Msg("Starting VimeoStrategy download")

	// 1. Get the player page content
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Mimic a modern browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://vimeo.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")

	resp, err := v.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	bodyString := string(bodyBytes)

	// 2. Extract the config JSON
	// Look for: var config = {...}; or window.vimeo.clip_page_config = ...
	// The regex captures the JSON object inside the assignment
	var configJSON string

	// Pattern 1: var config = {...}
	re1 := regexp.MustCompile(`var config = (\{.*?\});`)
	matches1 := re1.FindStringSubmatch(bodyString)
	if len(matches1) >= 2 {
		configJSON = matches1[1]
	} else {
		// Pattern 2: window.vimeo.clip_page_config = {...}
		re2 := regexp.MustCompile(`window\.vimeo\.clip_page_config = (\{.*?\});`)
		matches2 := re2.FindStringSubmatch(bodyString)
		if len(matches2) >= 2 {
			configJSON = matches2[1]
		}
	}

	if configJSON == "" {
		return fmt.Errorf("could not find vimeo config in page source")
	}

	var config VimeoConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		// Try to unmarshal into a generic map first if specific struct fails, for debugging
		return fmt.Errorf("failed to parse config json: %w", err)
	}

	// 3. Find the best progressive stream
	files := config.Request.Files.Progressive
	if len(files) == 0 {
		return fmt.Errorf("no progressive files found in vimeo config")
	}

	// Sort by width (resolution) descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].Width > files[j].Width
	})

	bestFile := files[0]
	log.Info().
		Int("width", bestFile.Width).
		Str("quality", bestFile.Quality).
		Str("url", bestFile.URL).
		Msg("Found best vimeo stream via custom strategy")

	// 4. Download the file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	videoReq, err := http.NewRequestWithContext(ctx, "GET", bestFile.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create video request: %w", err)
	}
	videoReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	videoResp, err := v.Client.Do(videoReq)
	if err != nil {
		return fmt.Errorf("failed to download video stream: %w", err)
	}
	defer videoResp.Body.Close()

	if videoResp.StatusCode != http.StatusOK {
		return fmt.Errorf("video download returned status: %d", videoResp.StatusCode)
	}

	_, err = io.Copy(out, videoResp.Body)
	if err != nil {
		return fmt.Errorf("failed to save video file: %w", err)
	}

	return nil
}
func (v *VimeoStrategy) TestURL(ctx context.Context, url string) ([]byte, error) {
	// temporary code
	cmd := exec.CommandContext(ctx, "vimeo-dl", "-i", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check vimeo url: %w", err)
	}
	return output, nil
}
