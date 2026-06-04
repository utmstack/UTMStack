package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

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
