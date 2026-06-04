package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/collectors/connectors"
	"github.com/utmstack/utmstack/backend/modules/collectors/dto"
)

type AgentHandler struct {
	uc connectors.AgentUsecase
}

func NewAgentHandler(uc connectors.AgentUsecase) *AgentHandler {
	return &AgentHandler{uc: uc}
}

// ListAgents godoc
// @Summary     List agents
// @Description Returns a paginated list of agents from agent-manager. AgentKey is always "SECRET".
// @Tags        Agents
// @Security    BearerAuth
// @Produce     json
// @Param       pageNumber   query int    false "Page number (1-based)"
// @Param       pageSize     query int    false "Page size (default 1000000)"
// @Param       searchQuery  query string false "Search query"
// @Param       sortBy       query string false "Sort field"
// @Success     200 {array}  dto.AgentResponse
// @Header      200 {string} X-Total-Count "Total number of agents"
// @Failure     500 {object} map[string]string
// @Router      /agent-manager/agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	var q dto.AgentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agents, total, err := h.uc.ListAgents(c.Request.Context(), q)
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	writePagedResponse(c, agents, total)
}

// ListAgentsWithCommands godoc
// @Summary     List agents with commands
// @Description Returns all agents at maximum page size (mirrors Java listAgentsWithCommands).
// @Tags        Agents
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array}  dto.AgentResponse
// @Header      200 {string} X-Total-Count "Total number of agents"
// @Failure     500 {object} map[string]string
// @Router      /agent-manager/agents-with-commands [get]
func (h *AgentHandler) ListAgentsWithCommands(c *gin.Context) {
	agents, total, err := h.uc.ListAgentsWithCommands(c.Request.Context())
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	writePagedResponse(c, agents, total)
}

// GetAgentByHostname godoc
// @Summary     Get agent by hostname
// @Description Returns the first agent matching the given hostname.
// @Tags        Agents
// @Security    BearerAuth
// @Produce     json
// @Param       hostname query string true "Agent hostname"
// @Success     200 {object} dto.AgentResponse
// @Failure     500 {object} map[string]string
// @Router      /agent-manager/agent-by-hostname [get]
func (h *AgentHandler) GetAgentByHostname(c *gin.Context) {
	hostname := c.Query("hostname")
	if hostname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hostname query parameter is required"})
		return
	}

	ag, err := h.uc.GetAgentByHostname(c.Request.Context(), hostname)
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	c.JSON(http.StatusOK, ag)
}

// CanRunCommand godoc
// @Summary     Check if agent can run a command
// @Description Returns true if the agent with the given hostname is ONLINE.
// @Tags        Agents
// @Security    BearerAuth
// @Produce     json
// @Param       hostname query string true "Agent hostname"
// @Success     200 {object} map[string]bool
// @Failure     500 {object} map[string]string
// @Router      /agent-manager/can-run-command [get]
func (h *AgentHandler) CanRunCommand(c *gin.Context) {
	hostname := c.Query("hostname")
	if hostname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hostname query parameter is required"})
		return
	}

	can, err := h.uc.CanRunCommand(c.Request.Context(), hostname)
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	c.JSON(http.StatusOK, can)
}

// ListAgentCommands godoc
// @Summary     List agent commands
// @Description Returns a paginated list of agent commands with nested agent info.
// @Tags        Agents
// @Security    BearerAuth
// @Produce     json
// @Param       pageNumber   query int    false "Page number (1-based)"
// @Param       pageSize     query int    false "Page size (default 10)"
// @Param       searchQuery  query string false "Search query"
// @Success     200 {array}  dto.AgentCommandResponse
// @Header      200 {string} X-Total-Count "Total number of commands"
// @Failure     500 {object} map[string]string
// @Router      /agent-manager/agent-commands [get]
func (h *AgentHandler) ListAgentCommands(c *gin.Context) {
	searchQuery := c.Query("searchQuery")

	page := int32(1)
	if v := c.Query("pageNumber"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			page = int32(n)
		}
	}
	size := int32(10)
	if v := c.Query("pageSize"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			size = int32(n)
		}
	}

	commands, total, err := h.uc.ListAgentCommands(c.Request.Context(), searchQuery, page, size)
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	writePagedResponse(c, commands, int32(total))
}

// UpdateAgentAttrs godoc
// @Summary     Update agent attributes
// @Description Updates the IP and hostname of an agent via agent-manager gRPC.
// @Tags        Agents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.UpdateAgentRequest true "Agent attributes to update"
// @Success     200 {object} dto.AgentAuthResponse
// @Failure     500 {object} map[string]string
// @Router      /agent-manager/update-agent-attrs [post]
func (h *AgentHandler) UpdateAgentAttrs(c *gin.Context) {
	var req dto.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.uc.UpdateAgentAttrs(c.Request.Context(), req)
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
