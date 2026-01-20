package infrastructure

import "regexp"

type VideoDetectorInfo struct {
	URL         string
	Title       string
	Format      string
	Quality     string
	Size        string
	Duration    string
	Thumbnail   string
	IsHLS       bool
	IsStreaming bool
}
type VideoDetector struct {
	patterns map[string]*regexp.Regexp
}

func NewVideoDetector() *VideoDetector {
	return &VideoDetector{
		patterns: map[string]*regexp.Regexp{
			"videoTag":  regexp.MustCompile(`<video[^>]+src="([^"]+)"`),
			"sourceTag": regexp.MustCompile(`<source[^>]+src="([^"]+)"`),
			"iframe":    regexp.MustCompile(`<iframe[^>]+src="([^"]+youtu[^"]+)"`),
			"m3u8":      regexp.MustCompile(`(https?://[^"\s]+\.m3u8[^"\s]*)`),
			"mpd":       regexp.MustCompile(`(https?://[^"\s]+\.mpd[^"\s]*)`),
			"jsonVideo": regexp.MustCompile(`"videoUrl"\s*:\s*"([^"]+)"`),
		},
	}
}
