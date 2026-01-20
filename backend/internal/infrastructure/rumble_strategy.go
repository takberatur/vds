package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

type RumbleStrategy struct {
}

func NewRumbleStrategy() *RumbleStrategy {
	return &RumbleStrategy{}
}

func (s *RumbleStrategy) Name() string {
	return "rumble-custom"
}

func (s *RumbleStrategy) GetVideoInfo(ctx context.Context, videoURL string) (*VideoInfo, error) {
	log.Info().Str("url", videoURL).Msg("Starting Rumble strategy")

	// Step 1: Get Embed ID via OEmbed
	embedID, title, thumbnail, err := s.getEmbedID(ctx, videoURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get embed ID")
		// Fallback: try to guess embed ID or proceed with direct page if possible?
		// For now, let's fail as the strategy relies on embed ID
		return nil, err
	}

	log.Info().Str("embed_id", embedID).Msg("Found Rumble embed ID")
	embedURL := fmt.Sprintf("https://rumble.com/embed/%s/", embedID)

	// Step 2: Use Chromedp to intercept network requests
	info, err := s.interceptVideo(ctx, embedURL, videoURL)
	if err != nil {
		return nil, err
	}

	// Fill in OEmbed info if missing
	if info.Title == "" {
		info.Title = title
	}
	if info.Thumbnail == "" {
		info.Thumbnail = thumbnail
	}

	return info, nil
}

func (s *RumbleStrategy) getEmbedID(ctx context.Context, videoURL string) (string, string, string, error) {
	oembedURL := fmt.Sprintf("https://rumble.com/api/Media/oembed.json?url=%s", url.QueryEscape(videoURL))
	req, err := http.NewRequestWithContext(ctx, "GET", oembedURL, nil)
	if err != nil {
		return "", "", "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("oembed api returned status: %d", resp.StatusCode)
	}

	var data struct {
		HTML         string `json:"html"`
		Title        string `json:"title"`
		ThumbnailUrl string `json:"thumbnail_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", "", err
	}

	// Regex to extract embed ID: src="https://rumble.com/embed/([a-zA-Z0-9_]+)/?
	re := regexp.MustCompile(`src="https://rumble\.com/embed/([a-zA-Z0-9_]+)/?`)
	match := re.FindStringSubmatch(data.HTML)
	if len(match) < 2 {
		return "", "", "", fmt.Errorf("could not find embed id in html")
	}

	return match[1], data.Title, data.ThumbnailUrl, nil
}

func (s *RumbleStrategy) interceptVideo(ctx context.Context, embedURL, originalURL string) (*VideoInfo, error) {
	// Create allocator options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create context
	ctx, cancel = chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(s string, args ...interface{}) {
			// Suppress noisy logs
			msg := fmt.Sprintf(s, args...)
			if !strings.Contains(msg, "cookiePartitionKey") {
				log.Debug().Msg(msg)
			}
		}),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var foundVideoURL string
	var foundTitle string
	var mu sync.Mutex
	done := make(chan struct{})

	// Listen for network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			resp := ev.Response
			urlStr := resp.URL

			// Check for direct video files
			if (strings.HasSuffix(urlStr, ".mp4") || strings.Contains(urlStr, ".mp4?")) ||
				(strings.HasSuffix(urlStr, ".m3u8") || strings.Contains(urlStr, ".m3u8?")) {

				mu.Lock()
				if foundVideoURL == "" {
					log.Info().Str("url", urlStr).Msg("Found video URL via network interception")
					foundVideoURL = urlStr
					close(done)
				}
				mu.Unlock()
			}

			// Check for JSON metadata (like /embedJS/)
			if strings.Contains(urlStr, "/embedJS/") && strings.Contains(resp.MimeType, "application/json") {
				// We could try to fetch body here, but for now capturing the MP4 directly is easier/faster
				// if the player loads it.
				// However, fetching body requires a separate action which is async.
			}
		}
	})

	log.Info().Str("url", embedURL).Msg("Navigating to embed page")

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(embedURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for hydration
		// Try to click play button if exists
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Select .bigPlayUI and click
			// We use a short timeout/check because it might autoplay or not exist
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			err := chromedp.Click(`.bigPlayUI`, chromedp.ByQuery).Do(ctx)
			if err != nil {
				// Ignore error, button might not exist
				log.Debug().Err(err).Msg("Play button not clicked (might not exist)")
			} else {
				log.Info().Msg("Clicked play button")
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp navigation failed: %w", err)
	}

	// Wait for video to be found or timeout
	select {
	case <-done:
		// Success
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for video url")
	case <-time.After(15 * time.Second):
		// Give it some time after navigation
		mu.Lock()
		if foundVideoURL == "" {
			mu.Unlock()
			return nil, fmt.Errorf("timeout waiting for video url (custom)")
		}
		mu.Unlock()
	}

	// Get basic info if possible (Title/Thumbnail might need another scrape or passed from OEmbed)
	// For now, we reuse the OEmbed call if we want title, but let's see if we can get it from page title
	var pageTitle string
	if err := chromedp.Run(ctx, chromedp.Title(&pageTitle)); err == nil {
		foundTitle = pageTitle
	}

	return &VideoInfo{
		Title:       foundTitle,
		DownloadURL: foundVideoURL,
		WebpageURL:  originalURL,
		Extractor:   "rumble-custom",
	}, nil
}
