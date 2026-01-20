package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Database struct {
	Pool *pgxpool.Pool
}

type MyQueryTracer struct{}

func (t *MyQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, "query_start_time", time.Now())
	ctx = context.WithValue(ctx, "query_sql", data.SQL)
	return ctx
}

func (t *MyQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value("query_start_time").(time.Time)
	if !ok {
		return
	}
	sql, _ := ctx.Value("query_sql").(string)
	duration := time.Since(start)

	if data.Err != nil {
		log.Error().Err(data.Err).Str("sql", sql).Str("duration", duration.String()).Msg("Database Query Failed")
	} else {
		if duration > 100*time.Millisecond {
			log.Warn().Str("sql", sql).Str("duration", duration.String()).Msg("Slow Database Query")
		} else {
			log.Debug().Str("sql", sql).Str("duration", duration.String()).Msg("Database Query Executed")
		}
	}
}

func NewPostgresClient(databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	config.ConnConfig.Tracer = &MyQueryTracer{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}
