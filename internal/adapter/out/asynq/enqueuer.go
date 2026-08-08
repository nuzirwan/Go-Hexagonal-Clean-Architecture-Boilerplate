package asynq

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/nuzirwan/go-boilerplate/internal/domain/event"
)

type TaskEnqueuer struct {
	client *asynq.Client
}

func NewTaskEnqueuer(redisAddr string) *TaskEnqueuer {
	return &TaskEnqueuer{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

func (e *TaskEnqueuer) Close() error {
	return e.client.Close()
}

// Publish implements port.EventPublisher.
func (e *TaskEnqueuer) Publish(ctx context.Context, events ...event.Event) error {
	for _, evt := range events {
		payload, err := sonic.Marshal(evt)
		if err != nil {
			return err
		}
		task := asynq.NewTask(evt.Type(), payload)
		_, err = e.client.EnqueueContext(ctx, task,
			asynq.MaxRetry(3),
			asynq.Timeout(30*time.Second),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
