package common_models

// FilterOperator is the vocabulary a filter is written in, shared by alert tag
// rules, the dashboards spec and every module that takes a filter list.
type FilterOperator string

const (
	OpEquals      FilterOperator = "IS"
	OpNotEquals   FilterOperator = "IS_NOT"
	OpIn          FilterOperator = "IS_ONE_OF_TERMS"
	OpInOr        FilterOperator = "IS_ONE_OF_TERMS_OR"
	OpLessOrEq    FilterOperator = "IS_LESS_THAN_OR_EQUALS"
	OpGreater     FilterOperator = "IS_GREATER_THAN"
	OpExists      FilterOperator = "EXIST"
	OpNotExists   FilterOperator = "DOES_NOT_EXIST"
	OpNotContains FilterOperator = "NOT_CONTAINS"

	// Extended operators ported from Java OperatorType enum.
	OpContain             FilterOperator = "CONTAIN"
	OpDoesNotContain      FilterOperator = "DOES_NOT_CONTAIN"
	OpContainOneOf        FilterOperator = "CONTAIN_ONE_OF"
	OpDoesNotContainOneOf FilterOperator = "DOES_NOT_CONTAIN_ONE_OF"
	OpIsOneOf             FilterOperator = "IS_ONE_OF"
	OpIsNotOneOf          FilterOperator = "IS_NOT_ONE_OF"
	OpIsBetween           FilterOperator = "IS_BETWEEN"
	OpIsNotBetween        FilterOperator = "IS_NOT_BETWEEN"
	OpIsInFields          FilterOperator = "IS_IN_FIELDS"
	OpIsNotInFields       FilterOperator = "IS_NOT_IN_FIELDS"
	OpEndsWith            FilterOperator = "ENDS_WITH"
	OpNotEndsWith         FilterOperator = "NOT_ENDS_WITH"
	OpStartWith           FilterOperator = "START_WITH"
	OpNotStartWith        FilterOperator = "NOT_START_WITH"
)

// FilterType is one predicate: a field, what to ask about it, and the value.
type FilterType struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    any            `json:"value,omitempty"`
}
