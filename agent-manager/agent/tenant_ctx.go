package agent

import "context"

// Handlers on the key/internal-key paths must fall back to the request
// payload (or tenantOrDefault) if tenantFromContext returns false.

type tenantCtxKey struct{}

func withTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantCtxKey{}, tenantID)
}

func tenantFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tenantCtxKey{}).(string)
	if !ok || t == "" {
		return "", false
	}
	return t, true
}
