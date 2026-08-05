package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
)

type CommandWSHandler struct {
	agentClient *agentmanager.AgentManagerClient
	signer      *jwtpkg.Signer
	variableUC  connectors.VariableUsecase
	apiKeyAuth  middleware.APIKeyAuthFunc
}

func NewCommandWSHandler(agentClient *agentmanager.AgentManagerClient, signer *jwtpkg.Signer, variableUC connectors.VariableUsecase) *CommandWSHandler {
	return &CommandWSHandler{
		agentClient: agentClient,
		signer:      signer,
		variableUC:  variableUC,
	}
}

func (h *CommandWSHandler) SetAPIKeyAuth(f middleware.APIKeyAuthFunc) {
	h.apiKeyAuth = f
}

func (h *CommandWSHandler) authenticate(c *gin.Context) bool {
	rawToken := ""
	if header := c.GetHeader("Authorization"); header != "" {
		t, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || t == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return false
		}
		rawToken = t
	} else if q := c.Query("token"); q != "" {
		rawToken = q
	}
	if rawToken != "" {
		claims, err := h.signer.Verify(rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return false
		}
		userID, err := claims.UserID()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
			return false
		}
		return middleware.EstablishActor(c, middleware.Actor{
			UserID:      userID,
			Login:       claims.Login,
			Email:       claims.Email,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
			SessionID:   claims.SessionID,
			TenantID:    claims.TenantID,
		})
	}

	if h.apiKeyAuth != nil {
		if apiKey := c.GetHeader("Utm-Api-Key"); apiKey != "" {
			actor, err := h.apiKeyAuth(c.Request.Context(), apiKey, c.ClientIP())
			if err != nil || actor == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return false
			}
			return middleware.EstablishActor(c, *actor)
		}
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
	return false
}

type wsCommandRequest struct {
	Command    string `json:"command"`
	OriginType string `json:"originType"`
	OriginID   string `json:"originId"`
	Reason     string `json:"reason"`
	Shell      string `json:"shell"`
}

type wsMessage struct {
	Type    string `json:"type"` // "output" | "error" | "done"
	Data    string `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// CommandStream godoc
//
//	@Summary     Stream command output via WebSocket
//	@Description Upgrades to WebSocket. Client sends one JSON command request,
//	             server streams output chunks from the agent in real time.
//	@Tags        Incident Command Stream
//	@Param       agentId path string true "Target agent id (agent-manager id; from the datasource's metadata.agentId)"
//	@Router      /soar/ws/command/{agentId} [get]
func (h *CommandWSHandler) CommandStream(c *gin.Context) {
	agentID := c.Param("agentId")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}

	if !h.authenticate(c) {
		return
	}

	actor := middleware.ActorFromGin(c)
	if err := authz.RequirePermission(actor, "soar.write"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "missing required permission: soar.write"})
		return
	}

	if !h.ownsAgent(c, agentID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow cross-origin in development
	})
	if err != nil {
		_ = catcher.Error("CommandStream: websocket accept", err, nil)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	var req wsCommandRequest
	if err := wsjson.Read(ctx, conn, &req); err != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "failed to read command")
		return
	}

	if h.agentClient == nil {
		_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "agent manager unavailable"})
		_ = conn.Close(websocket.StatusInternalError, "no agent client")
		return
	}
	command := req.Command
	if h.variableUC != nil {
		if interpolated, vErr := h.variableUC.InterpolateCommand(command); vErr != nil {
			_ = catcher.Error("CommandStream: variable interpolation", vErr, nil)
		} else {
			command = interpolated
		}
	}

	cmd := &agent.UtmCommand{
		AgentId:    agentID,
		Command:    command,
		ExecutedBy: loginFromCtx(c),
		OriginType: req.OriginType,
		OriginId:   req.OriginID,
		Reason:     req.Reason,
		Shell:      req.Shell,
	}
	resultCh, errCh := h.agentClient.ProcessCommandStream(ctx, cmd)

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				_ = wsjson.Write(ctx, conn, wsMessage{Type: "done"})
				_ = conn.Close(websocket.StatusNormalClosure, "stream complete")
				return
			}
			output := result.GetResult()
			if h.variableUC != nil {
				if masked, mErr := h.variableUC.MaskSecrets(output); mErr != nil {
					_ = catcher.Error("CommandStream: mask secrets", mErr, nil)
				} else {
					output = masked
				}
			}
			msg := wsMessage{Type: "output", Data: output}
			if writeErr := wsjson.Write(ctx, conn, msg); writeErr != nil {
				cancel()
				return
			}

		case grpcErr, ok := <-errCh:
			if !ok {
				return
			}
			if grpcErr != nil {
				_ = catcher.Error("CommandStream: grpc error", grpcErr, nil)
				errMsg, _ := json.Marshal(wsMessage{Type: "error", Message: grpcErr.Error()})
				_ = conn.Write(ctx, websocket.MessageText, errMsg)
				_ = conn.Close(websocket.StatusInternalError, "grpc error")
			}
			return

		case <-ctx.Done():
			return
		}
	}
}

// ownsAgent reports whether the agent belongs to the tenant making the request.
func (h *CommandWSHandler) ownsAgent(c *gin.Context, agentID string) bool {
	if h.agentClient == nil {
		return false
	}
	rows, _, err := h.agentClient.ListAgents(c.Request.Context(), "id.Is="+agentID)
	if err != nil {
		_ = catcher.Error("CommandStream: cannot verify the agent's tenant", err, map[string]any{"agent": agentID})
		return false
	}
	return len(rows) > 0
}
