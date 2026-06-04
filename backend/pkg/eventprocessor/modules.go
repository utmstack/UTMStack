package eventprocessor

import (
	"context"
	"encoding/json"
	"net/url"
)

type ModulePayload struct {
	ID                int64         `json:"id"`
	ServerID          int64         `json:"serverId"`
	PrettyName        string        `json:"prettyName,omitempty"`
	ModuleName        string        `json:"moduleName"`
	ModuleDescription string        `json:"moduleDescription,omitempty"`
	ModuleActive      bool          `json:"moduleActive"`
	ModuleIcon        string        `json:"moduleIcon,omitempty"`
	ModuleCategory    string        `json:"moduleCategory,omitempty"`
	LiteVersion       bool          `json:"liteVersion,omitempty"`
	NeedsRestart      bool          `json:"needsRestart,omitempty"`
	IsActivatable     bool          `json:"isActivatable,omitempty"`
	ModuleGroups      []ModuleGroup `json:"moduleGroups"`
}

type ModuleGroup struct {
	ID                        int64               `json:"id"`
	ModuleID                  int64               `json:"moduleId"`
	GroupName                 string              `json:"groupName"`
	GroupDescription          string              `json:"groupDescription,omitempty"`
	Collector                 *string             `json:"collector,omitempty"`
	ModuleGroupConfigurations []ModuleGroupConfig `json:"moduleGroupConfigurations"`
}

type ModuleGroupConfig struct {
	ID              int64  `json:"id"`
	GroupID         int64  `json:"groupId"`
	ConfKey         string `json:"confKey"`
	ConfValue       string `json:"confValue"`
	ConfName        string `json:"confName,omitempty"`
	ConfDescription string `json:"confDescription,omitempty"`
	ConfDataType    string `json:"confDataType"`
	ConfRequired    bool   `json:"confRequired"`
	ConfOptions     string `json:"confOptions,omitempty"`
	ConfVisibility  string `json:"confVisibility,omitempty"`
}

type ValidationError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Field   string `json:"field,omitempty"`
}

type ValidationResult struct {
	Errors []ValidationError `json:"errors,omitempty"`
}

// UpdateModule pushes a module-config change to event-processor-manager. The
// legacy controller wrapped the payload in a single-element array; we keep that
// shape verbatim so the receiving service does not need to be touched.
func (c *Client) UpdateModule(ctx context.Context, nameShort string, module ModulePayload) error {
	q := url.Values{}
	q.Set("nameShort", nameShort)
	_, _, err := c.post(ctx, "/api/v1/modules-config", q, []ModulePayload{module})
	return err
}

// ValidateModule asks event-processor-manager to pre-validate a module config
// before it is persisted. A nil error with empty Errors means OK.
func (c *Client) ValidateModule(ctx context.Context, nameShort string, module ModulePayload) (*ValidationResult, error) {
	q := url.Values{}
	q.Set("nameShort", nameShort)
	_, body, err := c.post(ctx, "/api/v1/modules-config/validate", q, module)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return &ValidationResult{}, nil
	}
	var result ValidationResult
	if jerr := json.Unmarshal(body, &result); jerr != nil {
		return &ValidationResult{}, nil
	}
	return &result, nil
}
