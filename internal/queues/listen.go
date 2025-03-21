package queues

import (
	"context"
	"log/slog"

	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/events"
)

func Listen(q QueueManager, ctx context.Context) {
	eventChan := events.GetEventChannel()

	for event := range eventChan {
		switch event.EventType {
		case events.DownloadFailed:
			id := event.Payload.(events.DownloadFailedEvent).ID
			q.DownloadFailed(ctx, id)
		case events.DownloadProgressed:
			q.upsertChunks(ctx, event.Payload.(downloads.DownloadStatus))
		case events.DownloadCompleted:
			handleDownloadCompleted(ctx, q, event.Payload)
		default:
			slog.Error("Unknown Event type", "eventType", event.EventType)
		}
		events.GetUIEventChannel() <- events.Event{
			EventType: event.EventType,
			Payload:   event.Payload,
		}
	}
}

func handleDownloadCompleted(ctx context.Context, q QueueManager, payload interface{}) {
	id := payload.(downloads.DownloadStatus).ID
	q.setDownloadState(ctx, id, string(downloads.StateCompleted))
	slog.Info("Download marked as completed", "downloadID", id)
	q.startNextDownloadIfPossibleByDownloadID(ctx, id)
}
