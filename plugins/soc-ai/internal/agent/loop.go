package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
)

const defaultMaxIters = 12

type EventKind string

const (
	EventToolCall   EventKind = "tool_call"
	EventToolResult EventKind = "tool_result"
	EventFinal      EventKind = "final"
	EventError      EventKind = "error"
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
	llm       LLMClient
	broker    *ToolBroker
	model     string
	maxTokens int
}

func New(llm LLMClient, broker *ToolBroker, model string, maxTokens int) *Agent {
	return &Agent{llm: llm, broker: broker, model: model, maxTokens: maxTokens}
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

var current atomic.Pointer[Agent]

func SetCurrent(a *Agent) {
	old := current.Swap(a)
	if old != nil && old.broker != nil && old != a {
		old.broker.Close()
	}
}

func Current() *Agent { return current.Load() }
