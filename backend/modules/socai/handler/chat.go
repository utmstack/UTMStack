package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
)

// classifyClientErr maps transport-level failures to a user-safe (httpStatus, message).
// Never echoes err.Error() — it can carry URLs, header values, or plugin internals.
func classifyClientErr(err error) (int, string) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "SOC AI agent timed out"
	case errors.Is(err, context.Canceled):
		return 499, "request canceled"
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return http.StatusGatewayTimeout, "SOC AI agent timed out"
	}
	return http.StatusBadGateway, "SOC AI agent unreachable"
}

// messageForStatus maps an upstream HTTP status to a fixed user-safe message,
// so upstream error bodies (which may contain the system prompt) never reach the client.
func messageForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return "SOC AI agent rejected credentials"
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return "SOC AI agent timed out"
	case http.StatusTooManyRequests:
		return "SOC AI agent rate limited"
	case http.StatusNotFound:
		return "SOC AI endpoint not found"
	case http.StatusBadGateway, http.StatusServiceUnavailable:
		return "SOC AI agent unreachable"
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "SOC AI agent rejected the request"
	}
	if status >= 500 {
		return "SOC AI agent internal error"
	}
	return "SOC AI agent error"
}

type socAIStreamer interface {
	StreamAgentTask(ctx context.Context, body []byte) (*http.Response, error)
}

type ChatHandler struct {
	client socAIStreamer
}

func NewChatHandler(client socAIStreamer) *ChatHandler {
	return &ChatHandler{client: client}
}

type chatTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Task    string     `json:"task" binding:"required"`
	Page    string     `json:"page"`
	Lang    string     `json:"lang"`
	History []chatTurn `json:"history,omitempty"`
}

// Chat godoc
//
// @Summary     Run a SOC AI agent task (streaming)
// @Description Proxies a free-form operations task to the soc-ai plugin agent and
// @Description streams its steps back as Server-Sent Events (tool_call, tool_result,
// @Description final, error). Chat-style: give the agent a task, watch it execute.
// @Tags        SOC AI
// @Security    BearerAuth
// @Accept      json
// @Produce     text/event-stream
// @Param       body body chatRequest true "Task to execute"
// @Success     200 {string} string "SSE stream"
// @Failure     400 {object} map[string]string
// @Failure     502 {object} map[string]string
// @Router      /soc-ai/chat [post]
func (h *ChatHandler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "task is required"})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "streaming unsupported"})
		return
	}

	body, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	resp, err := h.client.StreamAgentTask(c.Request.Context(), body)
	audit.Record(c, audit_connectors.Event{
		Action:       "socai.chat",
		ResourceType: "socai_task",
	}, audit_domain.SOCAI_CHAT_ATTEMPT, audit_domain.SOCAI_CHAT_SUCCESS, err)
	if err != nil {
		_ = catcher.Error("SocAIChat: stream request failed", err, nil)
		status, msg := classifyClientErr(err)
		c.JSON(status, gin.H{"status": "error", "message": msg})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Upstream error body may include the loaded plugin config (system prompt,
		// masked-but-still-sensitive fields). Log for ops, return a status-derived message.
		upstream, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = catcher.Error("SocAIChat: upstream error", nil, map[string]any{
			"status": resp.StatusCode,
			"body":   string(upstream),
		})
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": messageForStatus(resp.StatusCode)})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering
	c.Status(http.StatusOK)
	flusher.Flush()

	buf := make([]byte, 4096)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := c.Writer.Write(buf[:n]); werr != nil {
				return // client disconnected
			}
			flusher.Flush()
		}
		if rerr != nil {
			return // EOF or upstream/context error — stream is done
		}
	}
}
