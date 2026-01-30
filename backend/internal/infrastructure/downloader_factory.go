package infrastructure

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/scrapper"
)

type DownloaderStrategy interface {
	GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error)
	Name() string
}

type YtDlpStrategy struct {
	client *ytDlpClient
}

func (s *YtDlpStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	return s.client.GetVideoInfo(ctx, url)
}

func (s *YtDlpStrategy) Name() string {
	return "yt-dlp"
}

type YoutubeGoStrategy struct {
	client youtube.Client
}

func (s *YoutubeGoStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	_ = ctx
	video, err := s.client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	formats := video.Formats.WithAudioChannels()
	var bestURL string
	if len(formats) > 0 {
		bestURL = formats[0].URL
	}

	dur := video.Duration.Seconds()
	return &VideoInfo{
		ID:          video.ID,
		Title:       video.Title,
		Duration:    &dur,
		Thumbnail:   video.Thumbnails[0].URL, // Pick first thumbnail
		WebpageURL:  "https://www.youtube.com/watch?v=" + video.ID,
		Extractor:   "youtube",
		DownloadURL: bestURL,
	}, nil
}

func (s *YoutubeGoStrategy) Name() string {
	return "kkdai/youtube"
}

type YoutubeCustomStrategy struct {
	downloader *YoutubeDownloader
}

func (s *YoutubeCustomStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	_ = ctx
	details, err := s.downloader.GetVideoDetails(url)
	if err != nil {
		return nil, err
	}

	video, err := s.downloader.client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	dur := video.Duration.Seconds()
	thumb := ""
	if len(video.Thumbnails) > 0 {
		thumb = video.Thumbnails[0].URL
	}
	if len(video.Thumbnails) > 0 {
		thumb = video.Thumbnails[len(video.Thumbnails)-1].URL
	}

	formats := make([]FormatInfo, 0, len(details.Formats))
	for _, f := range details.Formats {
		formatID := fmt.Sprintf("%d", f.ItagNo)
		ext := "mp4"
		lower := strings.ToLower(f.MimeType)
		if strings.Contains(lower, "webm") {
			ext = "webm"
		} else if strings.Contains(lower, "mp4") {
			ext = "mp4"
		}
		formats = append(formats, FormatInfo{
			URL:      "",
			FormatID: formatID,
			Ext:      ext,
			Acodec:   "",
			Vcodec:   "",
		})
	}

	return &VideoInfo{
		ID:          video.ID,
		Title:       details.Title,
		Duration:    &dur,
		Thumbnail:   thumb,
		WebpageURL:  "https://www.youtube.com/watch?v=" + video.ID,
		Extractor:   "youtube",
		DownloadURL: "",
		Formats:     formats,
	}, nil
}

func (s *YoutubeCustomStrategy) Name() string {
	return "youtube-custom"
}

type YTDownStrategy struct {
	client *scrapper.YTDownService
}

