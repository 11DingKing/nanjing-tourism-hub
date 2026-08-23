package notifications

import (
	"context"
	"testing"
)

func TestDispatcherRetryAndIdempotency(t *testing.T) {
	sender := &MemorySender{Failures: 1}
	d := NewDispatcher()
	d.Senders[SMS] = sender
	if e := d.Dispatch(context.Background(), Message{ID: "m", Recipient: "138", Body: "warning", Channel: SMS}); e != nil {
		t.Fatal(e)
	}
	if len(sender.Sent) != 1 || sender.Sent[0].Attempts != 2 {
		t.Fatalf("sent=%+v", sender.Sent)
	}
	if e := d.Dispatch(context.Background(), Message{ID: "m", Recipient: "138", Body: "warning", Channel: SMS}); e != nil || len(sender.Sent) != 1 {
		t.Fatal("duplicate delivery")
	}
	if e := d.Dispatch(context.Background(), Message{ID: "m", Recipient: "138", Body: "different", Channel: SMS}); e == nil {
		t.Fatal("id reuse accepted")
	}
}
func TestRender(t *testing.T) {
	if Render("{{region}} {{level}}", map[string]string{"region": "North", "level": "high"}) != "North high" {
		t.Fatal("render")
	}
}
