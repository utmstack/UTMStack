package executor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// LLMStreamer is the narrow slice of the SOC-AI client that the LLM executors
// need. Defining it here keeps the executor package free of a socai import and
// lets tests inject a fake without spinning up an HTTP server.
type LLMStreamer interface {
	StreamAgentTask(ctx context.Context, body []byte) (*http.Response, error)
}

// LLM is one implementation backing two node types:
//   - llm_enrich (kind=enrichment): drives the SOC-AI agent with a prompt and
//     returns the final message parsed as JSON — becomes ancestor context for
//     downstream nodes.
//   - llm_action (kind=executor): drives the SOC-AI agent with a prompt so it
//     can use its own tools (list hosts, run commands, page oncall, etc.) and
//     succeeds when the stream ends on a `final` event.
type LLM struct {
	client LLMStreamer
	typ    string
}

// NewLLMEnrich registers a node type that expects a JSON `final` payload.
func NewLLMEnrich(c LLMStreamer) *LLM { return &LLM{client: c, typ: "llm_enrich"} }

// NewLLMAction registers a node type that treats the `final` payload as free
// text and only cares whether the stream ended cleanly.
func NewLLMAction(c LLMStreamer) *LLM { return &LLM{client: c, typ: "llm_action"} }

// enrichSystemPrompt is appended to the task of every llm_enrich execution. It
// pins the output shape: the node's `final` message must be a single JSON
// object with at least a `result` property (any JSON value), which becomes the
// node's output. Keep it in English regardless of the flow's lang — models
// follow a contract more reliably in their training language.
const enrichSystemPrompt = `OUTPUT CONTRACT (mandatory, overrides any conflicting instruction above):
Respond with EXACTLY ONE JSON object and nothing else — no prose before or after it, no markdown fences.
The object MUST contain a "result" property: {"result": ...}
- "result" may be a string, a JSON object, or a JSON array, carrying your complete finding.
- You may add extra sibling properties (e.g. "confidence") that downstream nodes will use.
Downstream automation parses this object verbatim; any other format fails the enrichment node.`

func (l *LLM) Type() string { return l.typ }

type llmParams struct {
	Prompt  string        `json:"prompt"`
	Page    string        `json:"page,omitempty"`
	Lang    string        `json:"lang,omitempty"`
	History []llmChatTurn `json:"history,omitempty"`
}

type llmChatTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (l *LLM) Execute(ctx context.Context, exec *domain.SoarExecution) (json.RawMessage, error) {
	if l.client == nil {
		return nil, errors.New("soar llm: SOC-AI client not configured")
	}
	var p llmParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar llm: params: %w", err)
		}
	}
	if strings.TrimSpace(p.Prompt) == "" {
		return nil, errors.New("soar llm: prompt is required")
	}

	// The SOC-AI client takes a single task body, so the enrichment output
	// contract travels inside the task. It is mandatory for this node type:
	// downstream nodes resolve $(<nodeId>.<field>) against the returned JSON,
	// so a missing or malformed `result` would leave them an empty bag.
	// llm_action leaves the task untouched — it only cares that the agent
	// finished cleanly.
	task := p.Prompt
	if exec.Kind == domain.NodeKindEnrichment {
		task = p.Prompt + "\n\n" + enrichSystemPrompt
	}

	body, err := json.Marshal(map[string]any{
		"task":    task,
		"page":    defaultString(p.Page, "soar"),
		"lang":    defaultString(p.Lang, "en"),
		"history": p.History,
	})
	if err != nil {
		return nil, fmt.Errorf("soar llm: build request: %w", err)
	}

	resp, err := l.client.StreamAgentTask(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("soar llm: stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("soar llm: upstream %d: %s", resp.StatusCode, string(raw))
	}

	events, finalRaw, errMsg, drainErr := drainSSE(resp.Body)
	exec.Result = truncate(events, 8192)
	if drainErr != nil {
		return nil, fmt.Errorf("soar llm: read stream: %w", drainErr)
	}
	if errMsg != "" {
		return nil, fmt.Errorf("soar llm: agent error: %s", errMsg)
	}
	if finalRaw == "" {
		return nil, errors.New("soar llm: stream ended without a final event")
	}

	if exec.Kind != domain.NodeKindEnrichment {
		// llm_action: success means the agent reported completion. Nothing
		// structured to hand downstream.
		return nil, nil
	}
	output, err := extractJSONOutput(finalRaw)
	if err != nil {
		return nil, fmt.Errorf("soar llm enrichment: final is not JSON: %w", err)
	}
	return output, nil
}

// drainSSE walks a text/event-stream body and returns the concatenated event
// log, the `data:` payload of the last `event: final`, and any `event: error`
// message. It stops on EOF or the first read error, mirroring the chat handler
// proxy — no reconnect logic.
func drainSSE(r io.Reader) (log string, finalData string, errMsg string, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(splitSSEFrames)

	var logBuf bytes.Buffer
	for scanner.Scan() {
		frame := scanner.Bytes()
		if logBuf.Len() > 0 {
			logBuf.WriteByte('\n')
		}
		logBuf.Write(frame)

		event, data := parseSSEFrame(frame)
		switch event {
		case "final":
			finalData = data
		case "error":
			errMsg = data
		}
	}
	if serr := scanner.Err(); serr != nil {
		return logBuf.String(), finalData, errMsg, serr
	}
	return logBuf.String(), finalData, errMsg, nil
}

// splitSSEFrames returns one SSE frame per Scanner call. Frames are separated
// by a blank line (`\n\n` or `\r\n\r\n`).
func splitSSEFrames(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		return i + 2, dropCR(data[:i]), nil
	}
	if i := bytes.Index(data, []byte("\r\n\r\n")); i >= 0 {
		return i + 4, dropCR(data[:i]), nil
	}
	if atEOF {
		return len(data), dropCR(data), nil
	}
	return 0, nil, nil
}

func dropCR(b []byte) []byte { return bytes.TrimRight(b, "\r") }

func parseSSEFrame(frame []byte) (event string, data string) {
	scanner := bufio.NewScanner(bytes.NewReader(frame))
	var dataBuf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	return event, dataBuf.String()
}

// extractJSONOutput accepts a few final-message shapes the SOC-AI agent tends
// to produce: bare JSON, a `content` field inside a JSON envelope, or a JSON
// blob wrapped in a ```json fence.
func extractJSONOutput(finalData string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(finalData)
	if trimmed == "" {
		return nil, errors.New("empty final message")
	}
	if raw, ok := tryJSON(trimmed); ok {
		// Envelope { "content": "..." } — unwrap and retry.
		var env struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(raw, &env); err == nil && strings.TrimSpace(env.Content) != "" {
			if inner, ok := tryJSON(strings.TrimSpace(env.Content)); ok {
				return inner, nil
			}
			if fenced, ok := stripJSONFence(env.Content); ok {
				return fenced, nil
			}
			return nil, fmt.Errorf("content is not JSON: %s", truncate(env.Content, 200))
		}
		return raw, nil
	}
	if fenced, ok := stripJSONFence(trimmed); ok {
		return fenced, nil
	}
	return nil, fmt.Errorf("not JSON: %s", truncate(trimmed, 200))
}

func tryJSON(s string) (json.RawMessage, bool) {
	var probe any
	if err := json.Unmarshal([]byte(s), &probe); err != nil {
		return nil, false
	}
	return json.RawMessage(s), true
}

func stripJSONFence(s string) (json.RawMessage, bool) {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "```") {
		return nil, false
	}
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return tryJSON(strings.TrimSpace(trimmed))
}

func defaultString(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
