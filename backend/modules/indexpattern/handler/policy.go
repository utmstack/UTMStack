package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/connectors"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
)

type PolicyHandler struct {
	usecase connectors.PolicyUsecase
}

func NewPolicyHandler(uc connectors.PolicyUsecase) *PolicyHandler {
	return &PolicyHandler{usecase: uc}
}

// @Summary     Get index policy settings (returns empty body if no policy configured)
// @Tags        Index Policy
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} domain.PolicySettings
// @Failure     500 {object} map[string]string
// @Router      /index-policy/policy [get]
func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	settings, err := h.usecase.GetPolicy(c.Request.Context())
	if err != nil {
		writeIPError(c, err)
		return
	}
	if settings == nil {
		c.Status(http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, settings)
}

// @Summary     Update index policy settings
// @Tags        Index Policy
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body domain.PolicySettings true "Policy settings (snapshotActive bool, deleteAfter string)"
// @Success     200 {object} domain.UpdateManagedIndexPolicyResponse
// @Failure     500 {object} map[string]string
// @Router      /index-policy/policy [put]
func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	var req domain.PolicySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.UpdatePolicy(c.Request.Context(), req)
	if err != nil {
		writeIPError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
