package usecase

import (
	"errors"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

var (
	ErrFlowNotFound      = errors.New("flow not found")
	ErrSystemFlowContent = errors.New("system flow content is read-only")
)

type FlowCondition struct {
	Operator string `yaml:"operator"`
	Field    string `yaml:"field"`
	Value    any    `yaml:"value"`
}

// FlowCommand is one step in a flow's command chain. Condition (nil on the
// first entry) is how this command joins to the previous one when the chain
// is concatenated into a single shell line.
type FlowCommand struct {
	Command   string            `yaml:"command"`
	Condition *domain.Condition `yaml:"condition,omitempty"`
}

// UnmarshalYAML accepts either a bare string (legacy `commands: [cmd, cmd]`)
// or a mapping `{command, condition}`. Bare strings decode with a nil
// Condition; the loader treats that as Always when it's not the first step.
func (fc *FlowCommand) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		fc.Command = value.Value
		return nil
	}
	type raw FlowCommand
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*fc = FlowCommand(r)
	return nil
}

type Flow struct {
	Name           string          `yaml:"name"`
	Description    string          `yaml:"description,omitempty"`
	Conditions     []FlowCondition `yaml:"conditions"`
	Commands       []FlowCommand   `yaml:"commands"`
	Shell          string          `yaml:"shell,omitempty"`
	AgentPlatform  string          `yaml:"agentPlatform,omitempty"`
	DefaultAgent   string          `yaml:"defaultAgent,omitempty"`
	ExcludedAgents []string        `yaml:"excludedAgents,omitempty"`
}

type StoredFlow struct {
	Flow

	RelPath  string
	Modified time.Time

	system  bool
	enabled bool
}

func (sf StoredFlow) Active() bool { return sf.enabled }

func (sf StoredFlow) SystemOwned() bool { return sf.system }
