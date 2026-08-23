package httpapi

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

type contextKey string

const (
	actorKey     contextKey = "actor"
	requestIDKey contextKey = "request_id"
)

func withActor(ctx context.Context, actor domain.Actor) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}
func actorFrom(ctx context.Context) (domain.Actor, bool) {
	actor, ok := ctx.Value(actorKey).(domain.Actor)
	return actor, ok
}
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
func requestIDFrom(ctx context.Context) string { id, _ := ctx.Value(requestIDKey).(string); return id }
