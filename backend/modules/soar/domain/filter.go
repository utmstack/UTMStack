package domain

type OperatorType string

const (
	OperatorIS         OperatorType = "IS"
	OperatorIsOneOf    OperatorType = "IS_ONE_OF"
	OperatorIsNotOneOf OperatorType = "IS_NOT_ONE_OF"
)

type FilterType struct {
	Operator OperatorType `json:"operator"`
	Field    string       `json:"field"`
	Value    any          `json:"value"` // string for IS; []string for IS_ONE_OF / IS_NOT_ONE_OF
}
