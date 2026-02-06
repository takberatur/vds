package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/user/video-downloader-backend/pkg/telegram"
)

type telegramLogWriter struct {
	enabled bool
	notify  func(string)

	mu       sync.Mutex
	lastSent map[string]time.Time
	minGap   time.Duration
}

func newTelegramLogWriter(enabled bool, notifier *telegram.TelegramNotifier) *telegramLogWriter {
	w := &telegramLogWriter{
		enabled: enabled && notifier != nil,
		lastSent: make(map[string]time.Time),
		minGap: 30 * time.Second,
	}
	if w.enabled {
		w.notify = func(s string) {
			_ = notifier.SendText(s)
		}
	}
	return w
}

func (w *telegramLogWriter) Write(p []byte) (int, error) {
	if !w.enabled {
		return len(p), nil
	}

	line := strings.TrimSpace(string(p))
	if line == "" {
		return len(p), nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return len(p), nil
	}

	level, _ := payload["level"].(string)
	if level != "error" && level != "fatal" && level != "panic" {
		return len(p), nil
	}

	msg, _ := payload["message"].(string)
	errStr, _ := payload["error"].(string)

	key := level + "|" + msg + "|" + errStr
	now := time.Now()
	w.mu.Lock()
	last := w.lastSent[key]
	if !last.IsZero() && now.Sub(last) < w.minGap {
		w.mu.Unlock()
		return len(p), nil
	}
	w.lastSent[key] = now
	w.mu.Unlock()

	service := os.Getenv("APP_ENV")
	if service == "" {
		service = "backend"
	}

	text := fmt.Sprintf(
		"[%s] %s\nmsg: %s\nerr: %s\nhost: %s\ngo: %s",
		strings.ToUpper(level),
		service,
		strings.TrimSpace(msg),
		strings.TrimSpace(errStr),
		safeHost(),
		runtime.Version(),
	)

	go w.notify(text)
	return len(p), nil
}

func safeHost() string {
	h, _ := os.Hostname()
	if h == "" {
		return "-"
	}
	return h
}

