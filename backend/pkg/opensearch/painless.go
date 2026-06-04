package opensearch

// Script holds a Painless script source and its parameters map.
// Always populate Params — never embed user input directly in Source.
type Script struct {
	// Source is the Painless script body (a constant string in the caller).
	Source string
	// Params maps parameter names to values injected at runtime.
	Params map[string]any
}

// Render produces the JSON-ready map expected by OpenSearch's update_by_query
// script block:
//
//	{
//	  "source": "...",
//	  "lang":   "painless",
//	  "params": { ... }
//	}
func (s Script) Render() map[string]any {
	p := s.Params
	if p == nil {
		p = map[string]any{}
	}
	return map[string]any{
		"source": s.Source,
		"lang":   "painless",
		"params": p,
	}
}
