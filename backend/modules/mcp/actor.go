package mcp

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// actorCtxKey is the context key under which the Gin lifter stashes the actor
// so MCP tool handlers can recover it.
type actorCtxKey struct{}

// ActorFromContext returns the actor stored in ctx, or nil if no lifter ran.
func ActorFromContext(ctx context.Context) *authz.Actor {
	if ctx == nil {
		return nil
	}
	a, _ := ctx.Value(actorCtxKey{}).(*authz.Actor)
	return a
}

// WithActor returns a context derived from parent with the actor attached.
// Exported for tests; the production path uses liftActorMiddleware.
func WithActor(parent context.Context, actor *authz.Actor) context.Context {
	return context.WithValue(parent, actorCtxKey{}, actor)
}

// liftActorMiddleware rebuilds the *authz.Actor from the gin context (set by
// middleware.Authenticate) and attaches it to the underlying *http.Request
// context. The SDK's HTTP handler reads ctx from req.Context(), so this is
// how the actor flows down into tool handlers.
func liftActorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := middleware.ActorFromGin(c)
		if actor != nil {
			c.Request = c.Request.WithContext(WithActor(c.Request.Context(), actor))
		}
		c.Next()
	}
}
