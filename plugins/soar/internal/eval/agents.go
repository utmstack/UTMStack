package eval

import (
	"github.com/tidwall/gjson"

	"github.com/utmstack/UTMStack/plugins/soar/internal/cache"
)

func BuildRuleConditions(rule cache.CachedRule, isAgent bool) ([]Condition, error) {
	base, err := ParseConditions(string(rule.Conditions))
	if err != nil {
		return nil, err
	}

	if len(rule.ExcludedAgents) > 0 {
		base = append(base, Condition{
			Field:    FieldDataSource,
			Operator: OperatorIsNotOneOf,
			Value:    toInterfaceSlice(rule.ExcludedAgents),
		})
	}

	if isAgent {
		base = append(base, Condition{
			Field:    FieldDataSource,
			Operator: OperatorIsOneOf,
			Value:    toInterfaceSlice(rule.AgentNames),
		})
	} else {
		base = append(base, Condition{
			Field:    FieldDataSource,
			Operator: OperatorIsNotOneOf,
			Value:    toInterfaceSlice(rule.AgentNames),
		})
	}
	return base, nil
}

func ResolveAgent(alertJSON string, rule cache.CachedRule, isAgent bool) string {
	if isAgent {
		return gjson.Get(alertJSON, FieldDataSource).String()
	}
	return rule.DefaultAgent
}

func toInterfaceSlice(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
