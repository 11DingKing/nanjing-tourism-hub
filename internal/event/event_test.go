package event

import (
	"context"
	"testing"
	"time"
)

func TestBusIdempotencyAndJournalCopy(t *testing.T) {
	b := New()
	count := 0
	if e := b.Subscribe("alert", func(_ context.Context, e Event) error { count++; return nil }); e != nil {
		t.Fatal(e)
	}
	e := Event{ID: "1", Topic: "alert", Payload: []byte("x")}
	if err := b.Publish(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if err := b.Publish(context.Background(), e); err != nil || count != 1 {
		t.Fatal("duplicate")
	}
	j := &Journal{}
	j.Append(e)
	s := j.Snapshot()
	s[0].Payload[0] = 'y'
	if j.Snapshot()[0].Payload[0] != 'x' {
		t.Fatal("payload alias")
	}
	b.Prune(time.Now().Add(time.Hour))
}
