package worker

import (
	"context"
	"fmt"
	"time"
)

func (r *Runner) processJobs(ctx context.Context) error {
	now := time.Now().UTC()
	jobs, err := r.repo.ClaimDueJobs(ctx, now, r.batch, 30*time.Second)
	if err != nil {
		return fmt.Errorf("claim jobs: %w", err)
	}
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := r.jobs.HandleJob(ctx, job.Kind, job.AggregateID, append([]byte(nil), job.Payload...))
		if err == nil {
			if completeErr := r.repo.CompleteJob(ctx, job.ID, time.Now().UTC()); completeErr != nil {
				return fmt.Errorf("complete job %s: %w", job.ID, completeErr)
			}
			continue
		}
		backoff := time.Duration(1<<min(job.Attempts, 8)) * time.Second
		if failErr := r.repo.FailJob(ctx, job.ID, err.Error(), time.Now().UTC().Add(backoff)); failErr != nil {
			return fmt.Errorf("fail job %s: %w", job.ID, failErr)
		}
	}
	return nil
}
