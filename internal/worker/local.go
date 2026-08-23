package worker

import (
	"context"
	"log/slog"
)

type LocalHandlers struct{ Logger *slog.Logger }

func (h LocalHandlers) HandleJob(ctx context.Context, kind, aggregateID string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.Logger.Info("job handled", "kind", kind, "aggregate_id", aggregateID, "payload_bytes", len(payload))
	return nil
}
func (h LocalHandlers) Publish(ctx context.Context, topic, aggregateID string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.Logger.Info("event published", "topic", topic, "aggregate_id", aggregateID, "payload_bytes", len(payload))
	return nil
}
