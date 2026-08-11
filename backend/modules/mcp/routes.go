package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterRoutes mounts the MCP module under the given /api/v1 group.
//
// All MCP traffic flows through `/mcp`. The Streamable HTTP transport uses
// the same URL for both directions: POST for client→server messages and GET
// for the SSE stream the server uses to push responses and notifications.
//
// userAuth runs before the SDK handler — so by the time a tool handler
// executes, the Gin context carries a populated Actor. liftActorMiddleware
// then copies that Actor into the request context so the SDK propagates it
// down to the tool handler via ctx.
//
// The optional /mcp/health endpoint is open to authenticated callers and is
// the easiest way for operators to confirm the server is live and how many
// tools it has registered.
func RegisterRoutes(group *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	if m == nil || m.server == nil {
		return
	}

	mcpGroup := group.Group("/mcp", userAuth, liftActorMiddleware())

	mcpGroup.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"enabled":          true,
			"tools_registered": m.ToolCount(),
			"server_version":   m.deps.ServerVersion,
			"server_name":      m.deps.ServerName,
		})
	})

	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return m.server
	}, &mcp.StreamableHTTPOptions{
		// Backend may be reached behind reverse proxies whose forwarded
		// Host headers don't match the loopback addresses the SDK rebinds
		// against by default. We disable that protection because the
		// outer middleware already enforces JWT/API-key auth.
		DisableLocalhostProtection: true,
		// Stateless: every POST runs Connect() with the current request's
		// context, so authz.WithTenantID and the actor set by userAuth /
		// liftActorMiddleware reach the tool handler on THIS call — not
		// only on the one that opened the session. This is what makes
		// Utm-Tenant-Id switching per-request work for platform / support
		// callers, and what keeps a tenant JWT scoped to its own data.
		Stateless: true,
	})

	// Streamable HTTP needs both POST and GET on the same path.
	mcpGroup.POST("", gin.WrapH(handler))
	mcpGroup.GET("", gin.WrapH(handler))
	mcpGroup.DELETE("", gin.WrapH(handler))
}
