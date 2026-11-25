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

// Verdict is what the quota says about one more request from a tenant.
type Verdict struct {
	Allowed bool
	Limit   int
	Used    int64
}

// Check counts one request against the tenant's allowance for today and says whether it fits. It is
// what the HTTP gate and the automatic triage both ask, so both give the same answer. When the
// allowance cannot be read it lets the request through: a broken counter must not stop the agent.
func (q *AIQuota) Check(ctx context.Context, tenant string) Verdict {
	if q == nil || q.LimitOf == nil || q.Consume == nil || tenant == "" {
		return Verdict{Allowed: true}
	}

	limit, err := q.LimitOf(ctx, tenant)
	if err != nil {
		_ = catcher.Error("cannot read the tenant's AI limit", err, map[string]any{"tenant": tenant})

		return Verdict{Allowed: true}
	}
	// Negative is uncapped. Zero is an allowance of none, which denies: the
	// operator handing out none of the instance's quota must not read as
	// handing out all of it.
	if limit < 0 {
		return Verdict{Allowed: true, Limit: limit}
	}

	used, err := q.Consume(ctx, tenant)
	if err != nil {
		_ = catcher.Error("cannot count an AI request", err, map[string]any{"tenant": tenant})

		return Verdict{Allowed: true, Limit: limit}
	}

	return Verdict{Allowed: limit != 0 && used <= int64(limit), Limit: limit, Used: used}
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

		if v := q.Check(ctx, tenant); !v.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "this tenant has used the AI requests its plan allows for today",
				"limit":   v.Limit,
				"used":    v.Used,
			})
			return
		}

		c.Next()
	}
}
