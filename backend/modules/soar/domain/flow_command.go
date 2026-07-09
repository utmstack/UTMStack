package domain

// Condition names how a flow command joins to the PREVIOUS one when the
// command chain is assembled into a single shell line. The first command in a
// chain carries no Condition (nil).
type Condition string

const (
	ConditionOnSuccess Condition = "OnSuccess" // &&
	ConditionOnFailure Condition = "OnFailure" // ||
	ConditionAlways    Condition = "Always"    // ;
)

// Operator returns the shell operator for this condition. Unknown values fall
// back to ";" so a malformed flow still runs sequentially rather than erroring.
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
