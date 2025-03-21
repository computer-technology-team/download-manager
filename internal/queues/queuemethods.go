package queues

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

func (q *queueManager) CreateQueue(ctx context.Context, createQueueParams state.CreateQueueParams) error {
	queue, err := q.queries.CreateQueue(ctx, createQueueParams)
	if err != nil {
		slog.Error("failed to create queue", "params", createQueueParams, "error", err)
		return fmt.Errorf("failed to create queue: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueCreated,
		Payload:   queue,
	}

	slog.Info("queue created successfully", "queueID", queue.ID)
	return nil
}

func (q *queueManager) DeleteQueue(ctx context.Context, id int64) error {
	err := q.queries.DeleteQueue(ctx, id)
	if err != nil {
		slog.Error("failed to delete queue", "queueID", id, "error", err)
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueDeleted,
		Payload:   id,
	}

	slog.Info("queue deleted successfully", "queueID", id)
	return nil
}

func (q *queueManager) EditQueue(ctx context.Context, arg state.UpdateQueueParams) error {
	queue, err := q.queries.UpdateQueue(ctx, arg)
	if err != nil {
		slog.Error("failed to update queue", "params", arg, "error", err)
		return fmt.Errorf("failed to update queue: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueEdited,
		Payload:   queue,
	}

	slog.Info("queue updated successfully", "queueID", queue.ID)
	return nil
}

func (q *queueManager) ListQueue(ctx context.Context) ([]state.Queue, error) {
	queues, err := q.queries.ListQueues(ctx)
	if err != nil {
		slog.Error("failed to list queues", "error", err)
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	slog.Info("listed queues", "count", len(queues))
	return queues, nil
}
