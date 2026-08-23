package batch

import (
	"context"
	"errors"
	"testing"
)

func TestBatchPartialResults(t *testing.T) {
	p := Processor[int]{Workers: 2, Handle: func(_ context.Context, v int) error {
		if v == 2 {
			return errors.New("bad")
		}
		return nil
	}}
	r := p.Run(context.Background(), []int{1, 2, 3})
	if len(r) != 3 {
		t.Fatalf("results=%d", len(r))
	}
	if Summarize(r) == nil {
		t.Fatal("failure hidden")
	}
}
