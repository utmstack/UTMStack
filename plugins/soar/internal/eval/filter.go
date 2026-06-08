package eval

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	OperatorIs         = "IS"
	OperatorIsOneOf    = "IS_ONE_OF"
	OperatorIsNotOneOf = "IS_NOT_ONE_OF"
)

const FieldDataSource = "dataSource"

type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func ParseConditions(raw string) ([]Condition, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var conds []Condition
	if err := json.Unmarshal([]byte(raw), &conds); err != nil {
		return nil, err
	}
	return conds, nil
}

func normalizeField(field string) string {
	return strings.Replace(field, ".keyword", "", -1)
}

func gjsonValue(alertJSON, field string) gjson.Result {
	return gjson.Get(alertJSON, normalizeField(field))
}

func equals(actual gjson.Result, expected interface{}) bool {
	if !actual.Exists() {
		return false
	}
	switch v := expected.(type) {
	case string:
		return actual.String() == v
	case float64:
		return actual.Num == v
	case bool:
		return actual.Bool() == v
	default:
		return actual.String() == toString(expected)
	}
}

func toStringList(value interface{}) []string {
	switch v := value.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, toString(item))
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

func contains(needles []string, hay gjson.Result) bool {
	hayStr := hay.String()
	for _, n := range needles {
		if n == hayStr {
			return true
		}
	}
	return false
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), `"`)
	}
}

func Matches(alertJSON string, conditions []Condition) bool {
	for _, c := range conditions {
		actual := gjsonValue(alertJSON, c.Field)
		switch c.Operator {
		case OperatorIs:
			if !equals(actual, c.Value) {
				return false
			}
		case OperatorIsOneOf:
			list := toStringList(c.Value)
			if len(list) == 0 || !contains(list, actual) {
				return false
			}
		case OperatorIsNotOneOf:
			list := toStringList(c.Value)
			if len(list) > 0 && contains(list, actual) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
