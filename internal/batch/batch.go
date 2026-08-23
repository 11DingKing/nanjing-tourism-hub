package batch

import (
	"context"
	"fmt"
	"sync"
)

type Result[T any] struct {
	Item T
	Err  error
}
type Processor[T any] struct {
	Workers int
	Handle  func(context.Context, T) error
}

func (p Processor[T]) Run(ctx context.Context, items []T) []Result[T] {
	workers := p.Workers
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan T)
	results := make(chan Result[T], len(items))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				if e := p.Handle(ctx, item); e != nil {
					results <- Result[T]{Item: item, Err: e}
				} else {
					results <- Result[T]{Item: item}
				}
			}
		}()
	}
	for _, item := range items {
		select {
		case jobs <- item:
		case <-ctx.Done():
			results <- Result[T]{Item: item, Err: ctx.Err()}
		}
	}
	close(jobs)
	wg.Wait()
	close(results)
	out := make([]Result[T], 0, len(items))
	for result := range results {
		out = append(out, result)
	}
	return out
}
func Summarize[T any](results []Result[T]) error {
	for _, result := range results {
		if result.Err != nil {
			return fmt.Errorf("batch item failed: %w", result.Err)
		}
	}
	return nil
}
