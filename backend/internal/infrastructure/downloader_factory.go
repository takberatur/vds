package infrastructure

import (
	"context"
	"fmt"
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
	video, err := s.client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	formats := video.Formats.WithAudioChannels()
	var bestURL string
	if len(formats) > 0 {
		bestURL = formats[0].URL
	}

	return &VideoInfo{
		ID:          video.ID,
		Title:       video.Title,
		Duration:    video.Duration.Seconds(),
		Thumbnail:   video.Thumbnails[0].URL, // Pick first thumbnail
		WebpageURL:  "https://www.youtube.com/watch?v=" + video.ID,
		Extractor:   "youtube",
		DownloadURL: bestURL,
	}, nil
}

func (s *YoutubeGoStrategy) Name() string {
	return "kkdai/youtube"
}

type FallbackDownloader struct {
	strategies []DownloaderStrategy
}

func NewFallbackDownloader() *FallbackDownloader {
	ytDlp := &YtDlpStrategy{client: &ytDlpClient{executablePath: "yt-dlp"}}
	ytGo := &YoutubeGoStrategy{client: youtube.Client{}}
	luxStrat := NewLuxStrategy()
	rumbleStrat := NewRumbleStrategy()
	chromedpStrat := NewChromedpStrategy()

	return &FallbackDownloader{
		strategies: []DownloaderStrategy{ytDlp, ytGo, luxStrat, rumbleStrat, chromedpStrat},
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

	var strategies []DownloaderStrategy

	// Filter and prioritize strategies
	if isYoutube {
		var youtubeStrategy DownloaderStrategy
		var others []DownloaderStrategy

		for _, strategy := range f.strategies {
			if strategy.Name() == "kkdai/youtube" {
				youtubeStrategy = strategy
				continue
			}
			others = append(others, strategy)
		}

		if youtubeStrategy != nil {
			strategies = append([]DownloaderStrategy{youtubeStrategy}, others...)
		} else {
			strategies = others
		}
	} else if isRumble {
		// For Rumble, use yt-dlp, lux, rumble-custom, and chromedp
		for _, strategy := range f.strategies {
			name := strategy.Name()
			if name == "yt-dlp" || name == "lux" || name == "rumble-custom" || name == "chromedp" {
				strategies = append(strategies, strategy)
			}
		}
	} else {
		// For non-YouTube URLs, exclude kkdai/youtube strategy
		for _, strategy := range f.strategies {
			if strategy.Name() == "kkdai/youtube" {
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
