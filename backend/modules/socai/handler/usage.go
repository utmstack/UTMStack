package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type UsageHandler struct {
	limitOf func(ctx context.Context, tenantID string) (int, error)
	used    func(ctx context.Context, tenantID string) (int64, error)
}

func NewUsageHandler(
	limitOf func(ctx context.Context, tenantID string) (int, error),
	used func(ctx context.Context, tenantID string) (int64, error),
) *UsageHandler {
	return &UsageHandler{limitOf: limitOf, used: used}
}

type usageResponse struct {
	Limit     int       `json:"limit"` // 0 means no limit
	Used      int64     `json:"used"`
	Remaining *int64    `json:"remaining,omitempty"` // absent when there is no limit
	ResetsAt  time.Time `json:"resetsAt"`
}

// Usage godoc
//
//	@Summary		What this tenant has spent on AI today
//	@Description	The window is the calendar day in UTC. A limit of 0 means no limit, and remaining is then absent.
//	@Tags			SOC AI
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	usageResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/soc-ai/usage [get]
func (h *UsageHandler) Usage(c *gin.Context) {
	ctx := c.Request.Context()
	tenant := authz.TenantIDFromContext(ctx)

	resp := usageResponse{ResetsAt: time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)}
	if tenant == "" {
		c.JSON(http.StatusOK, resp)
		return
	}

	limit, err := h.limitOf(ctx, tenant)
	if err != nil {
		_ = catcher.Error("cannot read the tenant's AI limit", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "could not read the AI usage"})
		return
	}

	used, err := h.used(ctx, tenant)
	if err != nil {
		_ = catcher.Error("cannot read the tenant's AI usage", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "could not read the AI usage"})
		return
	}

	resp.Limit, resp.Used = limit, used
	if limit > 0 {
		remaining := max(int64(limit)-used, 0)
		resp.Remaining = &remaining
	}

	c.JSON(http.StatusOK, resp)
}
