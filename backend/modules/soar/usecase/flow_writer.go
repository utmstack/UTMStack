package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

func writeFlowFile(path string, flow domain.Flow) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal([]domain.Flow{flow})
	if err != nil {
		return err
	}

	tf, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := tf.Name()
	defer func() { _ = os.Remove(tmp) }()

	if _, err := tf.Write(data); err != nil {
		_ = tf.Close()
		return err
	}
	if err := tf.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// readFlowFile reads one Flow from disk. It accepts the current DAG shape
// (`roots:` + `nodes:`) and transparently upgrades the legacy chain shape
// (`commands:` + `shell:`) so existing on-disk playbooks keep parsing.
func readFlowFile(path string) (domain.Flow, error) {
	var flow domain.Flow
	data, err := os.ReadFile(path)
	if err != nil {
		return flow, err
	}
	var list []legacyOrDAGFlow
	if err := yaml.Unmarshal(data, &list); err != nil {
		return flow, err
	}
	if len(list) == 0 {
		return flow, fmt.Errorf("flow file %s contains no flows", path)
	}
	return list[0].asFlow()
}

// legacyOrDAGFlow captures both YAML shapes so we can pick which one this file
// used and normalize to the DAG shape. It carries every field either shape
// recognises; the mapper checks whether nodes/roots or commands is populated
// and produces a canonical domain.Flow.
type legacyOrDAGFlow struct {
	Name        string                     `yaml:"name"`
	Description string                     `yaml:"description,omitempty"`
	Conditions  []domain.FilterType        `yaml:"conditions"`
	MaxDepth    int                        `yaml:"maxDepth,omitempty"`
	Roots       []string                   `yaml:"roots,omitempty"`
	Nodes       map[string]domain.FlowNode `yaml:"nodes,omitempty"`

	// Legacy chain shape. Platform + default agent are copied into every
	// legacy shell node so the upgrade preserves behavior.
	Commands       []legacyCommand `yaml:"commands,omitempty"`
	Shell          string          `yaml:"shell,omitempty"`
	AgentPlatform  string          `yaml:"agentPlatform,omitempty"`
	DefaultAgent   string          `yaml:"defaultAgent,omitempty"`
	ExcludedAgents []string        `yaml:"excludedAgents,omitempty"`
}

type legacyCommand struct {
	Command   string  `yaml:"command"`
	Condition *string `yaml:"condition,omitempty"`
}

// UnmarshalYAML lets old files write a bare string for a command, matching the
// previous behaviour where `- systemctl restart wazuh` was valid.
func (lc *legacyCommand) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		lc.Command = value.Value
		return nil
	}
	type raw legacyCommand
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*lc = legacyCommand(r)
	return nil
}

func (l legacyOrDAGFlow) asFlow() (domain.Flow, error) {
	base := domain.Flow{
		Name:        l.Name,
		Description: l.Description,
		Conditions:  l.Conditions,
		MaxDepth:    l.MaxDepth,
	}
	if len(l.Nodes) > 0 {
		base.Roots = l.Roots
		base.Nodes = l.Nodes
		return base, nil
	}
	// ponytail: legacy chain → linear DAG. Every command becomes one shell
	// node; Condition (OnSuccess/OnFailure/Always) decides which of the
	// previous node's edges points here. Legacy flow-level platform/agent
	// are stamped onto every shell node so behavior survives the upgrade.
	base.Nodes = make(map[string]domain.FlowNode, len(l.Commands))
	if len(l.Commands) == 0 {
		return base, nil
	}
	base.Roots = []string{"step_0"}
	for i, c := range l.Commands {
		id := fmt.Sprintf("step_%d", i)
		node := domain.FlowNode{
			Kind:           domain.NodeKindExecutor,
			Executor:       "shell",
			Command:        c.Command,
			Shell:          l.Shell,
			Platform:       l.AgentPlatform,
			Agent:          l.DefaultAgent,
			ExcludedAgents: l.ExcludedAgents,
		}
		if len(c.Command) > 0 && looksLikeJSONParams(c.Command) {
			// Not expected in existing bundled flows, but keeps the door open.
			node.Params = json.RawMessage(c.Command)
			node.Command = ""
		}
		base.Nodes[id] = node
		if i == 0 {
			continue
		}
		prevID := fmt.Sprintf("step_%d", i-1)
		prev := base.Nodes[prevID]
		branch := legacyBranchFromCondition(c.Condition)
		switch branch {
		case domain.EdgeBranchSuccess:
			prev.OnSuccess = append(prev.OnSuccess, id)
		case domain.EdgeBranchError:
			prev.OnError = append(prev.OnError, id)
		default:
			prev.OnSuccess = append(prev.OnSuccess, id)
			prev.OnError = append(prev.OnError, id)
		}
		base.Nodes[prevID] = prev
	}
	return base, nil
}

func legacyBranchFromCondition(cond *string) domain.EdgeBranch {
	if cond == nil {
		return "always"
	}
	switch *cond {
	case "OnSuccess":
		return domain.EdgeBranchSuccess
	case "OnFailure":
		return domain.EdgeBranchError
	default:
		return "always"
	}
}

func looksLikeJSONParams(s string) bool {
	if len(s) < 2 {
		return false
	}
	return s[0] == '{' && s[len(s)-1] == '}'
}

func renameFlowFile(src, dst string) error {
	if src == dst {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func removeFlowFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
