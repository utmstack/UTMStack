package domain

type OperatorType string

const (
	OperatorIS    OperatorType = "IS"
	OperatorISNot OperatorType = "IS_NOT"

	OperatorContains    OperatorType = "CONTAINS"
	OperatorNotContains OperatorType = "NOT_CONTAINS"

	OperatorExists    OperatorType = "EXISTS"
	OperatorNotExists OperatorType = "NOT_EXISTS"

	OperatorStartWith    OperatorType = "START_WITH"
	OperatorNotStartWith OperatorType = "NOT_START_WITH"

	OperatorEndsWith    OperatorType = "ENDS_WITH"
	OperatorNotEndsWith OperatorType = "NOT_ENDS_WITH"

	OperatorIsOneOf    OperatorType = "IS_ONE_OF"
	OperatorIsNotOneOf OperatorType = "IS_NOT_ONE_OF"
)

type FilterType struct {
	Operator OperatorType `json:"operator"`
	Field    string       `json:"field"`
	Value    any          `json:"value"` // string for IS; []string for IS_ONE_OF / IS_NOT_ONE_OF
}
