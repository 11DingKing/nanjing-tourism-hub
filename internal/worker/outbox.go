package worker

import (
	"context"
	"fmt"
	"time"
)

func (r *Runner) processOutbox(ctx context.Context) error {
	now := time.Now().UTC()
	events, err := r.repo.ClaimOutbox(ctx, now, r.batch)
	if err != nil {
		return fmt.Errorf("claim outbox: %w", err)
	}
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := r.events.Publish(ctx, event.Topic, event.AggregateID, append([]byte(nil), event.Payload...))
		if err == nil {
			if completeErr := r.repo.CompleteOutbox(ctx, event.ID, time.Now().UTC()); completeErr != nil {
				return fmt.Errorf("complete outbox %s: %w", event.ID, completeErr)
			}
			continue
		}
		backoff := time.Duration(1<<min(event.Attempts, 8)) * time.Second
		if failErr := r.repo.FailOutbox(ctx, event.ID, err.Error(), time.Now().UTC().Add(backoff)); failErr != nil {
			return fmt.Errorf("fail outbox %s: %w", event.ID, failErr)
		}
	}
	return nil
}
