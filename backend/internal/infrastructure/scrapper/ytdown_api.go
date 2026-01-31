package scrapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type YTDownService struct {
	httpClient *http.Client
}

type YTDownVideoInfo struct {
	Title       string
	Duration    string
	Author      string
	VideoURL    string
	DownloadURL string
	Quality     string
	Size        string
	Format      string
	Thumbnail   string
}

type YTDownDownloadOption struct {
	Type       string
	Quality    string
	Format     string
	Size       string
	URL        string
	Duration   string
	Resolution string
	Extension  string
}
type YTDownResponse struct {
	API struct {
		Service       string `json:"service"`
		Status        string `json:"status"`
		Message       string `json:"message"`
		ID            string `json:"id"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		PreviewURL    string `json:"previewUrl"`
		ImagePreview  string `json:"imagePreviewUrl"`
		PermanentLink string `json:"permanentLink"`
		UserInfo      struct {
			Name           string `json:"name"`
			Username       string `json:"username"`
			UserID         string `json:"userId"`
			UserAvatar     string `json:"userAvatar"`
			UserBio        string `json:"userBio"`
			InternalURL    string `json:"internalUrl"`
			ExternalURL    string `json:"externalUrl"`
			AccountCountry string `json:"accountCountry"`
			DateJoined     string `json:"dateJoined"`
		} `json:"userInfo"`
		MediaStats struct {
			MediaCount     string      `json:"mediaCount"`
			FollowersCount string      `json:"followersCount"`
			LikesCount     interface{} `json:"likesCount"`
			CommentsCount  interface{} `json:"commentsCount"`
			ViewsCount     string      `json:"viewsCount"`
		} `json:"mediaStats"`
		MediaItems []YTDownMediaItem `json:"mediaItems"`
	} `json:"api"`
}

type YTDownMediaItem struct {
	Type            string      `json:"type"`
	Name            string      `json:"name"`
	MediaID         int64       `json:"mediaId"`
	MediaURL        string      `json:"mediaUrl"`
	MediaPreviewURL string      `json:"mediaPreviewUrl"`
	MediaThumbnail  string      `json:"mediaThumbnail"`
	MediaRes        interface{} `json:"mediaRes"`
	MediaQuality    string      `json:"mediaQuality"`
	MediaDuration   string      `json:"mediaDuration"`
	MediaExtension  string      `json:"mediaExtension"`
	MediaFileSize   string      `json:"mediaFileSize"`
	MediaTask       string      `json:"mediaTask"`
}

func NewYTDownService() *YTDownService {
	jar, _ := cookiejar.New(nil)
	return &YTDownService{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Jar:     jar,
			Transport: &http.Transport{
				DisableKeepAlives: false,
				MaxIdleConns:      10,
				IdleConnTimeout:   90 * time.Second,
			},
		},
	}
}

func (yt *YTDownService) apiURL() string {
	if v := strings.TrimSpace(os.Getenv("YTDOWN_API_URL")); v != "" {
		return v
	}
	return "https://ytdown.to/proxy.php"
}

func (yt *YTDownService) userAgent() string {
	if v := strings.TrimSpace(os.Getenv("YTDOWN_USER_AGENT")); v != "" {
		return v
	}
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
}

func (yt *YTDownService) Fetch(ctx context.Context, videoURL string) (*YTDownResponse, []byte, error) {
	videoID := yt.extractYtDownVideoID(videoURL)
	if videoID == "" {
		return nil, nil, fmt.Errorf("invalid YouTube URL")
	}

	formData := url.Values{}
	formData.Set("url", videoURL)

	req, err := http.NewRequestWithContext(ctx, "POST", yt.apiURL(), strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", yt.userAgent())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://ytdown.to")
	req.Header.Set("Referer", "https://ytdown.to/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := yt.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(body)
		preview = strings.TrimSpace(preview)
		if len(preview) > 400 {
			preview = preview[:400]
		}
		return nil, body, fmt.Errorf("HTTP status %d: %s; body=%q", resp.StatusCode, resp.Status, preview)
	}

	var ytResp YTDownResponse
	if err := json.Unmarshal(body, &ytResp); err != nil {
		preview := string(body)
		preview = strings.TrimSpace(preview)
		if len(preview) > 400 {
			preview = preview[:400]
		}
		log.Error().Str("body", preview).Msg("ytdown: failed to parse JSON")
		return nil, body, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if !strings.EqualFold(strings.TrimSpace(ytResp.API.Status), "OK") {
		msg := strings.TrimSpace(ytResp.API.Message)
		if msg == "" {
			msg = "unknown API error"
		}
		return nil, body, fmt.Errorf("API error: %s", msg)
	}

	return &ytResp, body, nil
}

func (yt *YTDownService) GetVideoInfo(videoURL string) (*YTDownVideoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, _, err := yt.Fetch(ctx, videoURL)
	if err != nil {
		return nil, err
	}

	best := yt.pickBestMediaItem(resp.API.MediaItems)
	info := &YTDownVideoInfo{
		VideoURL:  videoURL,
		Title:     resp.API.Title,
		Author:    resp.API.UserInfo.Name,
		Duration:  "",
		Thumbnail: firstNonEmpty(resp.API.ImagePreview, resp.API.PreviewURL),
	}
	if best != nil {
		info.DownloadURL = firstNonEmpty(best.MediaURL, best.MediaPreviewURL)
		info.Format = strings.ToUpper(strings.TrimSpace(best.MediaExtension))
		info.Quality = strings.TrimSpace(best.MediaQuality)
		info.Size = strings.TrimSpace(best.MediaFileSize)
		if d := strings.TrimSpace(best.MediaDuration); d != "" {
			info.Duration = d
		}
	}

	return info, nil
}

func (yt *YTDownService) GetDownloadURLs(videoURL string) ([]YTDownDownloadOption, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, _, err := yt.Fetch(ctx, videoURL)
	if err != nil {
		return nil, err
	}

	var options []YTDownDownloadOption
	for _, item := range resp.API.MediaItems {
		opt := YTDownDownloadOption{
			Type:      item.Type,
			Quality:   strings.TrimSpace(item.MediaQuality),
			Format:    strings.ToUpper(strings.TrimSpace(item.MediaExtension)),
			Size:      strings.TrimSpace(item.MediaFileSize),
			URL:       firstNonEmpty(item.MediaURL, item.MediaPreviewURL),
			Duration:  strings.TrimSpace(item.MediaDuration),
			Extension: strings.ToLower(strings.TrimSpace(item.MediaExtension)),
		}

		if res, ok := item.MediaRes.(string); ok {
			opt.Resolution = strings.TrimSpace(res)
		}

		options = append(options, opt)
	}

	return options, nil
}

func (yt *YTDownService) DownloadToPath(ctx context.Context, videoURL, outputPath, preferredQuality string) error {
	options, err := yt.GetDownloadURLs(videoURL)
	if err != nil {
		return err
	}

	var selectedOption *YTDownDownloadOption
	for i := range options {
		opt := &options[i]
		isVideo := strings.Contains(strings.ToLower(opt.Type), "video") || strings.Contains(strings.ToLower(opt.Type), "mp4")
		if preferredQuality != "" {
			if strings.EqualFold(opt.Quality, preferredQuality) || strings.EqualFold(opt.Resolution, preferredQuality) {
				selectedOption = opt
				break
			}
			continue
		}
		if isVideo {
			selectedOption = opt
			break
		}
	}

	if selectedOption == nil {
		if len(options) > 0 {
			selectedOption = &options[0]
		} else {
			return fmt.Errorf("no download options available")
		}
	}

	if err := yt.downloadFile(ctx, selectedOption.URL, outputPath); err != nil {
		return err
	}
	if fi, err := os.Stat(outputPath); err == nil {
		if fi.Size() < 1024 {
			return fmt.Errorf("downloaded file too small (%d bytes)", fi.Size())
		}
	}
	return nil
}

func (yt *YTDownService) downloadFile(ctx context.Context, downloadURL, outputPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", yt.userAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://ytdown.to/")
	req.Header.Set("Origin", "https://ytdown.to")
	req.Header.Set("Sec-Fetch-Dest", "video")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Range", "bytes=0-")

	resp, err := yt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, resp.Status)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	buf := make([]byte, 8192)
	n, readErr := io.ReadAtLeast(resp.Body, buf, 1)
	if readErr != nil {
		if errors.Is(readErr, io.EOF) {
			return fmt.Errorf("empty response body")
		}
		if errors.Is(readErr, io.ErrUnexpectedEOF) {
			n = maxInt(n, 0)
		} else {
			return fmt.Errorf("failed to read response: %v", readErr)
		}
	}
	buf = buf[:n]

	ct := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if looksLikeBlockedOrHTML(ct, buf) {
		preview := strings.TrimSpace(string(buf))
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return fmt.Errorf("unexpected content (content-type=%q) preview=%q", ct, preview)
	}
	if looksLikePlaylist(buf) {
		preview := strings.TrimSpace(string(buf))
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return fmt.Errorf("got playlist content preview=%q", preview)
	}
	if !looksLikeVideoHeader(buf) {
		preview := strings.TrimSpace(string(buf))
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return fmt.Errorf("unknown content header (content-type=%q) preview=%q", ct, preview)
	}

	if _, err := file.Write(buf); err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	return nil
}
func (yt *YTDownService) extractYtDownVideoID(videoURL string) string {
	patterns := []struct {
		regex string
		group int
	}{
		{`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})`, 1},
		{`youtu\.be\/([a-zA-Z0-9_-]{11})`, 1},
		{`youtube\.com\/v\/([a-zA-Z0-9_-]{11})`, 1},
		{`youtube\.com\/shorts\/([a-zA-Z0-9_-]{11})`, 1},
		{`[?&]v=([^&#]{11})`, 1},
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		matches := re.FindStringSubmatch(videoURL)
		if len(matches) > pattern.group {
			return matches[pattern.group]
		}
	}
	return ""
}

