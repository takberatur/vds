package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

type ChromedpStrategy struct {
}

func NewChromedpStrategy() *ChromedpStrategy {
	return &ChromedpStrategy{}
}

func (s *ChromedpStrategy) Name() string {
	return "chromedp"
}

// GetCookies navigates to the URL and retrieves cookies as a formatted string
func (s *ChromedpStrategy) GetCookies(ctx context.Context, url string) (string, error) {
	// Create allocator options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var cookies []*network.Cookie
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Wait for cookies to be set
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)

	if err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	var cookieBuilder strings.Builder
	for i, cookie := range cookies {
		if i > 0 {
			cookieBuilder.WriteString("; ")
		}
		cookieBuilder.WriteString(fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	return cookieBuilder.String(), nil
}

func (s *ChromedpStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	// Create allocator options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("ignore-certificate-errors", true), // Ignore SSL errors
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Create context
	// Suppress cookie errors by using a custom logger
	ctx, cancel = chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(s string, args ...interface{}) {
			msg := fmt.Sprintf(s, args...)
			// Filter out cookie partition key errors which are noisy and harmless
			if !strings.Contains(msg, "cookiePartitionKey") && !strings.Contains(msg, "CookiePartitionKey") && !strings.Contains(msg, "partitionKey") {
				log.Debug().Msg(msg)
			}
		}),
		chromedp.WithErrorf(func(s string, args ...interface{}) {
			msg := fmt.Sprintf(s, args...)
			if !strings.Contains(msg, "cookiePartitionKey") && !strings.Contains(msg, "CookiePartitionKey") && !strings.Contains(msg, "partitionKey") {
				log.Warn().Msg(msg)
			}
		}),
	)
	defer cancel()

	// Set timeout for the whole operation
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var htmlContent string

	log.Info().Str("url", url).Msg("Navigating with Chromedp")

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(10*time.Second), // Increase wait time for JS to execute/hydrate and cloudflare challenge
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp navigation failed: %w", err)
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("goquery parse failed: %w", err)
	}

	// Try to find JSON-LD
	var videoInfo *VideoInfo

	// Debug: Check title to see if it's Cloudflare or actual page
	pageTitle := strings.TrimSpace(doc.Find("title").Text())
	log.Info().Str("url", url).Str("page_title", pageTitle).Msg("Chromedp page title")

	jsonLdScripts := doc.Find("script[type='application/ld+json']")
	log.Info().Int("count", jsonLdScripts.Length()).Msg("Found JSON-LD scripts")

	jsonLdScripts.Each(func(i int, s *goquery.Selection) {
		if videoInfo != nil {
			return
		}

		jsonText := s.Text()

		// Try parsing as map first
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonText), &data); err == nil {
			if processJsonLD(data, &videoInfo) {
				return
			}

			// Check graph
			if graph, ok := data["@graph"].([]interface{}); ok {
				for _, item := range graph {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if processJsonLD(itemMap, &videoInfo) {
							return
						}
					}
				}
			}
			return
		}

		// Try parsing as array
		var dataArray []interface{}
		if err := json.Unmarshal([]byte(jsonText), &dataArray); err == nil {
			for _, item := range dataArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if processJsonLD(itemMap, &videoInfo) {
						return
					}
				}
			}
		}
	})

	if videoInfo == nil {
		log.Info().Msg("JSON-LD failed, trying meta tags fallback")
		videoInfo = extractFromMetaTags(doc)
	}

	if videoInfo != nil {
		videoInfo.WebpageURL = url
		videoInfo.Extractor = "chromedp"
		// Fallback for ID if missing
		if videoInfo.ID == "" {
			// Extract ID from URL if possible, or just use hash
			videoInfo.ID = "rumble-video"
		}
		return videoInfo, nil
	}

	return nil, fmt.Errorf("failed to extract video info with chromedp")
}

func (s *ChromedpStrategy) IsVideoURL(url string) bool {
	videoExtensions := []string{
		".mp4", ".webm", ".mkv", ".flv", ".avi", ".mov",
		".m3u8", ".mpd", ".ts", ".m4v",
	}

	for _, ext := range videoExtensions {
		if strings.Contains(strings.ToLower(url), ext) {
			return true
		}
	}

	patterns := []string{
		"/video/", "/videos/", "/stream/", "/videoplayback",
		"googlevideo.com", "video.twimg.com",
	}

	for _, pattern := range patterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}

	return false
}

func processJsonLD(data map[string]interface{}, info **VideoInfo) bool {
	typeStr, ok := data["@type"].(string)
	if !ok || typeStr != "VideoObject" {
		return false
	}

	contentUrl, _ := data["contentUrl"].(string)
	embedUrl, _ := data["embedUrl"].(string)

	if contentUrl == "" && embedUrl == "" {
		return false
	}

	title, _ := data["name"].(string)
	// description, _ := data["description"].(string)
	thumbnailUrl, _ := data["thumbnailUrl"].(string)

	// Handle thumbnailUrl which can be array
	if thumbnailUrl == "" {
		if thumbs, ok := data["thumbnailUrl"].([]interface{}); ok && len(thumbs) > 0 {
			thumbnailUrl, _ = thumbs[0].(string)
		}
	}

	// durationStr, _ := data["duration"].(string) // ISO 8601 duration

	// Basic parsing
	*info = &VideoInfo{
		Title:       title,
		Thumbnail:   thumbnailUrl,
		DownloadURL: contentUrl,
	}

	if contentUrl != "" {
		(*info).Formats = []FormatInfo{
			{
				URL: contentUrl,
				Ext: "mp4", // Assumption for now
			},
		}
	}

	return true
}

func extractFromMetaTags(doc *goquery.Document) *VideoInfo {
	var info VideoInfo

	// Try standard OpenGraph video
	info.DownloadURL = doc.Find("meta[property='og:video']").AttrOr("content", "")
	if info.DownloadURL == "" {
		info.DownloadURL = doc.Find("meta[property='og:video:secure_url']").AttrOr("content", "")
	}
	if info.DownloadURL == "" {
		info.DownloadURL = doc.Find("meta[property='og:video:url']").AttrOr("content", "")
	}
	if info.DownloadURL == "" {
		info.DownloadURL = doc.Find("meta[itemprop='contentUrl']").AttrOr("content", "")
	}

	// If still empty, try to find video tag src
	if info.DownloadURL == "" {
		info.DownloadURL = doc.Find("video source").AttrOr("src", "")
	}
	if info.DownloadURL == "" {
		info.DownloadURL = doc.Find("video").AttrOr("src", "")
	}

	if info.DownloadURL == "" {
		return nil
	}

	info.Title = doc.Find("meta[property='og:title']").AttrOr("content", "")
	if info.Title == "" {
		info.Title = doc.Find("title").Text()
	}

	info.Thumbnail = doc.Find("meta[property='og:image']").AttrOr("content", "")

	info.Formats = []FormatInfo{
		{
			URL: info.DownloadURL,
			Ext: "mp4",
		},
	}

	return &info
}
