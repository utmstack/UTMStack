package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
)

type AgentHandler struct {
	usecase connectors.AgentUsecase
}

func NewAgentHandler(uc connectors.AgentUsecase) *AgentHandler {
	return &AgentHandler{usecase: uc}
}

// @Summary     List agent names by OS platform (internal)
// @Description Returns asset_name of every datasource with source_kind=agent and metadata.osPlatform=<platform>.
// @Tags        SOAR Agents
// @Produce     json
// @Param       platform query string true "OS platform (matches metadata.osPlatform)"
// @Success     200 {array} string
// @Header      200 {integer} X-Total-Count "Total number of records"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/agents [get]
func (h *AgentHandler) List(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform query parameter is required"})
		return
	}
	names, err := h.usecase.ListByPlatform(c.Request.Context(), platform)
	if err != nil {
		writeARRError(c, err)
		return
	}
	writePagedArray(c, names, int64(len(names)))
}
