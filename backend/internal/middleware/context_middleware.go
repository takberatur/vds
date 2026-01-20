package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	ctxKey     = "ctx"
	timeoutKey = "route_timeout"
	minTimeout = 5 * time.Second
	maxTimeout = 6 * time.Hour
)

func ContextMiddleware(defaultTimeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		parentCtx := c.UserContext()
		if parentCtx == nil {
			parentCtx = context.Background()
		}

		routeTimeout := defaultTimeout
		if t, ok := c.Locals(timeoutKey).(time.Duration); ok {
			routeTimeout = t
		}

		if routeTimeout <= 0 {
			routeTimeout = defaultTimeout
		}
		if routeTimeout < minTimeout {
			routeTimeout = minTimeout
		}
		if routeTimeout > maxTimeout {
			routeTimeout = maxTimeout
		}

		ctx, cancel := context.WithTimeout(parentCtx, routeTimeout)
		defer cancel()

		c.Locals("ctx", ctx)

		if deadline, ok := ctx.Deadline(); ok {
			log.Info().Str("path", c.Path()).Dur("timeout", routeTimeout).Time("deadline", deadline).Msg("Context initialized")
		}

		return c.Next()
	}
}
func SetTimeoutContext(duration time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(timeoutKey, duration)
		log.Info().Str("path", c.Path()).Dur("duration", duration).Msg("The context time has been set plus for the route")
		return c.Next()
	}
}
func FromContext(c *fiber.Ctx) context.Context {
	if ctx, ok := c.Locals(ctxKey).(context.Context); ok && ctx != nil {
		return ctx
	}
	return context.Background()
}
func HandlerContext(c *fiber.Ctx) context.Context {
	ctx, ok := c.Locals(ctxKey).(context.Context)
	if !ok || ctx == nil {
		return context.Background()
	}
	log.Info().Str("path", c.Path()).Msg("The context time has been set plus for the route")

	return ctx
}
