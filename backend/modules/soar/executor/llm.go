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
//     returns its final message normalized to {"result": ...} — becomes
//     ancestor context for downstream nodes.
//   - llm_action (kind=executor): drives the SOC-AI agent with a prompt so it
//     can use its own tools (list hosts, run commands, page oncall, etc.) and
//     succeeds when the stream ends on a `final` event.
type LLM struct {
	client LLMStreamer
	typ    string
}

// NewLLMEnrich registers a node type that normalizes its output to {"result": ...}.
func NewLLMEnrich(c LLMStreamer) *LLM { return &LLM{client: c, typ: "llm_enrich"} }

// NewLLMAction registers a node type that treats the `final` payload as free
// text and only cares whether the stream ended cleanly.
func NewLLMAction(c LLMStreamer) *LLM { return &LLM{client: c, typ: "llm_action"} }

// enrichSystemPrompt is appended to the task of every llm_enrich execution.
// It pins the output shape; the backend then enforces it in
// normalizeEnrichmentOutput, so downstream nodes may always rely on
// $(<nodeId>.result). Keep it in English regardless of the flow's lang —
// models follow a contract more reliably in their training language.
const enrichSystemPrompt = `OUTPUT CONTRACT (mandatory, overrides any conflicting instruction above):
Respond with EXACTLY ONE JSON object and nothing else - no prose before or after it, no markdown fences.
The object MUST contain a "result" property holding your complete finding:
  {"result": ...}
- "result" may be a string, a JSON object, or a JSON array.
- You may add a few sibling properties (e.g. "confidence").
Downstream automation resolves <this-node-id>.result from your object verbatim.`

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
	// downstream nodes resolve $(<nodeId>.result) against the normalized
	// output. llm_action leaves the task untouched — it only cares that the
	// agent finished cleanly.
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
	return normalizeEnrichmentOutput(finalRaw)
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

// normalizeEnrichmentOutput is the backend-side half of the enrichment
// contract: whatever the model returns, the node output is ALWAYS a JSON
// object whose "result" property carries the finding, so downstream nodes can
// unconditionally reference $(<nodeId>.result).
//
//   - JSON object already carrying "result"      -> passed through unchanged
//     (siblings such as "confidence" survive).
//   - JSON object without "result"               -> encapsulated: the whole
//     object becomes the "result" value.
//   - JSON array or JSON scalar                  -> encapsulated as "result".
//   - JSON hidden in a {"content": "..."} envelope
//     or a ``` fence (model/transport quirk)     -> unwrapped first, then
//     re-run through the same rules.
//   - plain text (contract ignored)              -> {"result": "<text>"}.
func normalizeEnrichmentOutput(finalRaw string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(finalRaw)
	if trimmed == "" {
		return nil, errors.New("soar llm enrichment: empty final message")
	}

	if raw, ok := tryJSON(trimmed); ok {
		return finishFromJSON(unwrapEnvelope(raw))
	}
	if fenced, ok := stripJSONFence(trimmed); ok {
		return finishFromJSON(unwrapEnvelope(fenced))
	}
	// No JSON at all: the model answered in prose. Still succeed — the whole
	// text becomes "result". The contract prompt exists to avoid this branch.
	quoted, err := json.Marshal(trimmed)
	if err != nil {
		return nil, fmt.Errorf("soar llm enrichment: encapsulate result: %w", err)
	}
	return appendJSONValue(quoted), nil
}

// finishFromJSON passes a JSON value through when it is already an object
// carrying "result"; otherwise it encapsulates the value under "result".
func finishFromJSON(raw json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err == nil {
		if _, has := m["result"]; has {
			return raw, nil
		}
	}
	return appendJSONValue(raw), nil
}

// unwrapEnvelope resolves a JSON value to its payload: a {"content": "..."}
// envelope whose content is JSON (possibly fenced) is unwrapped to that inner
// value; content that is plain prose becomes a JSON string. The model
// sometimes wraps its answer in the chat-message envelope instead of sending
// the object directly.
func unwrapEnvelope(raw json.RawMessage) json.RawMessage {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw // array or scalar — nothing to unwrap
	}
	c, ok := m["content"]
	if !ok {
		return raw
	}
	var cs string
	if err := json.Unmarshal(c, &cs); err != nil {
		return raw // content is not a string — leave the envelope as-is
	}
	cs = strings.TrimSpace(cs)
	if cs == "" {
		return raw
	}
	if inner, isJSON := tryJSON(cs); isJSON {
		return inner
	}
	if fenced, isJSON := stripJSONFence(cs); isJSON {
		return fenced
	}
	quoted, err := json.Marshal(cs)
	if err != nil {
		return raw
	}
	return quoted
}

// appendJSONValue wraps a JSON value under "result", keeping objects, arrays
// and scalars as real JSON values (not escaped strings).
func appendJSONValue(raw json.RawMessage) json.RawMessage {
	out := make([]byte, 0, len(raw)+12)
	out = append(out, `{"result":`...)
	out = append(out, raw...)
	out = append(out, '}')
	return out
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

func defaultString(s string, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