func (s *YTDownStrategy) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("YTDOWN_DISABLED")), "true") {
		return nil, fmt.Errorf("ytdown disabled")
	}

	resp, _, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	best := pickBestYTDownItem(resp.API.MediaItems)
	downloadURL := ""
	thumb := strings.TrimSpace(resp.API.ImagePreview)
	if thumb == "" {
		thumb = strings.TrimSpace(resp.API.PreviewURL)
	}

	var dur *float64
	if best != nil {
		downloadURL = strings.TrimSpace(best.MediaURL)
		if downloadURL == "" {
			downloadURL = strings.TrimSpace(best.MediaPreviewURL)
		}
		if seconds := parseDurationSeconds(best.MediaDuration); seconds > 0 {
			dur = &seconds
		}
	}

	formats := make([]FormatInfo, 0, len(resp.API.MediaItems))
	seen := make(map[string]struct{}, 64)
	for _, item := range resp.API.MediaItems {
		u := strings.TrimSpace(item.MediaURL)
		if u == "" {
			u = strings.TrimSpace(item.MediaPreviewURL)
		}
		if u == "" {
			continue
		}
		ext := strings.ToLower(strings.TrimSpace(item.MediaExtension))
		if ext == "" {
			ext = "mp4"
		}
		height := parseHeightFromString(item.MediaQuality)
		if height == 0 {
			if res, ok := item.MediaRes.(string); ok {
				height = parseHeightFromString(res)
			}
		}
		formatID := "best"
		if height > 0 {
			formatID = fmt.Sprintf("%dp", height)
		}

		key := formatID + "|" + ext
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		var hPtr *int
		if height > 0 {
			h := height
			hPtr = &h
		}

		acodec := ""
		vcodec := ""
		lt := strings.ToLower(item.Type)
		if strings.Contains(lt, "audio") {
			vcodec = "none"
			if ext == "mp4" {
				acodec = "aac"
			} else if ext == "webm" {
				acodec = "opus"
			}
		} else {
			if ext == "mp4" {
				vcodec = "h264"
				acodec = "aac"
			} else if ext == "webm" {
				vcodec = "vp9"
				acodec = "opus"
			}
		}

		formats = append(formats, FormatInfo{
			URL:      u,
			FormatID: formatID,
			Ext:      ext,
			Acodec:   acodec,
			Vcodec:   vcodec,
			Height:   hPtr,
		})
	}

	return &VideoInfo{
		ID:          extractYouTubeID(url),
		Title:       strings.TrimSpace(resp.API.Title),
		Duration:    dur,
		Thumbnail:   thumb,
		WebpageURL:  url,
		Extractor:   "youtube",
		DownloadURL: downloadURL,
		Formats:     formats,
	}, nil
}

func (s *YTDownStrategy) Name() string {
	return "ytdown"
}

type FallbackDownloader struct {
	strategies []DownloaderStrategy
}

func NewFallbackDownloader() *FallbackDownloader {
	ytDown := &YTDownStrategy{client: scrapper.NewYTDownService()}
	ytDlp := &YtDlpStrategy{client: &ytDlpClient{executablePath: "python3"}}
	ytCustom := &YoutubeCustomStrategy{downloader: NewYoutubeDownloader()}
	ytGo := &YoutubeGoStrategy{client: youtube.Client{}}
	luxStrat := NewLuxStrategy()
	rumbleStrat := NewRumbleStrategy()
	chromedpStrat := NewChromedpStrategy()
	vimeoStrat := NewVimeoStrategy()

	return &FallbackDownloader{
		// Global default order: yt-dlp -> lux -> custom strategies -> chromedp (fallback)
		strategies: []DownloaderStrategy{ytDown, ytCustom, ytDlp, luxStrat, ytGo, rumbleStrat, vimeoStrat, chromedpStrat},
	}
}

func (f *FallbackDownloader) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 25*time.Second)
	defer cancel()

	return f.GetVideoInfoWithType(subCtx, url, "")
}

