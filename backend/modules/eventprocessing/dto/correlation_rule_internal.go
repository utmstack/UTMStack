package dto

type InternalDeactivateRuleRequest struct {
	RuleName string `json:"ruleName" binding:"required"`
}

type InternalDeactivateRuleResponse struct {
	Changed bool `json:"changed"`
}
