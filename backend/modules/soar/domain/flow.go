package domain

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

type NodeKind string

const (
	NodeKindExecutor   NodeKind = "executor"
	NodeKindEnrichment NodeKind = "enrichment"
)

const DefaultMaxDepth = 50

type Flow struct {
	Name        string              `yaml:"name"                  json:"name"`
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
	Conditions  []FilterType        `yaml:"conditions"            json:"conditions"`
	MaxDepth    int                 `yaml:"maxDepth,omitempty"    json:"maxDepth,omitempty"`
	Roots       []string            `yaml:"roots"                 json:"roots"`
	Nodes       map[string]FlowNode `yaml:"nodes"                 json:"nodes"`
}

type FlowNode struct {
	Kind      NodeKind        `yaml:"kind"                     json:"kind"`
	Executor  string          `yaml:"executor"                 json:"executor"`
	Command   string          `yaml:"command,omitempty"        json:"command,omitempty"`
	Shell     string          `yaml:"shell,omitempty"          json:"shell,omitempty"`
	Platform  string          `yaml:"platform,omitempty"       json:"platform,omitempty"`
	Agent     string          `yaml:"agent,omitempty"          json:"agent,omitempty"`
	// ExcludedAgents lists hostnames that must not run this node, even when
	// the resolved target (via Agent or the alert source) would otherwise pick
	// them. Applied only when Agent is empty (auto-resolve mode).
	ExcludedAgents []string        `yaml:"excludedAgents,omitempty" json:"excludedAgents,omitempty"`
	Params         json.RawMessage `yaml:"-"                        json:"params,omitempty"`
	OnSuccess      []string        `yaml:"onSuccess,omitempty"      json:"onSuccess,omitempty"`
	OnError        []string        `yaml:"onError,omitempty"        json:"onError,omitempty"`
}

// UnmarshalYAML lets nodes express Params as a native YAML mapping while the
// runtime still holds them as json.RawMessage (executors work on JSON, tests
// diff on JSON, and the DB column is jsonb). The YAML value is decoded into a
// generic any and re-encoded to JSON.
func (n *FlowNode) UnmarshalYAML(value *yaml.Node) error {
	type shadow struct {
		Kind           NodeKind  `yaml:"kind"`
		Executor       string    `yaml:"executor"`
		Command        string    `yaml:"command"`
		Shell          string    `yaml:"shell"`
		Platform       string    `yaml:"platform"`
		Agent          string    `yaml:"agent"`
		ExcludedAgents []string  `yaml:"excludedAgents"`
		Params         yaml.Node `yaml:"params"`
		OnSuccess      []string  `yaml:"onSuccess"`
		OnError        []string  `yaml:"onError"`
	}
	var s shadow
	if err := value.Decode(&s); err != nil {
		return err
	}
	n.Kind = s.Kind
	n.Executor = s.Executor
	n.Command = s.Command
	n.Shell = s.Shell
	n.Platform = s.Platform
	n.Agent = s.Agent
	n.ExcludedAgents = s.ExcludedAgents
	n.OnSuccess = s.OnSuccess
	n.OnError = s.OnError
	if s.Params.Kind == 0 {
		return nil
	}
	var raw any
	if err := s.Params.Decode(&raw); err != nil {
		return err
	}
	if raw == nil {
		return nil
	}
	raw = normalizeYAMLValue(raw)
	buf, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	n.Params = json.RawMessage(buf)
	return nil
}

// MarshalYAML emits Params as a plain mapping when it holds JSON, so writing a
// flow back to disk produces the same shape a human would type.
func (n FlowNode) MarshalYAML() (any, error) {
	out := map[string]any{
		"kind":     n.Kind,
		"executor": n.Executor,
	}
	if n.Command != "" {
		out["command"] = n.Command
	}
	if n.Shell != "" {
		out["shell"] = n.Shell
	}
	if n.Platform != "" {
		out["platform"] = n.Platform
	}
	if n.Agent != "" {
		out["agent"] = n.Agent
	}
	if len(n.ExcludedAgents) > 0 {
		out["excludedAgents"] = n.ExcludedAgents
	}
	if len(n.Params) > 0 {
		var v any
		if err := json.Unmarshal(n.Params, &v); err != nil {
			return nil, err
		}
		out["params"] = v
	}
	if len(n.OnSuccess) > 0 {
		out["onSuccess"] = n.OnSuccess
	}
	if len(n.OnError) > 0 {
		out["onError"] = n.OnError
	}
	return out, nil
}

// normalizeYAMLValue rewrites map[any]any (yaml.v3 default for untyped
// mappings) into map[string]any so encoding/json accepts it.
func normalizeYAMLValue(v any) any {
	switch t := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = normalizeYAMLValue(vv)
		}
		return out
	case map[string]any:
		for k, vv := range t {
			t[k] = normalizeYAMLValue(vv)
		}
		return t
	case []any:
		for i, vv := range t {
			t[i] = normalizeYAMLValue(vv)
		}
		return t
	}
	return v
}

// IncomingCounts returns, for every node id in the flow, how many other nodes
// reference it via on_success or on_error. AND-join sizing depends on this.
func (f Flow) IncomingCounts() map[string]int {
	counts := make(map[string]int, len(f.Nodes))
	for _, node := range f.Nodes {
		for _, id := range node.OnSuccess {
			counts[id]++
		}
		for _, id := range node.OnError {
			counts[id]++
		}
	}
	return counts
}

// ResolvedMaxDepth returns the flow's max_depth, falling back to DefaultMaxDepth.
func (f Flow) ResolvedMaxDepth() int {
	if f.MaxDepth > 0 {
		return f.MaxDepth
	}
	return DefaultMaxDepth
}

type StoredFlow struct {
	Flow
	RelPath  string    `yaml:"-"`
	Modified time.Time `yaml:"-"`
	System   bool      `yaml:"-"`
	Enabled  bool      `yaml:"-"`
	Tenant   string    `yaml:"-"`
}

func (sf StoredFlow) Active() bool { return sf.Enabled }

func (sf StoredFlow) SystemOwned() bool { return sf.System }
