package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// Select is a lightweight enrichment: it composes an output object by pulling
// gjson paths out of the current context bag. Handy when downstream nodes want
// a subset (or a renamed subset) of ancestor data without dragging in a jq
// dependency. Meant for kind=enrichment.
// ponytail: gjson is already vendored (used in variable + execution
// interpolation); output built via encoding/json — no new dep.
type Select struct{}

func NewSelect() *Select { return &Select{} }

func (Select) Type() string { return "select" }

type selectParams struct {
	Fields map[string]string `json:"fields"`
}

func (s *Select) Execute(_ context.Context, exec *domain.SoarExecution) (json.RawMessage, error) {
	if exec.Kind != domain.NodeKindEnrichment {
		return nil, errors.New("soar select: kind must be enrichment")
	}
	var p selectParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar select: params: %w", err)
		}
	}
	if len(p.Fields) == 0 {
		return json.RawMessage(`{}`), nil
	}
	src := string(exec.Context)
	if src == "" {
		src = "{}"
	}
	out := make(map[string]json.RawMessage, len(p.Fields))
	for name, path := range p.Fields {
		val := gjson.Get(src, path)
		if !val.Exists() {
			continue
		}
		out[name] = json.RawMessage(val.Raw)
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("soar select: marshal: %w", err)
	}
	exec.Result = string(raw)
	return raw, nil
}
