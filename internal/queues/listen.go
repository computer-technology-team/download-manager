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
			q.UpsertChunks(ctx, event.Payload.(downloads.DownloadStatus))
		case events.DownloadCompleted:
			id := event.Payload.(downloads.DownloadStatus).ID
			q.DownloadCompleted(ctx, id)
		default:
			slog.Error("Unknown Event type", "eventType", event.EventType)
		}
		events.GetUIEventChannel() <- events.Event{
			EventType: event.EventType,
			Payload:   event.Payload,
		}
	}
}
