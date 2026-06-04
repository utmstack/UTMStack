package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type socAIAnalyzer interface {
	Analyze(ctx context.Context, alertJSON []byte) (int, []byte, error)
}

type SocAIHandler struct {
	client socAIAnalyzer
}

func NewSocAIHandler(client socAIAnalyzer) *SocAIHandler {
	return &SocAIHandler{client: client}
}

type alertIDEnvelope struct {
	ID interface{} `json:"id"`
}

// Analyze godoc
//
// @Summary     Submit an alert for SOC AI analysis
// @Description Forwards the alert JSON payload to the external SOC AI service.
// @Description The call is synchronous — "queued" in the response is a legacy label,
// @Description not an indication of async processing.
// @Tags        SOC AI
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       alert body object true "Full alert JSON"
// @Success     202 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soc-ai/analyze [post]
func (h *SocAIHandler) Analyze(c *gin.Context) {
	const ctx = "SocAIHandler.Analyze"

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, fmt.Sprintf("%s: read body: %s", ctx, err.Error()))
		return
	}

	// Java validates alert != null && alert.getId() != null → 400.
	var envelope alertIDEnvelope
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil || envelope.ID == nil {
		logger.Error(ctx + ": alert ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Alert ID is required"})
		return
	}

	statusCode, _, err := h.client.Analyze(c.Request.Context(), bodyBytes)
	if err != nil {
		writeError(c, fmt.Sprintf("%s: %s", ctx, err.Error()))
		return
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		writeError(c, fmt.Sprintf("%s: unexpected response from SOC AI service: %d", ctx, statusCode))
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "queued",
		"alertId": envelope.ID,
		"message": "Alert queued for SOC-AI analysis",
	})
}
