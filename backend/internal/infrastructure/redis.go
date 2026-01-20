package infrastructure

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisHook struct{}

func (h *RedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(start)

		if err != nil && err != redis.Nil {
			log.Error().Err(err).Str("command", cmd.Name()).Str("args", fmt.Sprintf("%v", cmd.Args())).Str("duration", duration.String()).Msg("Redis Command Failed")
		} else {
			log.Debug().Str("command", cmd.Name()).Str("args", fmt.Sprintf("%v", cmd.Args())).Str("duration", duration.String()).Msg("Redis Command Executed")
		}

		return err
	}
}

func (h *RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		duration := time.Since(start)

		if err != nil {
			log.Error().Err(err).Int("pipeline_size", len(cmds)).Str("duration", duration.String()).Msg("Redis Pipeline Failed")
		} else {
			log.Debug().Int("pipeline_size", len(cmds)).Str("duration", duration.String()).Msg("Redis Pipeline Executed")
		}

		return err
	}
}

func NewRedisClient(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Add Logger Hook
	rdb.AddHook(&RedisHook{})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}