func (f *FallbackDownloader) GetVideoInfoWithType(ctx context.Context, url string, downloadType string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 25*time.Second)
	defer cancel()

	var lastErr error

	normalizedType := strings.ToLower(downloadType)
	isYoutube := normalizedType == "youtube" || normalizedType == "youtube-to-mp3" || strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
	isRumble := normalizedType == "rumble" || strings.Contains(url, "rumble.com")
	isVimeo := normalizedType == "vimeo" || strings.Contains(url, "vimeo.com")
	isTikTok := normalizedType == "tiktok" || strings.Contains(url, "tiktok.com")
	isTwitter := normalizedType == "twitter" || normalizedType == "x" || strings.Contains(url, "twitter.com") || strings.Contains(url, "x.com")
	isDailymotion := normalizedType == "dailymotion" || strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly")

	var strategies []DownloaderStrategy

	// Filter and prioritize strategies
	if isYoutube {
		// Requested order: ytdown -> yt-dlp -> youtube-custom -> chromedp
		var ytDownStrat, ytDlpStrat, ytCustomStrat, chromedpStrat DownloaderStrategy
		for _, strategy := range f.strategies {
			if strategy.Name() == "ytdown" {
				ytDownStrat = strategy
			} else if strategy.Name() == "yt-dlp" {
				ytDlpStrat = strategy
			} else if strategy.Name() == "youtube-custom" {
				ytCustomStrat = strategy
			} else if strategy.Name() == "chromedp" {
				chromedpStrat = strategy
			}
		}

		if ytDownStrat != nil {
			strategies = append(strategies, ytDownStrat)
		}
		if ytDlpStrat != nil {
			strategies = append(strategies, ytDlpStrat)
		}
		if ytCustomStrat != nil {
			strategies = append(strategies, ytCustomStrat)
		}
		if chromedpStrat != nil {
			strategies = append(strategies, chromedpStrat)
		}
	} else if isRumble {
		// For Rumble, use yt-dlp, lux, rumble-custom, and chromedp
		// Order in f.strategies (init in NewFallbackDownloader) is already: yt-dlp, lux, rumble-custom, chromedp
		// So we can just filter
		for _, strategy := range f.strategies {
			name := strategy.Name()
			if name == "yt-dlp" || name == "lux" || name == "rumble-custom" || name == "chromedp" {
				strategies = append(strategies, strategy)
			}
		}
	} else if isVimeo {
		var vimeoStrategy, ytDlpStrat, luxStrat, chromedpStrat DownloaderStrategy

		for _, strategy := range f.strategies {
			if strategy.Name() == "vimeo-custom" {
				vimeoStrategy = strategy
			} else if strategy.Name() == "yt-dlp" {
				ytDlpStrat = strategy
			} else if strategy.Name() == "lux" {
				luxStrat = strategy
			} else if strategy.Name() == "chromedp" {
				chromedpStrat = strategy
			}
		}

		if vimeoStrategy != nil {
			strategies = append(strategies, vimeoStrategy)
		}
		if ytDlpStrat != nil {
			strategies = append(strategies, ytDlpStrat)
		}
		if luxStrat != nil {
			strategies = append(strategies, luxStrat)
		}
		if chromedpStrat != nil {
			strategies = append(strategies, chromedpStrat)
		}
	} else if isTikTok {
		var ytDlpStrat, chromedpStrat, luxStrat DownloaderStrategy

		for _, strategy := range f.strategies {
			if strategy.Name() == "yt-dlp" {
				ytDlpStrat = strategy
			} else if strategy.Name() == "chromedp" {
				chromedpStrat = strategy
			} else if strategy.Name() == "lux" {
				luxStrat = strategy
			}
		}

		if ytDlpStrat != nil {
			strategies = append(strategies, ytDlpStrat)
		}
		if luxStrat != nil {
			strategies = append(strategies, luxStrat)
		}
		if chromedpStrat != nil {
			strategies = append(strategies, chromedpStrat)
		}
	} else if isTwitter || isDailymotion {
		var ytDlpStrat, luxStrat, chromedpStrat DownloaderStrategy

		for _, strategy := range f.strategies {
			name := strategy.Name()
			if name == "yt-dlp" {
				ytDlpStrat = strategy
			} else if name == "lux" {
				luxStrat = strategy
			} else if name == "chromedp" {
				chromedpStrat = strategy
			}
		}

		if isDailymotion {
			if chromedpStrat != nil {
				strategies = append(strategies, chromedpStrat)
			}
			if ytDlpStrat != nil {
				strategies = append(strategies, ytDlpStrat)
			}
			if luxStrat != nil {
				strategies = append(strategies, luxStrat)
			}
		} else {
			if ytDlpStrat != nil {
				strategies = append(strategies, ytDlpStrat)
			}
			if luxStrat != nil {
				strategies = append(strategies, luxStrat)
			}
			if chromedpStrat != nil {
				strategies = append(strategies, chromedpStrat)
			}
		}
	} else {
		// For non-YouTube URLs, exclude YouTube-specific strategies
		for _, strategy := range f.strategies {
			if strategy.Name() == "kkdai/youtube" || strategy.Name() == "youtube-custom" || strategy.Name() == "ytdown" {
				continue
			}
			strategies = append(strategies, strategy)
		}
	}

	for _, strategy := range strategies {
		log.Info().Str("strategy", strategy.Name()).Str("url", url).Msg("Attempting download with strategy")
		info, err := strategy.GetVideoInfo(subCtx, url)
		if err == nil {
			log.Info().Str("strategy", strategy.Name()).Msg("Download info success")
			return info, nil
		}
		log.Error().Err(err).Str("strategy", strategy.Name()).Msg("Strategy failed")
		lastErr = err
	}

	return nil, fmt.Errorf("all download strategies failed: %w", lastErr)
}

