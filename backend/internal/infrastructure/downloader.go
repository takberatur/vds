package infrastructure

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
)

type VideoInfo struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Duration    *float64          `json:"duration"` // seconds
	Thumbnail   string            `json:"thumbnail"`
	WebpageURL  string            `json:"webpage_url"`
	Extractor   string            `json:"extractor"` // youtube, tiktok, etc.
	Filename    string            `json:"filename,omitempty"`
	Filesize    *int64            `json:"filesize,omitempty"`
	DownloadURL string            `json:"url,omitempty"` // Direct link if available
	UserAgent   string            `json:"user_agent,omitempty"`
	Cookies     map[string]string `json:"cookies,omitempty"` // Cookies required for download
	Formats     []FormatInfo      `json:"formats,omitempty"`
}

type FormatInfo struct {
	URL      string   `json:"url"`
	Filesize *int64   `json:"filesize,omitempty"`
	FormatID string   `json:"format_id,omitempty"`
	Acodec   string   `json:"acodec,omitempty"`
	Vcodec   string   `json:"vcodec,omitempty"`
	Ext      string   `json:"ext,omitempty"`
	Height   *int     `json:"height,omitempty"`
	Width    *int     `json:"width,omitempty"`
	Tbr      *float64 `json:"tbr,omitempty"`
}

type DownloaderClient interface {
	GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error)
	DownloadVideo(ctx context.Context, url string) (*VideoInfo, error)
	DownloadToPath(ctx context.Context, url string, formatID string, outputPath string, cookies map[string]string) error
}

type ytDlpClient struct {
	executablePath string
}

func shouldUseCookiesFile(path string) bool {
	if strings.EqualFold(sanitizeEnvString(os.Getenv("DISABLE_COOKIES_FILE")), "true") {
		return false
	}
	return IsValidNetscapeCookiesFile(path)
}

func NewDownloaderClient() DownloaderClient {
	return &ytDlpClient{
		executablePath: "python3",
	}
}

func (c *ytDlpClient) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 45*time.Second) // Increased timeout
	defer cancel()

	var userAgent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

	// Only use mobile UA for Instagram if needed, but manual tests suggest Desktop might be better or equal if cookies are issue.
	// For TikTok, manual test with default UA worked, so we revert the forced mobile UA.
	if strings.Contains(url, "instagram.com") {
		// Instagram often requires login, Mobile UA sometimes triggers a lighter page but can also trigger different checks.
		// Let's stick to Desktop for now as per manual test insights, or try a very standard one.
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	}

	args := []string{
		"-m", "yt_dlp", // Run as python module to ensure curl-cffi is found
		"--js-runtimes", defaultJSRuntime(),
		"--dump-json",
		"--no-playlist",
		"--no-check-certificate",
	}

	if proxyURL := sanitizeEnvString(os.Getenv("OUTBOUND_PROXY_URL")); proxyURL != "" && shouldUseProxyForURL(url) {
		args = append(args, "--proxy", proxyURL)
	}

	addImpersonate := false
	if strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly") {
		addImpersonate = true
	}
	if strings.Contains(url, "tiktok.com") {
		addImpersonate = true
	}
	imp := sanitizeEnvString(os.Getenv("YTDLP_IMPERSONATE"))
	if imp == "" {
		imp = "chrome"
	}

	if !strings.Contains(url, "tiktok.com") && !strings.Contains(url, "youtube.com") && !strings.Contains(url, "dailymotion.com") && !strings.Contains(url, "dai.ly") {
		args = append(args, "--user-agent", userAgent)
	}

	// NOTE: Removed --impersonate because it causes issues with missing dependencies in the current Docker environment.
	// We will rely on standard yt-dlp behavior or cookie usage.
	// if strings.Contains(url, "tiktok.com") {
	// 	args = append(args, "--impersonate", "chrome110")
	// }

	if strings.Contains(url, "vimeo.com") {
		args = append(args, "--referer", "https://vimeo.com/")
	}

	// Check if cookies.txt exists and use it
	cookiePath := os.Getenv("COOKIES_FILE_PATH")
	if shouldUseCookiesFile(cookiePath) {
		args = append(args, "--cookies", cookiePath)
	} else if _, err := os.Stat("cookies.txt"); err == nil {
		if shouldUseCookiesFile("cookies.txt") {
			args = append(args, "--cookies", "cookies.txt")
		}
	}

	// Add verbose logging for debugging
	args = append(args, "--verbose")

	if strings.Contains(url, "rumble.com") {
		args = append(args, "--referer", "https://rumble.com/")
		args = append(args, "--add-header", "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		args = append(args, "--add-header", "Accept-Language: en-US,en;q=0.9")
	}

	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		// Switch to tv client which is often less restricted
		args = append(args, "--extractor-args", "youtube:player_client=tv")
	}

	if strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly") {
		args = append(args, "--referer", "https://www.dailymotion.com/")
	}

	args = append(args, url)

	argsWithImp := args
	if addImpersonate {
		argsWithImp = make([]string, 0, len(args)+2)
		argsWithImp = append(argsWithImp, args[:len(args)-1]...)
		argsWithImp = append(argsWithImp, "--impersonate", imp, args[len(args)-1])
	}

	tryRun := func(a []string) ([]byte, error) {
		cmd := exec.CommandContext(subCtx, c.executablePath, a...)
		return cmd.Output()
	}

	output, err := tryRun(argsWithImp)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if addImpersonate && strings.Contains(stderr, "Impersonate target") {
				output2, err2 := tryRun(args)
				if err2 == nil {
					output = output2
					goto Parse
				}
				if exitErr2, ok2 := err2.(*exec.ExitError); ok2 {
					log.Error().Str("url", url).Str("stderr", string(exitErr2.Stderr)).Err(err2).Msg("yt-dlp failed")
				} else {
					log.Error().Str("url", url).Err(err2).Msg("yt-dlp failed")
				}
				return nil, fmt.Errorf("failed to fetch video info: %w", err2)
			}
			if (strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")) && (strings.Contains(stderr, "Sign in to confirm") || strings.Contains(stderr, "not a bot")) {
				legacyArgs := []string{
					"-m", "yt_dlp",
					"--js-runtimes", defaultJSRuntime(),
					"--dump-json",
					"--no-playlist",
					"--no-check-certificate",
					"-f", "18",
				}
				if proxyURL := sanitizeEnvString(os.Getenv("OUTBOUND_PROXY_URL")); proxyURL != "" && shouldUseProxyForURL(url) {
					legacyArgs = append(legacyArgs, "--proxy", proxyURL)
				}
				legacyArgs = append(legacyArgs, "--verbose", url)
				if out2, err2 := tryRun(legacyArgs); err2 == nil {
					output = out2
					goto Parse
				}
			}
			log.Error().Str("url", url).Str("stderr", stderr).Err(err).Msg("yt-dlp failed")
		} else {
			log.Error().Str("url", url).Err(err).Msg("yt-dlp failed")
		}
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

