package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func NotifyTelegram(format string, args ...any) {
	enabled, _ := strconv.ParseBool(os.Getenv("TELEGRAM_NOTIFICATIONS"))
	if !enabled {
		return
	}
	text := strings.TrimSpace(fmt.Sprintf(format, args...))
	if text == "" {
		return
	}

	telegramMu.RLock()
	n := telegramNotifier
	telegramMu.RUnlock()
	if n == nil {
		return
	}

	go func() {
		_ = n.SendText(text)
	}()
}