func (f *FallbackDownloader) DownloadVideo(ctx context.Context, url string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 30*time.Second)
	defer cancel()

	return f.GetVideoInfo(subCtx, url)
}

func (f *FallbackDownloader) DownloadToPath(ctx context.Context, url string, formatID string, outputPath string, cookies map[string]string) error {
	isYoutube := strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
	if isYoutube {
		if !strings.EqualFold(strings.TrimSpace(os.Getenv("YTDOWN_DISABLED")), "true") {
			yt := scrapper.NewYTDownService()
			if err := yt.DownloadToPath(ctx, url, outputPath, ""); err == nil {
				return nil
			}
		}

		client := &ytDlpClient{executablePath: "python3"}
		if err := client.DownloadToPath(ctx, url, formatID, outputPath, cookies); err == nil {
			return nil
		}

		if !strings.EqualFold(strings.TrimSpace(os.Getenv("YOUTUBE_CUSTOM_DISABLED")), "true") {
			yd := NewYoutubeDownloader()
			if _, err := yd.DownloadToPath(ctx, url, formatID, outputPath); err == nil {
				return nil
			}
		}

		if chromedp := NewChromedpStrategy(); chromedp != nil {
			if info, err := chromedp.GetVideoInfo(ctx, url); err == nil {
				direct := strings.TrimSpace(info.DownloadURL)
				if direct != "" && !strings.HasPrefix(strings.ToLower(direct), "blob:") {
					client := &ytDlpClient{executablePath: "python3"}
					return client.DownloadToPath(ctx, direct, "", outputPath, cookies)
				}
			}
		}
		return fmt.Errorf("all YouTube download strategies failed")
	}

	client := &ytDlpClient{executablePath: "python3"}
	return client.DownloadToPath(ctx, url, formatID, outputPath, cookies)
}

func pickBestYTDownItem(items []scrapper.YTDownMediaItem) *scrapper.YTDownMediaItem {
	var best *scrapper.YTDownMediaItem
	bestScore := -1
	for i := range items {
		it := &items[i]
		lt := strings.ToLower(it.Type)
		if strings.Contains(lt, "audio") {
			continue
		}
		ext := strings.ToLower(strings.TrimSpace(it.MediaExtension))
		score := 0
		if strings.Contains(lt, "video") {
			score += 10
		}
		if ext == "mp4" {
			score += 5
		}
		height := parseHeightFromString(it.MediaQuality)
		if height == 0 {
			if res, ok := it.MediaRes.(string); ok {
				height = parseHeightFromString(res)
			}
		}
		score += height
		if score > bestScore {
			best = it
			bestScore = score
		}
	}
	if best != nil {
		return best
	}
	if len(items) > 0 {
		return &items[0]
	}
	return nil
}

func parseHeightFromString(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`(\d{3,4})p`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

func parseDurationSeconds(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0
	}
	secPart := strings.TrimSpace(parts[len(parts)-1])
	minPart := strings.TrimSpace(parts[len(parts)-2])
	hourPart := "0"
	if len(parts) == 3 {
		hourPart = strings.TrimSpace(parts[0])
	}
	h, errH := strconv.Atoi(hourPart)
	m, errM := strconv.Atoi(minPart)
	sec, errS := strconv.Atoi(secPart)
	if errH != nil || errM != nil || errS != nil {
		return 0
	}
	if h < 0 || m < 0 || sec < 0 {
		return 0
	}
	return float64(h*3600 + m*60 + sec)
}

func extractYouTubeID(videoURL string) string {
	patterns := []string{
		`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})`,
		`youtu\.be\/([a-zA-Z0-9_-]{11})`,
		`youtube\.com\/v\/([a-zA-Z0-9_-]{11})`,
		`youtube\.com\/shorts\/([a-zA-Z0-9_-]{11})`,
		`[?&]v=([^&#]{11})`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		m := re.FindStringSubmatch(videoURL)
		if len(m) > 1 {
			return m[1]
		}
	}
	return ""
}
