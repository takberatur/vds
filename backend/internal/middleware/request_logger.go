package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		
		msg := "Request processed"
		
		var event *zerolog.Event
		if err != nil || status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		} else {
			event = log.Info()
		}

		event.Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Str("ip", c.IP()).
			Str("latency", duration.String()).
			Str("user_agent", c.Get("User-Agent"))

		if err != nil {
			event.Err(err)
		}

		event.Msg(msg)

		return err
	}
}
