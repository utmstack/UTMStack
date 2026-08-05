package socai

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type AIQuota struct {
	LimitOf func(ctx context.Context, tenantID string) (int, error)
	Consume func(ctx context.Context, tenantID string) (int64, error)
	Used    func(ctx context.Context, tenantID string) (int64, error)
}

func (q *AIQuota) Gate() gin.HandlerFunc {
	if q == nil || q.LimitOf == nil || q.Consume == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		tenant := authz.TenantIDFromContext(ctx)
		if tenant == "" {
			c.Next()
			return
		}

		limit, err := q.LimitOf(ctx, tenant)
		if err != nil {
			_ = catcher.Error("cannot read the tenant's AI limit", err, map[string]any{"tenant": tenant})
			c.Next()
			return
		}
		if limit <= 0 {
			c.Next()
			return
		}

		used, err := q.Consume(ctx, tenant)
		if err != nil {
			_ = catcher.Error("cannot count an AI request", err, map[string]any{"tenant": tenant})
			c.Next()
			return
		}

		if used > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "this tenant has used the AI requests its plan allows for today",
				"limit":   limit,
				"used":    used,
			})
			return
		}

		c.Next()
	}
}
