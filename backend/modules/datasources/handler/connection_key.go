package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
)

// ConnectionKeyHandler proxies the agent-installation connection key owned by the
// agent-manager (the agent-manager validates it locally on agent registration).
// It lives in the datasources module because agents are datasources.
type ConnectionKeyHandler struct {
	client *agentmanager.AgentManagerClient
}

func NewConnectionKeyHandler(client *agentmanager.AgentManagerClient) *ConnectionKeyHandler {
	return &ConnectionKeyHandler{client: client}
}

type connectionKeyResponse struct {
	ConnectionKey string `json:"connectionKey"`
}

// Get godoc
//
//	@Summary     Get the agent-installation connection key
//	@Tags        Datasources
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {object} connectionKeyResponse
//	@Failure     503 {object} map[string]string
//	@Router      /datasources/connection-key [get]
func (h *ConnectionKeyHandler) Get(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent-manager unavailable"})
		return
	}
	key, err := h.client.GetConnectionKey(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, connectionKeyResponse{ConnectionKey: key})
}

// Rotate godoc
//
//	@Summary     Rotate the agent-installation connection key
//	@Tags        Datasources
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {object} connectionKeyResponse
//	@Failure     503 {object} map[string]string
//	@Router      /datasources/connection-key/rotate [post]
func (h *ConnectionKeyHandler) Rotate(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent-manager unavailable"})
		return
	}
	key, err := h.client.RotateConnectionKey(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, connectionKeyResponse{ConnectionKey: key})
}
