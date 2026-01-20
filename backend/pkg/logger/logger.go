package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var once sync.Once

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

		multiWriter := io.MultiWriter(os.Stdout, logFile)

		log.Logger = zerolog.New(multiWriter).With().Timestamp().Stack().Logger()
	})
}
