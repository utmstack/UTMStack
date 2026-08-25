package usecase

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
)

// placeholderRE matches `$(a.b.c)` — a dotted path whose first segment is
// either "alert", a variable-context key, or an enrichment node id.
var placeholderRE = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z0-9_-]+)*)\)`)

// ContextBag is the merged JSON context an execution sees. Top-level keys are
// "alert" plus one entry per ancestor enrichment node.
type ContextBag json.RawMessage

// NewRootContext builds the context every root execution starts with: just the
// frozen alert payload keyed under "alert". Every downstream node inherits and
// augments this bag.
func NewRootContext(alertJSON json.RawMessage) ContextBag {
	if len(alertJSON) == 0 {
		return ContextBag(`{"alert":{}}`)
	}
	// Wrap the alert JSON under an "alert" key. Since alertJSON may be raw and
	// large, we build the wrapper by string concat rather than round-tripping
	// through map[string]any.
	buf := make([]byte, 0, len(alertJSON)+16)
	buf = append(buf, `{"alert":`...)
	buf = append(buf, alertJSON...)
	buf = append(buf, '}')
	return ContextBag(buf)
}

// MergeContexts produces a child's context bag from its resolved parents.
// Each parent contributes its own context, plus (for enrichment parents only)
// its output stored under the parent's node id. Later parents overwrite earlier
// ones on key collision — cycles land the deepest instance last, so retries
// win, matching the "most recent value" intuition.
func MergeContexts(parents []ParentContribution) ContextBag {
	merged := map[string]json.RawMessage{}
	for _, p := range parents {
		if len(p.Context) > 0 {
			var kv map[string]json.RawMessage
			if err := json.Unmarshal(p.Context, &kv); err == nil {
				for k, v := range kv {
					merged[k] = v
				}
			}
		}
		if p.EnrichmentNodeID != "" && len(p.Output) > 0 {
			merged[p.EnrichmentNodeID] = p.Output
		}
	}
	if len(merged) == 0 {
		return ContextBag(`{}`)
	}
	raw, err := json.Marshal(merged)
	if err != nil {
		return ContextBag(`{}`)
	}
	return ContextBag(raw)
}

// ParentContribution is one parent's slice of state that flows into the child's
// context. EnrichmentNodeID is empty when the parent is an executor node.
type ParentContribution struct {
	EnrichmentNodeID string
	Context          json.RawMessage
	Output           json.RawMessage
}

// Interpolate substitutes `$(...)` placeholders in `input` using the merged
// context and then applies `$[variables.NAME]` secrets via the variable
// usecase. The two syntaxes are independent — a template may use either or
// both. Returns the input verbatim when nothing matches.
func Interpolate(ctx context.Context, vars connectors.VariableUsecase, bag ContextBag, input string) (string, error) {
	if input == "" {
		return "", nil
	}
	out := placeholderRE.ReplaceAllStringFunc(input, func(match string) string {
		path := strings.TrimSuffix(strings.TrimPrefix(match, "$("), ")")
		val := gjson.GetBytes(bag, path)
		if !val.Exists() {
			return match
		}
		return val.String()
	})
	if vars == nil {
		return out, nil
	}
	return vars.InterpolateCommand(ctx, out)
}

// InterpolateJSON is a shortcut for interpolating a JSON blob and returning it
// as json.RawMessage. When input is empty, returns nil (nothing to write).
func InterpolateJSON(ctx context.Context, vars connectors.VariableUsecase, bag ContextBag, input json.RawMessage) (json.RawMessage, error) {
	if len(input) == 0 {
		return nil, nil
	}
	s, err := Interpolate(ctx, vars, bag, string(input))
	if err != nil {
		return nil, err
	}
	return json.RawMessage(s), nil
}
