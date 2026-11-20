package domain

import (
	"encoding/json"
	"regexp"
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
	Kind     NodeKind `yaml:"kind"                     json:"kind"`
	Executor string   `yaml:"executor"                 json:"executor"`
	Command  string   `yaml:"command,omitempty"        json:"command,omitempty"`
	Shell    string   `yaml:"shell,omitempty"          json:"shell,omitempty"`
	Platform string   `yaml:"platform,omitempty"       json:"platform,omitempty"`
	Agent    string   `yaml:"agent,omitempty"          json:"agent,omitempty"`
	// ExcludedAgents lists hostnames that must not run this node, even when
	// the resolved target (via Agent or the alert source) would otherwise pick
	// them. Applied only when Agent is empty (auto-resolve mode).
	ExcludedAgents []string `yaml:"excludedAgents,omitempty" json:"excludedAgents,omitempty"`
	// WaitFor is an explicit barrier list: the child waits for these parents
	// to respond (reach a terminal state) before running. When nil, the child
	// implicitly waits for every direct predecessor (full fan-in). Entries
	// must be direct predecessors — validateFlow enforces it.
	WaitFor   []string        `yaml:"waitFor,omitempty" json:"waitFor,omitempty"`
	Params    json.RawMessage `yaml:"-"                   json:"params,omitempty"`
	OnSuccess []string        `yaml:"onSuccess,omitempty"      json:"onSuccess,omitempty"`
	OnError   []string        `yaml:"onError,omitempty"        json:"onError,omitempty"`
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
		WaitFor        []string  `yaml:"waitFor"`
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
	n.WaitFor = s.WaitFor
	if len(n.WaitFor) == 0 {
		// Empty list == nil: the full fan-in default. YAML `waitFor: []`
		// must not silently narrow the gate to referenced parents only.
		n.WaitFor = nil
	}
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
	if len(n.WaitFor) > 0 {
		out["waitFor"] = n.WaitFor
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

// predecessorSet returns every node that has an edge into childID.
func (f Flow) predecessorSet(childID string) map[string]bool {
	out := make(map[string]bool)
	for id, n := range f.Nodes {
		for _, t := range n.OnSuccess {
			if t == childID {
				out[id] = true
			}
		}
		for _, t := range n.OnError {
			if t == childID {
				out[id] = true
			}
		}
	}
	return out
}

// WaitSet returns the parent node ids that must have responded (reached a
// terminal state) before the child may run. An explicit WaitFor wins over the
// implicit full fan-in, but any direct predecessor referenced via $(parent.*)
// in the node's interpolable text is always added — a node must never read a
// parent's value before that parent settles.
func (f Flow) WaitSet(childID string) []string {
	node, ok := f.Nodes[childID]
	if !ok {
		return nil
	}
	preds := f.predecessorSet(childID)
	set := make(map[string]bool)
	order := make([]string, 0, len(preds)+len(node.WaitFor))
	add := func(id string) {
		if !set[id] {
			set[id] = true
			order = append(order, id)
		}
	}
	if node.WaitFor == nil {
		for id := range preds {
			add(id)
		}
		return order
	}
	for _, id := range node.WaitFor {
		add(id)
	}
	for _, id := range node.ReferencedParents(preds) {
		add(id)
	}
	return order
}

// referencedParentRE grabs the first segment of a $(a.b) placeholder.
var referencedParentRE = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*)`) // $(first-segment...) in interpolable text

// ReferencedParents returns the allowed node ids referenced via $(id.*)
// placeholders anywhere in the node's interpolable text (command, shell, agent,
// params). Unknown first segments (alert, variables, ...) are dropped.
func (n FlowNode) ReferencedParents(allowed map[string]bool) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, 2)
	add := func(s string) {
		for _, m := range referencedParentRE.FindAllStringSubmatch(s, -1) {
			id := m[1]
			if allowed[id] && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	add(n.Command)
	add(n.Shell)
	add(n.Agent)
	if len(n.Params) > 0 {
		add(string(n.Params))
	}
	return out
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
