package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/usecase"
)

type PlaygroundHandler struct {
	usecase connectors.PlaygroundUsecase
}

func NewPlaygroundHandler(uc connectors.PlaygroundUsecase) *PlaygroundHandler {
	return &PlaygroundHandler{usecase: uc}
}

func (h *PlaygroundHandler) TestPipeline(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, usecase.MaxPlaygroundBodyBytes)

	var req dto.TestPipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.TestPipeline(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{
		Action:       "eventprocessing.playground.test_pipeline",
		ResourceType: "playground",
		Metadata:     map[string]any{"had_custom_content": req.Pipeline != nil},
	}, audit_domain.PLAYGROUND_TEST_FILTER_ATTEMPT, audit_domain.PLAYGROUND_TEST_FILTER_SUCCESS, err)

	if err != nil {
		writePlaygroundError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PlaygroundHandler) TestRule(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, usecase.MaxPlaygroundBodyBytes)

	var req dto.TestRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.TestRule(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{
		Action:       "eventprocessing.playground.test_rule",
		ResourceType: "playground",
		Metadata:     map[string]any{"had_custom_content": req.Rule != nil},
	}, audit_domain.PLAYGROUND_TEST_RULE_ATTEMPT, audit_domain.PLAYGROUND_TEST_RULE_SUCCESS, err)

	if err != nil {
		writePlaygroundError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
