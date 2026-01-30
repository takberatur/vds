package infrastructure

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/rs/zerolog/log"
)

type YoutubeDownloader struct {
	client youtube.Client
}

type VideoDetails struct {
	Title       string
	Author      string
	Duration    string
	Description string
	Formats     []YoutubeFormatInfo
}

type YoutubeFormatInfo struct {
	ItagNo       int
	Quality      string
	MimeType     string
	Bitrate      int
	AudioQuality string
	Size         int64
}

func NewYoutubeDownloader() *YoutubeDownloader {
	ua := strings.TrimSpace(os.Getenv("YOUTUBE_HTTP_USER_AGENT"))
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	}

	cookieHeader := ""
	if strings.EqualFold(strings.TrimSpace(os.Getenv("YOUTUBE_USE_COOKIES")), "true") {
		cookiePath := strings.TrimSpace(os.Getenv("YOUTUBE_COOKIES_FILE_PATH"))
		if cookiePath == "" {
			cookiePath = strings.TrimSpace(os.Getenv("COOKIES_FILE_PATH"))
		}
		if cookiePath == "" {
			cookiePath = "/app/cookies.txt"
		}
		cookieHeader = netscapeCookiesToHeader(cookiePath, []string{"youtube.com", "google.com", "accounts.google.com"})
	}

	httpClient := &http.Client{
		Timeout: 25 * time.Second,
		Transport: &headerRoundTripper{
			base:         http.DefaultTransport,
			userAgent:    ua,
			cookieHeader: cookieHeader,
		},
	}

	return &YoutubeDownloader{
		client: youtube.Client{HTTPClient: httpClient},
	}
}

type headerRoundTripper struct {
	base         http.RoundTripper
	userAgent    string
	cookieHeader string
}

func (rt *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if rt.userAgent != "" && r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", rt.userAgent)
	}
	if r.Header.Get("Accept-Language") == "" {
		r.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}
	if rt.cookieHeader != "" {
		r.Header.Set("Cookie", rt.cookieHeader)
	}
	base := rt.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(r)
}

func (yd *YoutubeDownloader) GetVideoDetails(videoURL string) (*VideoDetails, error) {
	if !yd.isValidYoutubeURL(videoURL) {
		return nil, fmt.Errorf("invalid youtube URL")
	}

	video, err := yd.client.GetVideo(videoURL)
	if err != nil {
		return nil, fmt.Errorf("error getting video: %v", err)
	}

	details := &VideoDetails{
		Title:       video.Title,
		Author:      video.Author,
		Duration:    video.Duration.String(),
		Description: video.Description,
		Formats:     make([]YoutubeFormatInfo, 0),
	}

	// Put available formats
	for _, format := range video.Formats {
		details.Formats = append(details.Formats, YoutubeFormatInfo{
			ItagNo:       format.ItagNo,
			Quality:      format.Quality,
			MimeType:     format.MimeType,
			Bitrate:      format.Bitrate,
			AudioQuality: format.AudioQuality,
		})
	}

	return details, nil
}
func (yd *YoutubeDownloader) DownloadVideo(videoURL string, itagNo int) (*youtube.Video, error) {
	video, err := yd.client.GetVideo(videoURL)
	if err != nil {
		return nil, fmt.Errorf("error getting video: %v", err)
	}

	// Get selected format
	var selectedFormat *youtube.Format
	if itagNo > 0 {
		// Use selected format
		for i := range video.Formats {
			if video.Formats[i].ItagNo == itagNo {
				selectedFormat = &video.Formats[i]
				break
			}
		}
	} else {
		// Pick best format (video+audio)
		selectedFormat = yd.getBestFormat(video)
	}

	if selectedFormat == nil {
		return nil, fmt.Errorf("no suitable format found")
	}

	log.Info().Str("format", "\n📊 Selected Format:\n").Msg(fmt.Sprintf("   Quality: %s\n", selectedFormat.Quality))
	log.Info().Str("format", "   MimeType: %s\n").Msg(selectedFormat.MimeType)
	if selectedFormat.Bitrate > 0 {
		log.Info().Str("format", "   Bitrate: %d\n").Msg(fmt.Sprintf("%d", selectedFormat.Bitrate))
	}

	// Get stream info
	stream, size, err := yd.client.GetStream(video, selectedFormat)
	if err != nil {
		return nil, fmt.Errorf("error getting stream: %v", err)
	}
	defer stream.Close()

	// bar := progressbar.DefaultBytes(
	// 	size,
	// 	"⏬ Downloading",
	// )

	// Download dengan progress
	// _, err = io.Copy(io.MultiWriter(os.Stdout, bar), stream)
	// if err != nil {
	// 	return nil, fmt.Errorf("error downloading: %v", err)
	// }

	log.Info().Str("format", "   Size: %d\n").Msg(fmt.Sprintf("%d", size))

	log.Info().Str("format", "\n\n✅ Download success!\n").Msg("")
	return video, nil
}

