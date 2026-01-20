package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/internal/model"
)

const (
	TypeVideoDownload    = "video:download"
	DownloadEventChannel = "download:events"
)

type TaskClient interface {
	EnqueueVideoDownload(task *model.DownloadTask) error
}

type asynqTaskClient struct {
	client *asynq.Client
}

func NewTaskClient(redisAddr string, redisPassword string) TaskClient {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		DB:       1,
		Password: redisPassword,
	})
	return &asynqTaskClient{client: client}
}

func NewTaskServer(redisAddr string, redisPassword string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			DB:       1,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
}

func (c *asynqTaskClient) EnqueueVideoDownload(task *model.DownloadTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	t := asynq.NewTask(TypeVideoDownload, payload)
	_, err = c.client.Enqueue(t)
	return err
}

type RedisClient interface {
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
}
