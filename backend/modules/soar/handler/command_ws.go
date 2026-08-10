package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
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
	executionUC connectors.ExecutionUsecase
	apiKeyAuth  middleware.APIKeyAuthFunc
}

func NewCommandWSHandler(
	agentClient *agentmanager.AgentManagerClient,
	signer *jwtpkg.Signer,
	variableUC connectors.VariableUsecase,
	executionUC connectors.ExecutionUsecase,
) *CommandWSHandler {
	return &CommandWSHandler{
		agentClient: agentClient,
		signer:      signer,
		variableUC:  variableUC,
		executionUC: executionUC,
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
		interpolated, vErr := h.variableUC.InterpolateCommand(ctx, command)
		if vErr != nil {
			// Sending it anyway would put a literal "$[variables.NAME]" on the
			// agent's command line as an argument.
			_ = catcher.Error("CommandStream: variable interpolation", vErr, nil)
			_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "a variable in this command could not be resolved"})
			_ = conn.Close(websocket.StatusInternalError, "interpolation failed")
			return
		}
		command = interpolated
	}

	// The record opens before the command is sent, so a stream that dies still
	// leaves evidence that someone ran this here. Failing to record does not
	// stop the command: the operator is entitled to their console.
	executedBy := loginFromCtx(c)
	var execID uuid.UUID
	if h.executionUC != nil {
		id, recErr := h.executionUC.StartManual(ctx, agentID, command, executedBy)
		if recErr != nil {
			_ = catcher.Error("CommandStream: recording the execution", recErr, nil)
		} else {
			execID = id
		}
	}
	var collected strings.Builder
	finish := func(status domain.ExecutionStatus) {
		if h.executionUC == nil || execID == uuid.Nil {
			return
		}
		// A fresh context: ctx is cancelled by the time this runs on a
		// disconnect, and the row still has to be closed.
		done, cancelDone := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelDone()
		_ = h.executionUC.FinishManual(done, execID, status, collected.String())
	}

	cmd := &agent.UtmCommand{
		AgentId:    agentID,
		Command:    command,
		ExecutedBy: executedBy,
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
				finish(domain.ExecutionStatusExecuted)
				_ = wsjson.Write(ctx, conn, wsMessage{Type: "done"})
				_ = conn.Close(websocket.StatusNormalClosure, "stream complete")
				return
			}
			output := result.GetResult()
			if h.variableUC != nil {
				if masked, mErr := h.variableUC.MaskSecrets(ctx, output); mErr != nil {
					_ = catcher.Error("CommandStream: mask secrets", mErr, nil)
				} else {
					output = masked
				}
			}
			collected.WriteString(output)
			msg := wsMessage{Type: "output", Data: output}
			if writeErr := wsjson.Write(ctx, conn, msg); writeErr != nil {
				// The viewer went away; the command did not.
				finish(domain.ExecutionStatusExecuted)
				cancel()
				return
			}

		case grpcErr, ok := <-errCh:
			if !ok {
				finish(domain.ExecutionStatusExecuted)
				return
			}
			if grpcErr != nil {
				finish(domain.ExecutionStatusFailed)
				_ = catcher.Error("CommandStream: grpc error", grpcErr, nil)
				errMsg, _ := json.Marshal(wsMessage{Type: "error", Message: grpcErr.Error()})
				_ = conn.Write(ctx, websocket.MessageText, errMsg)
				_ = conn.Close(websocket.StatusInternalError, "grpc error")
			}
			return

		case <-ctx.Done():
			finish(domain.ExecutionStatusExecuted)
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