func (yd *YoutubeDownloader) DownloadToPath(ctx context.Context, videoURL string, formatID string, outputPath string) (*youtube.Video, error) {
	_ = ctx
	video, err := yd.client.GetVideo(videoURL)
	if err != nil {
		return nil, fmt.Errorf("error getting video: %v", err)
	}

	itagNo := 0
	if strings.TrimSpace(formatID) != "" {
		n, perr := strconv.Atoi(strings.TrimSpace(formatID))
		if perr == nil {
			itagNo = n
		}
	}

	var selectedFormat *youtube.Format
	if itagNo > 0 {
		for i := range video.Formats {
			if video.Formats[i].ItagNo == itagNo {
				selectedFormat = &video.Formats[i]
				break
			}
		}
	} else {
		selectedFormat = yd.getBestMuxedMP4Format(video)
		if selectedFormat == nil {
			selectedFormat = yd.getBestFormat(video)
		}
	}

	if selectedFormat == nil {
		return nil, fmt.Errorf("no suitable format found")
	}

	if selectedFormat.AudioChannels == 0 {
		return nil, fmt.Errorf("selected format has no audio; fallback required")
	}

	stream, _, err := yd.client.GetStream(video, selectedFormat)
	if err != nil {
		return nil, fmt.Errorf("error getting stream: %v", err)
	}
	defer stream.Close()

	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	_, copyErr := io.Copy(f, stream)
	closeErr := f.Close()
	if copyErr != nil {
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}

	return video, nil
}
func (yd *YoutubeDownloader) getBestFormat(video *youtube.Video) *youtube.Format {
	// Filter format for video+audio
	formats := video.Formats.WithAudioChannels()

	if len(formats) == 0 {
		// Fallback to video-only format if no audio+video format is available
		formats = video.Formats
	}

	if len(formats) == 0 {
		return nil
	}

	// Sort by quality (highest bitrate first)
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].Bitrate > formats[j].Bitrate
	})

	return &formats[0]
}

func (yd *YoutubeDownloader) getBestMuxedMP4Format(video *youtube.Video) *youtube.Format {
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return nil
	}

	mp4 := make([]youtube.Format, 0, len(formats))
	for _, f := range formats {
		if strings.Contains(strings.ToLower(f.MimeType), "video/mp4") {
			mp4 = append(mp4, f)
		}
	}
	if len(mp4) == 0 {
		return nil
	}
	sort.Slice(mp4, func(i, j int) bool {
		return mp4[i].Bitrate > mp4[j].Bitrate
	})
	return &mp4[0]
}

func netscapeCookiesToHeader(path string, domainSuffixes []string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(b), "\n")
	m := make(map[string]string, 64)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		domain := strings.TrimSpace(fields[0])
		domain = strings.TrimPrefix(domain, ".")
		ok := false
		for _, suf := range domainSuffixes {
			if strings.HasSuffix(domain, suf) {
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		name := strings.TrimSpace(fields[5])
		val := strings.TrimSpace(fields[6])
		if name == "" || val == "" {
			continue
		}
		if _, exists := m[name]; !exists {
			m[name] = val
		}
	}
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, "; ")
}
func (yd *YoutubeDownloader) ListFormats(videoURL string) ([]YoutubeFormatInfo, error) {
	video, err := yd.client.GetVideo(videoURL)
	if err != nil {
		return nil, err
	}

	var formats []YoutubeFormatInfo
	for _, format := range video.Formats {
		formats = append(formats, YoutubeFormatInfo{
			ItagNo:       format.ItagNo,
			Quality:      format.Quality,
			MimeType:     format.MimeType,
			Bitrate:      format.Bitrate,
			AudioQuality: format.AudioQuality,
		})
	}

	return formats, nil
}
func (d *YoutubeDownloader) DownloadWithRetry(url string, maxRetries int, itagNo int) (*youtube.Video, error) {
	var lastErr error
	video := &youtube.Video{}

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Info().Str("format", "\nRetry %d/%d...\n").Msg(fmt.Sprintf("%d", i+1))
			time.Sleep(2 * time.Second)
		}

		video, err := d.DownloadVideo(url, itagNo)
		if err == nil {
			return video, nil
		}

		lastErr = err
		log.Error().Str("format", "Error: %v\n").Msg(fmt.Sprintf("%v", err))
	}
	if lastErr != nil {
		return nil, fmt.Errorf("download failed after %d retries: %v", maxRetries, lastErr)
	}

	return video, nil
}
func (d *YoutubeDownloader) DownloadAllFormat(url string, maxRetries int) ([]*youtube.Video, error) {
	formats, err := d.ListFormats(url)
	if err != nil {
		return nil, fmt.Errorf("error listing formats: %v", err)
	}

	successCount := 0
	failedCount := 0
	totalStartTime := time.Now()

	videos := make([]*youtube.Video, 0)
	for _, format := range formats {
		video, err := d.DownloadWithRetry(url, maxRetries, format.ItagNo)
		if err != nil {
			failedCount++
			log.Error().Str("format", "Error downloading format %d: %v\n").Msg(fmt.Sprintf("%d", format.ItagNo))
			continue
		}
		successCount++
		videos = append(videos, video)
	}

	totalDuration := time.Since(totalStartTime)
	log.Info().Str("format", "\nDownload completed in %v\n").Msg(fmt.Sprintf("%v", totalDuration))
	log.Info().Str("format", "Successfully downloaded %d formats\n").Msg(fmt.Sprintf("%d", successCount))
	if failedCount > 0 {
		log.Error().Str("format", "Failed to download %d formats\n").Msg(fmt.Sprintf("%d", failedCount))
	}

	return videos, nil
}
func (yd *YoutubeDownloader) isValidYoutubeURL(url string) bool {
	if url == "" {
		return false
	}

	validPrefixes := []string{
		"https://www.youtube.com/watch?v=",
		"https://youtube.com/watch?v=",
		"http://www.youtube.com/watch?v=",
		"http://youtube.com/watch?v=",
		"https://youtu.be/",
		"http://youtu.be/",
		"https://www.youtube.com/shorts/",
		"https://youtube.com/shorts/",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}

	return false
}
