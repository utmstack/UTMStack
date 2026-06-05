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
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
)

type CommandWSHandler struct {
	agentClient *agentmanager.AgentManagerClient
	signer      *jwtpkg.Signer
	variableUC  connectors.VariableUsecase
}

func NewCommandWSHandler(agentClient *agentmanager.AgentManagerClient, signer *jwtpkg.Signer, variableUC connectors.VariableUsecase) *CommandWSHandler {
	return &CommandWSHandler{
		agentClient: agentClient,
		signer:      signer,
		variableUC:  variableUC,
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
//	@Tags        Incident Command Stream
//	@Param       hostname path string true "Target agent hostname"
//	@Router      /soar/ws/command/{hostname} [get]
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

	command := req.Command
	if h.variableUC != nil {
		if interpolated, vErr := h.variableUC.InterpolateCommand(command); vErr != nil {
			_ = catcher.Error("CommandStream: variable interpolation", vErr, nil)
		} else {
			command = interpolated
		}
	}

	cmd := &agent.UtmCommand{
		AgentId:    fmt.Sprintf("%d", ag.GetId()),
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