Parse:
	var info VideoInfo

	// Use an alias to avoid recursion and skip the original Cookies field
	type Alias VideoInfo
	aux := &struct {
		Cookies interface{} `json:"cookies,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&info),
	}

	if err := json.Unmarshal(output, &aux); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	// Handle flexible Cookies field
	if aux.Cookies != nil {
		switch v := aux.Cookies.(type) {
		case map[string]interface{}:
			info.Cookies = make(map[string]string)
			for k, val := range v {
				if strVal, ok := val.(string); ok {
					info.Cookies[k] = strVal
				}
			}
		case string:
			// If it's a string, we can't easily map it to map[string]string without parsing.
			// Log it and ignore to avoid crash.
			log.Warn().Str("url", url).Msg("yt-dlp returned cookies as string, ignoring")
		}
	}

	if info.DownloadURL == "" && len(info.Formats) > 0 {
		if best := pickBestFormat(info.Formats); best != nil {
			info.DownloadURL = best.URL
			if info.Filesize == nil && best.Filesize != nil {
				info.Filesize = best.Filesize
			}
		}
	}

	return &info, nil
}

func isHLSFormat(f FormatInfo) bool {
	return strings.Contains(f.URL, ".m3u8")
}

func pickBestFormat(formats []FormatInfo) *FormatInfo {
	if len(formats) == 0 {
		return nil
	}

	var bestNonHLS *FormatInfo
	for i := range formats {
		f := &formats[i]
		if f.URL == "" || isHLSFormat(*f) {
			continue
		}
		if bestNonHLS == nil {
			bestNonHLS = f
			continue
		}

		fHeight := 0
		if f.Height != nil {
			fHeight = *f.Height
		}
		bestHeight := 0
		if bestNonHLS.Height != nil {
			bestHeight = *bestNonHLS.Height
		}

		if fHeight > bestHeight {
			bestNonHLS = f
			continue
		}

		fTbr := 0.0
		if f.Tbr != nil {
			fTbr = *f.Tbr
		}
		bestTbr := 0.0
		if bestNonHLS.Tbr != nil {
			bestTbr = *bestNonHLS.Tbr
		}

		if fHeight == bestHeight && fTbr > bestTbr {
			bestNonHLS = f
			continue
		}

		fSize := int64(0)
		if f.Filesize != nil {
			fSize = *f.Filesize
		}
		bestSize := int64(0)
		if bestNonHLS.Filesize != nil {
			bestSize = *bestNonHLS.Filesize
		}

		if fHeight == bestHeight && fTbr == bestTbr && fSize > bestSize {
			bestNonHLS = f
			continue
		}
	}

	if bestNonHLS != nil {
		return bestNonHLS
	}

	var best *FormatInfo
	for i := range formats {
		f := &formats[i]
		if f.URL == "" {
			continue
		}
		if best == nil {
			best = f
			continue
		}

		fHeight := 0
		if f.Height != nil {
			fHeight = *f.Height
		}
		bestHeight := 0
		if best.Height != nil {
			bestHeight = *best.Height
		}

		if fHeight > bestHeight {
			best = f
			continue
		}

		fTbr := 0.0
		if f.Tbr != nil {
			fTbr = *f.Tbr
		}
		bestTbr := 0.0
		if best.Tbr != nil {
			bestTbr = *best.Tbr
		}

		if fHeight == bestHeight && fTbr > bestTbr {
			best = f
			continue
		}

		fSize := int64(0)
		if f.Filesize != nil {
			fSize = *f.Filesize
		}
		bestSize := int64(0)
		if best.Filesize != nil {
			bestSize = *best.Filesize
		}

		if fHeight == bestHeight && fTbr == bestTbr && fSize > bestSize {
			best = f
			continue
		}
	}
	return best
}

func (c *ytDlpClient) DownloadVideo(ctx context.Context, url string) (*VideoInfo, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 25*time.Second)
	defer cancel()

	info, err := c.GetVideoInfo(subCtx, url)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (c *ytDlpClient) DownloadToPath(ctx context.Context, url string, formatID string, outputPath string, cookies map[string]string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 20*time.Minute) // Increased timeout for large downloads
	defer cancel()

	addImpersonate := false
	if strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly") {
		addImpersonate = true
	}
	if strings.Contains(url, "tiktok.com") {
		addImpersonate = true
	}
	imp := sanitizeEnvString(os.Getenv("YTDLP_IMPERSONATE"))
	if imp == "" {
		imp = "chrome"
	}

	args := []string{
		"-m", "yt_dlp", // Run as python module
		"--js-runtimes", defaultJSRuntime(),
		"--no-playlist",
		"--no-check-certificate",
		"--force-overwrites",
		"--no-part",
		"-o", outputPath,
	}

	if proxyURL := sanitizeEnvString(os.Getenv("OUTBOUND_PROXY_URL")); proxyURL != "" && shouldUseProxyForURL(url) {
		args = append(args, "--proxy", proxyURL)
	}

	if !strings.Contains(url, "tiktok.com") && !strings.Contains(url, "youtube.com") && !strings.Contains(url, "dailymotion.com") && !strings.Contains(url, "dai.ly") {
		args = append(args, "--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	}

	// if strings.Contains(url, "tiktok.com") {
	// 	args = append(args, "--impersonate", "chrome110")
	// }

	if strings.Contains(url, "vimeo.com") {
		args = append(args, "--referer", "https://vimeo.com/")
	}

	// Check if cookies.txt exists and use it
	cookiePath := "/app/cookies.txt"
	isYouTube := strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
	if shouldUseCookiesFile(cookiePath) && (!isYouTube || strings.EqualFold(sanitizeEnvString(os.Getenv("YOUTUBE_USE_COOKIES")), "true")) {
		args = append(args, "--cookies", cookiePath)
	} else if _, err := os.Stat("cookies.txt"); err == nil {
		if shouldUseCookiesFile("cookies.txt") && (!isYouTube || strings.EqualFold(sanitizeEnvString(os.Getenv("YOUTUBE_USE_COOKIES")), "true")) {
			args = append(args, "--cookies", "cookies.txt")
		}
	}

	if len(cookies) > 0 {
		var cookieParts []string
		for k, v := range cookies {
			cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", k, v))
		}
		cookieStr := strings.Join(cookieParts, "; ")
		args = append(args, "--add-header", fmt.Sprintf("Cookie: %s", cookieStr))
	}

	if formatID != "" {
		args = append(args, "-f", formatID)
	}

	if strings.Contains(url, "rumble.com") {
		args = append(args, "--referer", "https://rumble.com/")
	}

	if isYouTube {
		if client := sanitizeEnvString(os.Getenv("YOUTUBE_PLAYER_CLIENT")); client != "" {
			args = append(args, "--extractor-args", "youtube:player_client="+client)
		}
	}

	if strings.Contains(url, "dailymotion.com") || strings.Contains(url, "dai.ly") {
		args = append(args, "--referer", "https://www.dailymotion.com/")
	}

	if strings.Contains(url, "snapchat.com") {
		args = append(args, "--allow-untrusted-extensions")
		if formatID == "" {
			args = append(args, "-f", "best[ext=mp4]/best")
		}
		args = append(args, "--merge-output-format", "mp4")
	}

	args = append(args, url)
	argsWithImp := args
	if addImpersonate {
		argsWithImp = make([]string, 0, len(args)+2)
		argsWithImp = append(argsWithImp, args[:len(args)-1]...)
		argsWithImp = append(argsWithImp, "--impersonate", imp, args[len(args)-1])
	}

	run := func(a []string) (string, error) {
		cmd := exec.CommandContext(subCtx, c.executablePath, a...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stderr.String(), err
	}

	if addImpersonate {
		if stderr, err := run(argsWithImp); err == nil {
			_ = stderr
			return nil
		} else if strings.Contains(stderr, "Impersonate target") {
			if stderr2, err2 := run(args); err2 != nil {
				return fmt.Errorf("yt-dlp download failed: %w, stderr: %s", err2, stderr2)
			}
			return nil
		} else if (strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")) && (strings.Contains(stderr, "Sign in to confirm") || strings.Contains(stderr, "not a bot")) {
			legacyArgs := []string{
				"-m", "yt_dlp",
				"--js-runtimes", defaultJSRuntime(),
				"--no-playlist",
				"--no-check-certificate",
				"--force-overwrites",
				"--no-part",
				"-o", outputPath,
				"-f", "18",
				url,
			}
			if proxyURL := sanitizeEnvString(os.Getenv("OUTBOUND_PROXY_URL")); proxyURL != "" && shouldUseProxyForURL(url) {
				legacyArgs = append(legacyArgs[:len(legacyArgs)-1], "--proxy", proxyURL, legacyArgs[len(legacyArgs)-1])
			}
			if stderr2, err2 := run(legacyArgs); err2 != nil {
				return fmt.Errorf("yt-dlp download failed: %w, stderr: %s", err2, stderr2)
			}
			return nil
		} else {
			return fmt.Errorf("yt-dlp download failed: %w, stderr: %s", err, stderr)
		}
	}

	if stderr, err := run(args); err != nil {
		return fmt.Errorf("yt-dlp download failed: %w, stderr: %s", err, stderr)
	}

	return nil
}

func sanitizeEnvString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "`")
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	return strings.TrimSpace(s)
}

func IsValidNetscapeCookiesFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	seenHeader := false
	seenData := false
	checked := 0
	for sc.Scan() && checked < 200 {
		checked++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "# Netscape HTTP Cookie File") {
			seenHeader = true
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		if !seenHeader {
			return false
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[") || strings.HasPrefix(lower, "<!doctype") || strings.HasPrefix(lower, "<html") {
			return false
		}

		if strings.Count(line, "\t") >= 6 {
			seenData = true
			continue
		}

		return false
	}

	return seenHeader && seenData
}

func shouldUseProxyForURL(rawURL string) bool {
	if strings.EqualFold(sanitizeEnvString(os.Getenv("PROXY_FOR_ALL")), "true") {
		return true
	}

	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}

	exclude := sanitizeEnvString(os.Getenv("PROXY_EXCLUDE_HOSTS"))
	if exclude != "" {
		for _, p := range strings.Split(exclude, ",") {
			p = strings.ToLower(strings.TrimSpace(p))
			if p == "" {
				continue
			}
			if host == p || strings.HasSuffix(host, "."+p) {
				return false
			}
		}
	}

	include := sanitizeEnvString(os.Getenv("PROXY_INCLUDE_HOSTS"))
	var patterns []string
	if include != "" {
		patterns = strings.Split(include, ",")
	} else {
		patterns = []string{
			"tiktok.com",
			"dailymotion.com",
			"dai.ly",
			"rumble.com",
			"snackvideo.com",
			"pinterest.com",
			"pin.it",
			"twitch.tv",
			"snapchat.com",
			"linkedin.com",
		}
	}

	for _, p := range patterns {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if host == p || strings.HasSuffix(host, "."+p) {
			return true
		}
	}

	return false
}

func defaultJSRuntime() string {
	if v := sanitizeEnvString(os.Getenv("YTDLP_JS_RUNTIME")); v != "" {
		return v
	}
	if _, err := os.Stat("/usr/local/bin/deno"); err == nil {
		return "deno:/usr/local/bin/deno"
	}
	if _, err := os.Stat("/usr/bin/deno"); err == nil {
		return "deno:/usr/bin/deno"
	}
	if _, err := os.Stat("/usr/bin/node"); err == nil {
		return "node:/usr/bin/node"
	}
	if _, err := os.Stat("/usr/bin/nodejs"); err == nil {
		return "node:/usr/bin/nodejs"
	}
	return "node"
}
