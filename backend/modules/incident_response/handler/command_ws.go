package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/incident_response/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type CommandWSHandler struct {
	variableUC  connectors.VariableUsecase
	agentClient *agentmanager.AgentManagerClient
	signer      *jwtpkg.Signer
}

func NewCommandWSHandler(variableUC connectors.VariableUsecase, agentClient *agentmanager.AgentManagerClient, signer *jwtpkg.Signer) *CommandWSHandler {
	return &CommandWSHandler{
		variableUC:  variableUC,
		agentClient: agentClient,
		signer:      signer,
	}
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
//	             Secret variable values are masked in the output.
//	@Tags        Incident Command Stream
//	@Param       hostname path string true "Target agent hostname"
//	@Router      /ws/incident-command/{hostname} [get]
func (h *CommandWSHandler) CommandStream(c *gin.Context) {
	hostname := c.Param("hostname")
	if hostname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hostname is required"})
		return
	}

	rawToken := ""
	if header := c.GetHeader("Authorization"); header != "" {
		t, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || t == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		rawToken = t
	} else if q := c.Query("token"); q != "" {
		rawToken = q
	}
	if rawToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	if _, err := h.signer.Verify(rawToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow cross-origin in development
	})
	if err != nil {
		logger.Error(fmt.Sprintf("CommandStream: websocket accept: %s", err.Error()))
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

	interpolated, err := h.variableUC.InterpolateCommand(req.Command)
	if err != nil {
		logger.Error("CommandStream: interpolate: " + err.Error())
		_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "variable interpolation failed"})
		_ = conn.Close(websocket.StatusInternalError, "interpolation error")
		return
	}

	if h.agentClient == nil {
		_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "agent manager unavailable"})
		_ = conn.Close(websocket.StatusInternalError, "no agent client")
		return
	}
	agents, _, listErr := h.agentClient.ListAgents(ctx, fmt.Sprintf("hostname.Is=%s", hostname))
	if listErr != nil || len(agents) == 0 {
		_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "agent not found"})
		_ = conn.Close(websocket.StatusNormalClosure, "agent not found")
		return
	}
	ag := agents[0]

	if ag.GetStatus() != agent.Status_ONLINE {
		_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "agent is not online"})
		_ = conn.Close(websocket.StatusNormalClosure, "agent offline")
		return
	}

	cmd := &agent.UtmCommand{
		AgentId:    fmt.Sprintf("%d", ag.GetId()),
		Command:    interpolated,
		ExecutedBy: currentUser(c),
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
			raw := result.GetResult()
			masked, maskErr := h.variableUC.MaskSecrets(raw)
			if maskErr != nil {
				logger.Error("CommandStream: secret masking failed: " + maskErr.Error())
				_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "output masking failed"})
				cancel()
				_ = conn.Close(websocket.StatusInternalError, "masking error")
				return
			}
			msg := wsMessage{Type: "output", Data: masked}
			if writeErr := wsjson.Write(ctx, conn, msg); writeErr != nil {
				cancel()
				return
			}

		case grpcErr, ok := <-errCh:
			if !ok {
				return
			}
			if grpcErr != nil {
				logger.Error("CommandStream: grpc error: " + grpcErr.Error())
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
