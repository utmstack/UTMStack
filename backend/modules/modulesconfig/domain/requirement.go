package domain

type ModuleRequirementStatus string

const (
	RequirementOK   ModuleRequirementStatus = "OK"
	RequirementFail ModuleRequirementStatus = "FAIL"
)

type ModuleRequirement struct {
	CheckName   string                  `json:"checkName"`
	CheckStatus ModuleRequirementStatus `json:"checkStatus"`
	FailMessage string                  `json:"failMessage,omitempty"`
}
