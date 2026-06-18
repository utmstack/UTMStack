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
// @Param       rulePath.equals         query string false "Filter by flow path (rule identity)"
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

// @Summary     Report a flow match (internal)
// @Description The active-response plugin posts {rulePath, alert} when an alert matched a flow's conditions. The backend resolves the target agent, builds the command(s) and enqueues the execution(s) for the dispatcher.
// @Tags        SOAR Executions
// @Accept      json
// @Param       input body dto.MatchRequest true "Matched flow + raw alert"
// @Success     202
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rule-executions [post]
func (h *ExecutionHandler) Match(c *gin.Context) {
	var req dto.MatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.HandleMatch(c.Request.Context(), req); err != nil {
		writeARRError(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

