package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AdversaryHandler struct {
	usecase connectors.AdversaryUsecase
}

func NewAdversaryHandler(uc connectors.AdversaryUsecase) *AdversaryHandler {
	return &AdversaryHandler{usecase: uc}
}

// SearchAdversaryAlerts aggregates V11 alerts grouped by adversary host,
// falling back to adversary IP for alerts with no host (e.g. Internet-facing
// threats identified by IP rather than hostname).
//
// @Summary     Fetch adversary alerts grouped by adversary host (falls back to IP)
// @Tags        Alerts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       filters body []common_models.FilterType false "Optional filters"
// @Success     200 {array} dto.AdversaryResponse
// @Success     204 "No content — index missing or no results"
// @Failure     500 {object} map[string]string
// @Router      /adversary/alerts [post]
func (h *AdversaryHandler) SearchAdversaryAlerts(c *gin.Context) {
	var filters []common_models.FilterType
	_ = c.ShouldBindJSON(&filters)

	results, err := h.usecase.FetchAdversaryAlerts(c.Request.Context(), filters)
	if err != nil {
		_ = catcher.Error("adversary alerts", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
		return
	}

	if len(results) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, results)
}
