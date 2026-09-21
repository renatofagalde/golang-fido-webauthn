package flowid

import "context"

type contextKey string

const key contextKey = "flow_id"

// WithID injeta o flow-id no context. Chamado pelo middleware.
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}
