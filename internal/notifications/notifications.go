package notifications

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Channel string

const (
	SMS   Channel = "sms"
	Email Channel = "email"
	InApp Channel = "in_app"
)

type Message struct {
	ID, Recipient, Subject, Body string
	Channel                      Channel
	CreatedAt                    time.Time
	Attempts                     int
}
type Sender interface {
	Send(context.Context, Message) error
}
type MemorySender struct {
	mu       sync.Mutex
	Sent     []Message
	Failures int
}

func (m *MemorySender) Send(_ context.Context, v Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Failures > 0 {
		m.Failures--
		return errors.New("temporary sender failure")
	}
	m.Sent = append(m.Sent, v)
	return nil
}

type Dispatcher struct {
	Senders     map[Channel]Sender
	MaxAttempts int
	mu          sync.Mutex
	History     map[string]Message
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{Senders: map[Channel]Sender{}, MaxAttempts: 3, History: map[string]Message{}}
}
func (d *Dispatcher) Dispatch(ctx context.Context, v Message) error {
	if strings.TrimSpace(v.ID) == "" || strings.TrimSpace(v.Recipient) == "" {
		return errors.New("message identity required")
	}
	sender, ok := d.Senders[v.Channel]
	if !ok {
		return fmt.Errorf("channel %s unavailable", v.Channel)
	}
	d.mu.Lock()
	if old, ok := d.History[v.ID]; ok {
		d.mu.Unlock()
		if old.Body != v.Body {
			return errors.New("message id reused")
		}
		return nil
	}
	d.mu.Unlock()
	for attempt := 0; attempt < d.MaxAttempts; attempt++ {
		v.Attempts = attempt + 1
		if e := sender.Send(ctx, v); e == nil {
			d.mu.Lock()
			d.History[v.ID] = v
			d.mu.Unlock()
			return nil
		} else if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return errors.New("delivery exhausted")
}
func (d *Dispatcher) Delivered(id string) (Message, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	v, ok := d.History[id]
	return v, ok
}
func Render(template string, data map[string]string) string {
	out := template
	for k, v := range data {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
