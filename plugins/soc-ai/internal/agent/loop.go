package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	defaultMaxIters     = 12
	compactionThreshold = 0.80
	summaryMaxTokens    = 400 // ~200 words + slack
)

var modelContextWindow = []struct {
	prefix string
	window int
}{
	{"claude", 200000},
	{"gpt-5", 400000},
	{"gpt-4o", 128000},
	{"gpt-4-turbo", 128000},
	{"gpt-4", 8192},
	{"gpt-3.5", 16385},
}

func resolveContextWindow(model string) int {
	m := strings.ToLower(model)
	for _, e := range modelContextWindow {
		if strings.HasPrefix(m, e.prefix) {
			return e.window
		}
	}
	return 128000
}

type EventKind string

const (
	EventToolCall    EventKind = "tool_call"
	EventToolResult  EventKind = "tool_result"
	EventFinal       EventKind = "final"
	EventError       EventKind = "error"
	EventCompaction  EventKind = "compaction"
)

type Event struct {
	Kind    EventKind       `json:"kind"`
	Step    int             `json:"step,omitempty"`
	Tool    string          `json:"tool,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Output  string          `json:"output,omitempty"`
	IsError bool            `json:"isError,omitempty"`
	Text    string          `json:"text,omitempty"`
}

type EventSink func(Event)

func (s EventSink) emit(e Event) {
	if s != nil {
		s(e)
	}
}

type RunTask struct {
	System        string // system prompt
	Input         string // the user turn (alert JSON for triage, free task for ops)
	EnabledGroups []string
	AlwaysAllow   []string
	MaxIters      int
}

type RunResult struct {
	Final     string
	Steps     int
	ToolCalls int
}

type Agent struct {
	llm           LLMClient
	broker        *ToolBroker
	model         string
	maxTokens     int
	contextWindow int // 0 = compaction disabled
}

func New(llm LLMClient, broker *ToolBroker, model string, maxTokens, contextWindow int) *Agent {
	cw := contextWindow
	if cw == 0 {
		cw = resolveContextWindow(model)
	} else if cw < 0 {
		cw = 0
	}
	return &Agent{llm: llm, broker: broker, model: model, maxTokens: maxTokens, contextWindow: cw}
}

func (a *Agent) Broker() *ToolBroker { return a.broker }

func (a *Agent) Run(ctx context.Context, task RunTask, sink EventSink) (RunResult, error) {
	specs, err := a.broker.ListSpecs(ctx)
	if err != nil {
		sink.emit(Event{Kind: EventError, Text: "could not load tools: " + err.Error()})
		return RunResult{}, fmt.Errorf("list tools: %w", err)
	}
	allowed := filterTools(specs, task)
	allowedSet := make(map[string]bool, len(allowed))
	for _, s := range allowed {
		allowedSet[s.Name] = true
	}

	maxIters := task.MaxIters
	if maxIters <= 0 {
		maxIters = defaultMaxIters
	}

	msgs := []Message{{Role: RoleUser, Content: task.Input}}
	result := RunResult{}

	for step := 1; step <= maxIters; step++ {
		result.Steps = step

		if a.contextWindow > 0 && len(msgs) > 1 &&
			estimateTokens(task.System, msgs) >= int(compactionThreshold*float64(a.contextWindow)) {
			newMsgs, cErr := a.compact(ctx, task.Input, msgs)
			if cErr != nil {
				_ = catcher.Error("context compaction failed, continuing with full history", cErr, map[string]any{
					"process": "plugin_com.utmstack.soc-ai",
				})
			} else {
				msgs = newMsgs
				sink.emit(Event{Kind: EventCompaction, Step: step})
			}
		}

		resp, err := a.llm.Complete(ctx, CompletionRequest{
			System:    task.System,
			Messages:  msgs,
			Tools:     allowed,
			Model:     a.model,
			MaxTokens: a.maxTokens,
		})
		if err != nil {
			sink.emit(Event{Kind: EventError, Text: err.Error()})
			return result, err
		}

		if len(resp.ToolCalls) == 0 {
			sink.emit(Event{Kind: EventFinal, Text: resp.Content})
			result.Final = resp.Content
			return result, nil
		}

		msgs = append(msgs, Message{Role: RoleAssistant, Content: resp.Content, ToolCalls: resp.ToolCalls})

		for _, tc := range resp.ToolCalls {
			result.ToolCalls++
			sink.emit(Event{Kind: EventToolCall, Step: step, Tool: tc.Name, Args: tc.Args})

			if !allowedSet[tc.Name] {
				const msg = "tool not permitted in this mode"
				msgs = append(msgs, Message{Role: RoleTool, ToolResult: &ToolResult{ID: tc.ID, Name: tc.Name, Content: msg, IsError: true}})
				sink.emit(Event{Kind: EventToolResult, Step: step, Tool: tc.Name, Output: msg, IsError: true})
				continue
			}

			out, isErr, callErr := a.broker.Call(ctx, tc.Name, tc.Args)
			if callErr != nil {
				out = callErr.Error()
				isErr = true
			}
			msgs = append(msgs, Message{Role: RoleTool, ToolResult: &ToolResult{ID: tc.ID, Name: tc.Name, Content: out, IsError: isErr}})
			sink.emit(Event{Kind: EventToolResult, Step: step, Tool: tc.Name, Output: out, IsError: isErr})
		}
	}

	const exhausted = "Reached the maximum number of tool iterations before finishing."
	sink.emit(Event{Kind: EventFinal, Text: exhausted})
	result.Final = exhausted
	return result, nil
}

func filterTools(specs []ToolSpec, task RunTask) []ToolSpec {
	enabled := make(map[string]bool, len(task.EnabledGroups))
	for _, g := range task.EnabledGroups {
		enabled[g] = true
	}
	always := make(map[string]bool, len(task.AlwaysAllow))
	for _, n := range task.AlwaysAllow {
		always[n] = true
	}
	out := make([]ToolSpec, 0, len(specs))
	for _, s := range specs {
		if s.ReadOnly || always[s.Name] {
			out = append(out, s)
			continue
		}
		if g := groupOf(s.Name); g != "" && enabled[g] {
			out = append(out, s)
		}
	}
	return out
}

func estimateTokens(system string, msgs []Message) int {
	n := len(system)
	for _, m := range msgs {
		n += len(m.Content)
		for _, tc := range m.ToolCalls {
			n += len(tc.Name) + len(tc.Args)
		}
		if m.ToolResult != nil {
			n += len(m.ToolResult.Content) + len(m.ToolResult.Name)
		}
	}
	return n / 4
}

func (a *Agent) compact(ctx context.Context, userInput string, msgs []Message) ([]Message, error) {
	var b strings.Builder
	for _, m := range msgs {
		fmt.Fprintf(&b, "[%s] %s\n", m.Role, m.Content)
		for _, tc := range m.ToolCalls {
			fmt.Fprintf(&b, "  tool_call %s(%s)\n", tc.Name, string(tc.Args))
		}
		if m.ToolResult != nil {
			fmt.Fprintf(&b, "  tool_result %s: %s\n", m.ToolResult.Name, m.ToolResult.Content)
		}
	}
	resp, err := a.llm.Complete(ctx, CompletionRequest{
		System:    "Summarize the following SOC analyst conversation in ~200 words. Preserve key facts, tool outputs, decisions, and unresolved next steps. Do not use tools.",
		Messages:  []Message{{Role: RoleUser, Content: b.String()}},
		Model:     a.model,
		MaxTokens: summaryMaxTokens,
	})
	if err != nil {
		return msgs, err
	}
	if strings.TrimSpace(resp.Content) == "" {
		return msgs, fmt.Errorf("empty summary")
	}
	return []Message{{
		Role:    RoleUser,
		Content: "Original task:\n" + userInput + "\n\nProgress so far (summary of prior context):\n" + resp.Content + "\n\nContinue the task.",
	}}, nil
}

var current atomic.Pointer[Agent]

func SetCurrent(a *Agent) {
	old := current.Swap(a)
	if old != nil && old.broker != nil && old != a {
		old.broker.Close()
	}
}

func Current() *Agent { return current.Load() }