func (yt *YTDownService) pickBestMediaItem(items []YTDownMediaItem) *YTDownMediaItem {
	var best *YTDownMediaItem
	bestScore := -1
	for i := range items {
		it := &items[i]
		lt := strings.ToLower(it.Type)
		ext := strings.ToLower(strings.TrimSpace(it.MediaExtension))
		if strings.Contains(lt, "audio") {
			continue
		}
		if ext == "" {
			ext = "mp4"
		}
		score := 0
		if strings.Contains(lt, "video") {
			score += 10
		}
		if ext == "mp4" {
			score += 5
		}
		height := 0
		if res, ok := it.MediaRes.(string); ok {
			height = parseHeightFromString(res)
		}
		if height == 0 {
			height = parseHeightFromString(it.MediaQuality)
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
	n := 0
	for _, ch := range m[1] {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

func looksLikeBlockedOrHTML(contentType string, head []byte) bool {
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		return true
	}
	s := strings.ToLower(strings.TrimSpace(string(head)))
	if strings.HasPrefix(s, "<!doctype") || strings.HasPrefix(s, "<html") || strings.Contains(s, "access denied") {
		return true
	}
	return false
}

func looksLikePlaylist(head []byte) bool {
	s := strings.TrimSpace(string(head))
	return strings.HasPrefix(s, "#EXTM3U")
}

func looksLikeVideoHeader(head []byte) bool {
	if len(head) >= 12 {
		if string(head[4:8]) == "ftyp" {
			return true
		}
	}
	if len(head) >= 4 {
		if head[0] == 0x1A && head[1] == 0x45 && head[2] == 0xDF && head[3] == 0xA3 {
			return true
		}
	}
	return false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseNumericOrClockDurationSeconds(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if n, err := strconv.Atoi(s); err == nil {
		if n < 0 {
			return 0
		}
		return n
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
	return h*3600 + m*60 + sec
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
