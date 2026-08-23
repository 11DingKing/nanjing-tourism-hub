package worker

import "time"

func Backoff(attempt int) time.Duration {
	if attempt < 1 {
		return time.Second
	}
	if attempt > 6 {
		attempt = 6
	}
	return time.Duration(1<<attempt) * time.Second
}
func Retryable(attempts int) bool { return attempts < 5 }
