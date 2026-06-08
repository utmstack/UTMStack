package usecase

import (
	"errors"
	"time"
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

type Flow struct {
	Name           string          `yaml:"name"`
	Description    string          `yaml:"description,omitempty"`
	Conditions     []FlowCondition `yaml:"conditions"`
	Commands       []string        `yaml:"commands"`
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
