package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
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

type FallbackDownloader struct {
	strategies []DownloaderStrategy
}

func NewFallbackDownloader() *FallbackDownloader {
	ytDlp := &YtDlpStrategy{client: &ytDlpClient{executablePath: "python3"}}
	ytCustom := &YoutubeCustomStrategy{downloader: NewYoutubeDownloader()}
	ytGo := &YoutubeGoStrategy{client: youtube.Client{}}
	luxStrat := NewLuxStrategy()
	rumbleStrat := NewRumbleStrategy()
	chromedpStrat := NewChromedpStrategy()
	vimeoStrat := NewVimeoStrategy()

	return &FallbackDownloader{
		// Global default order: yt-dlp -> lux -> custom strategies -> chromedp (fallback)
		strategies: []DownloaderStrategy{ytCustom, ytDlp, luxStrat, ytGo, rumbleStrat, vimeoStrat, chromedpStrat},
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
		// Requested order: youtube-custom -> yt-dlp -> lux -> chromedp
		var ytCustomStrat, ytDlpStrat, luxStrat, chromedpStrat DownloaderStrategy
		for _, strategy := range f.strategies {
			if strategy.Name() == "youtube-custom" {
				ytCustomStrat = strategy
			} else if strategy.Name() == "yt-dlp" {
				ytDlpStrat = strategy
			} else if strategy.Name() == "lux" {
				luxStrat = strategy
			} else if strategy.Name() == "chromedp" {
				chromedpStrat = strategy
			}
		}

		if ytCustomStrat != nil {
			strategies = append(strategies, ytCustomStrat)
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
			if strategy.Name() == "kkdai/youtube" || strategy.Name() == "youtube-custom" {
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
	if isYoutube && !strings.EqualFold(strings.TrimSpace(os.Getenv("YOUTUBE_CUSTOM_DISABLED")), "true") {
		yd := NewYoutubeDownloader()
		if _, err := yd.DownloadToPath(ctx, url, formatID, outputPath); err == nil {
			return nil
		} else {
			log.Error().Err(err).Str("strategy", "youtube-custom").Msg("Strategy failed, falling back to yt-dlp")
		}
	}

	client := &ytDlpClient{executablePath: "python3"}
	return client.DownloadToPath(ctx, url, formatID, outputPath, cookies)
}
