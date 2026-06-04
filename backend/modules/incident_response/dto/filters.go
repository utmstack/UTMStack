package dto

type ActionFilter struct {
	Page           int     `form:"page"`
	Size           int     `form:"size"`
	ActionCommand  *string `form:"actionCommand"`
	ActionType     *int    `form:"actionType"`
	ActionEditable *bool   `form:"actionEditable"`
}

type ActionCommandFilter struct {
	Page       int     `form:"page"`
	Size       int     `form:"size"`
	ActionID   *int64  `form:"actionId"`
	OsPlatform *string `form:"osPlatform"`
	Command    *string `form:"command"`
}

type JobFilter struct {
	Page       int     `form:"page"`
	Size       int     `form:"size"`
	ActionID   *int64  `form:"actionId"`
	Agent      *string `form:"agent"`
	Status     *int    `form:"status"`
	OriginID   *int    `form:"originId"`
	OriginType *string `form:"originType"`
}

type VariableFilter struct {
	Page         int     `form:"page"`
	Size         int     `form:"size"`
	VariableName *string `form:"variableName"`
}
