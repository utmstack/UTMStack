package domain

type LogRef struct {
	ID    string
	Index string
}

type CorrelationExpr struct {
	Field    string
	Operator string // filter_term | must_not_term | filter_match | must_not_match
	Value    any
}

type CorrelationStep struct {
	IndexPattern string
	With         []CorrelationExpr
	Or           []CorrelationStep
	Within       string // duration string (e.g. "2m"); empty = no time bound
	Count        uint64 // threshold the rule needed to fire (mirrors SearchRequest.Count)
}
