package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
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
		c.Set("user_login", claims.Login)
		return true
	}

	if h.apiKeyAuth != nil {
		if apiKey := c.GetHeader("Utm-Api-Key"); apiKey != "" {
			actor, err := h.apiKeyAuth(c.Request.Context(), apiKey, c.ClientIP())
			if err != nil || actor == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return false
			}
			c.Set("user_login", actor.Login)
			return true
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
	Type    string `json:"type"` // "output" | "error" | "ready"
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

	if h.agentClient == nil {
		_ = wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: "agent manager unavailable"})
		_ = conn.Close(websocket.StatusInternalError, "no agent client")
		return
	}
	// No pre-lookup: the agentId comes from the datasource the user opened, and the
	// agent-manager validates it on ProcessCommandStream (offline/unknown → stream error,
	// surfaced over the WS). Liveness for the UI comes from datasources, not from here.

	// Persistent session: read a command, execute, emit output + ready, wait for next.
	// The socket stays open until the client disconnects or ctx is cancelled.
	for {
		var req wsCommandRequest
		if err := wsjson.Read(ctx, conn, &req); err != nil {
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
		if !h.runCommand(ctx, conn, cmd) {
			return
		}
	}
}

// runCommand executes one command over a fresh gRPC stream, forwards output to
// the WS, then emits "ready" so the client can send the next command. Returns
// false if the WS is unusable (write failed / ctx cancelled) — caller must exit.
func (h *CommandWSHandler) runCommand(ctx context.Context, conn *websocket.Conn, cmd *agent.UtmCommand) bool {
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultCh, errCh := h.agentClient.ProcessCommandStream(cmdCtx, cmd)

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				return writeReady(ctx, conn)
			}
			output := result.GetResult()
			if h.variableUC != nil {
				if masked, mErr := h.variableUC.MaskSecrets(output); mErr != nil {
					_ = catcher.Error("CommandStream: mask secrets", mErr, nil)
				} else {
					output = masked
				}
			}
			if err := wsjson.Write(ctx, conn, wsMessage{Type: "output", Data: output}); err != nil {
				return false
			}
			// ponytail: protocol is 1-command→1-result (agent-manager sends one Send per
			// command). First result IS end-of-stream — emit ready and return to caller,
			// deferred cancel unblocks the gRPC recv. Upgrade if streaming ever becomes real.
			return writeReady(ctx, conn)

		case grpcErr, ok := <-errCh:
			if !ok {
				return writeReady(ctx, conn)
			}
			if grpcErr != nil {
				_ = catcher.Error("CommandStream: grpc error", grpcErr, nil)
				if err := wsjson.Write(ctx, conn, wsMessage{Type: "error", Message: grpcErr.Error()}); err != nil {
					return false
				}
			}
			return writeReady(ctx, conn)

		case <-ctx.Done():
			return false
		}
	}
}

func writeReady(ctx context.Context, conn *websocket.Conn) bool {
	return wsjson.Write(ctx, conn, wsMessage{Type: "ready"}) == nil
}
