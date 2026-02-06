package logger

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/user/video-downloader-backend/pkg/telegram"
)

var once sync.Once

var telegramMu sync.RWMutex
var telegramNotifier *telegram.TelegramNotifier

func Init() {
	once.Do(func() {
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			panic("Failed to create logs directory: " + err.Error())
		}

		logFile, err := os.OpenFile(filepath.Join(logDir, "logs.json"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}

		zerolog.TimeFieldFormat = time.RFC3339
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

		tgWriter := initTelegramWriter()
		multiWriter := io.MultiWriter(os.Stdout, logFile, tgWriter)

		log.Logger = zerolog.New(multiWriter).With().Timestamp().Stack().Logger()
	})
}

func initTelegramWriter() io.Writer {
	enabled, _ := strconv.ParseBool(os.Getenv("TELEGRAM_NOTIFICATIONS"))
	if !enabled {
		return io.Discard
	}

	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	chatIDRaw := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))
	if token == "" || chatIDRaw == "" {
		return io.Discard
	}
	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		return io.Discard
	}

	n, err := telegram.NewTelegramNotifier(token, chatID)
	if err != nil {
		return io.Discard
	}
	telegramMu.Lock()
	telegramNotifier = n
	telegramMu.Unlock()
	return newTelegramLogWriter(true, n)
}
