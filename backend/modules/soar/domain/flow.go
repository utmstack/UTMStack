package domain

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Flow struct {
	Name           string        `yaml:"name"           json:"name"`
	Description    string        `yaml:"description,omitempty" json:"description,omitempty"`
	Conditions     []FilterType  `yaml:"conditions"     json:"conditions"`
	Commands       []FlowCommand `yaml:"commands"      json:"commands"`
	Shell          string        `yaml:"shell,omitempty" json:"shell,omitempty"`
	AgentPlatform  string        `yaml:"agentPlatform,omitempty" json:"agentPlatform,omitempty"`
	DefaultAgent   string        `yaml:"defaultAgent,omitempty" json:"defaultAgent,omitempty"`
	ExcludedAgents []string      `yaml:"excludedAgents,omitempty" json:"excludedAgents,omitempty"`
}

type Condition string

const (
	ConditionOnSuccess Condition = "OnSuccess" // &&
	ConditionOnFailure Condition = "OnFailure" // ||
	ConditionAlways    Condition = "Always"    // ;
)

func (c Condition) Operator() string {
	switch c {
	case ConditionOnSuccess:
		return "&&"
	case ConditionOnFailure:
		return "||"
	default:
		return ";"
	}
}

type FlowCommand struct {
	Command   string     `yaml:"command"             json:"command"`
	Condition *Condition `yaml:"condition,omitempty" json:"condition,omitempty"`
}

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
