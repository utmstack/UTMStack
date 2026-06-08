package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

// Create and Update are internal-only (called by the SOAR plugin). They are not
// audited — high-frequency machine traffic, and the plugin is identified only as
// "internal" by the auth layer, so an audit row would carry no useful actor.

type ExecutionHandler struct {
	usecase connectors.ExecutionUsecase
}

func NewExecutionHandler(uc connectors.ExecutionUsecase) *ExecutionHandler {
	return &ExecutionHandler{usecase: uc}
}

// @Summary     List alert response rule executions
// @Tags        SOAR Executions
// @Security    BearerAuth
// @Produce     json
// @Param       page                    query int    false "Page number (0-based)"
// @Param       size                    query int    false "Page size"
// @Param       ruleId.equals           query int    false "Filter by rule ID"
// @Param       alertId.contains        query string false "Filter by alert ID (substring)"
// @Param       executionStatus.equals  query string false "Filter by execution status"
// @Success     200 {array}  dto.ExecutionResponse
// @Header      200 {integer} X-Total-Count "Total number of records"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rule-executions [get]
func (h *ExecutionHandler) List(c *gin.Context) {
	var f dto.ExecutionFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	writePagedArray(c, result.Items, result.Total)
}

// @Summary     Create alert response rule execution (internal)
// @Tags        SOAR Executions
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateExecutionRequest true "Execution to record (status forced to PENDING)"
// @Success     201 {object} dto.ExecutionResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rule-executions [post]
func (h *ExecutionHandler) Create(c *gin.Context) {
	var req dto.CreateExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Patch execution status (internal)
// @Tags        SOAR Executions
// @Accept      json
// @Param       id    path int                        true  "Execution ID"
// @Param       input body dto.UpdateExecutionRequest true  "Partial update — only non-nil fields are written"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rule-executions/{id} [patch]
func (h *ExecutionHandler) UpdateStatus(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.UpdateStatus(c.Request.Context(), id, req); err != nil {
		writeARRError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
