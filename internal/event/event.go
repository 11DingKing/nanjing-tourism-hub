package event

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Event struct {
	ID, Topic, AggregateID string
	Payload                []byte
	OccurredAt             time.Time
	Attempt                int
}
type Handler func(context.Context, Event) error
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	seen     map[string]time.Time
}

func New() *Bus { return &Bus{handlers: map[string][]Handler{}, seen: map[string]time.Time{}} }
func (b *Bus) Subscribe(topic string, h Handler) error {
	if topic == "" || h == nil {
		return errors.New("topic and handler required")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], h)
	return nil
}
func (b *Bus) Publish(ctx context.Context, e Event) error {
	if e.ID == "" || e.Topic == "" {
		return errors.New("event identity required")
	}
	b.mu.Lock()
	if _, ok := b.seen[e.ID]; ok {
		b.mu.Unlock()
		return nil
	}
	b.seen[e.ID] = time.Now()
	handlers := append([]Handler(nil), b.handlers[e.Topic]...)
	b.mu.Unlock()
	for _, h := range handlers {
		if err := h(ctx, e); err != nil {
			return fmt.Errorf("handle %s: %w", e.Topic, err)
		}
	}
	return nil
}
func (b *Bus) Topics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := []string{}
	for topic := range b.handlers {
		out = append(out, topic)
	}
	return out
}
func (b *Bus) Prune(before time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, at := range b.seen {
		if at.Before(before) {
			delete(b.seen, id)
		}
	}
}

type Journal struct {
	mu     sync.Mutex
	Events []Event
}

func (j *Journal) Append(e Event) {
	j.mu.Lock()
	defer j.mu.Unlock()
	e.Payload = append([]byte(nil), e.Payload...)
	j.Events = append(j.Events, e)
}
func (j *Journal) Snapshot() []Event {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]Event, len(j.Events))
	copy(out, j.Events)
	for i := range out {
		out[i].Payload = append([]byte(nil), out[i].Payload...)
	}
	return out
}
