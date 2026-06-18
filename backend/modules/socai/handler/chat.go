package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
)

type socAIStreamer interface {
	StreamAgentTask(ctx context.Context, body []byte) (*http.Response, error)
}

type ChatHandler struct {
	client socAIStreamer
}

func NewChatHandler(client socAIStreamer) *ChatHandler {
	return &ChatHandler{client: client}
}

type chatRequest struct {
	Task string `json:"task" binding:"required"`
	Page string `json:"page"`
	Lang string `json:"lang"`
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

	body, err := json.Marshal(map[string]string{"task": req.Task, "page": req.Page, "lang": req.Lang})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	resp, err := h.client.StreamAgentTask(c.Request.Context(), body)
	if err != nil {
		_ = catcher.Error("SocAIChat: stream request failed", err, nil)
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "SOC AI agent unreachable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": string(msg)})
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
