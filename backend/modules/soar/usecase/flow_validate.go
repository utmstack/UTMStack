package usecase

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// knownExecutors mirrors module.go's registry; keep in sync when adding one.
var knownExecutors = map[string]bool{
	"shell":       true,
	"http":        true,
	"conditional": true,
	"llm_enrich":  true,
	"llm_action":  true,
	"notify":      true,
	"incident":    true,
	"mail":        true,
}

// validateFlow rejects flows that could never run correctly. Called from
// writeFlowFile so every API create/update goes through it; system seeds copy
// raw files and skip it, so old shipped YAML never breaks startup over new rules.
// Errors accumulate: one request surfaces every problem instead of forcing
// per-fixup round-trips.
func validateFlow(f domain.Flow) error {
	var errs []string

	if strings.TrimSpace(f.Name) == "" {
		errs = append(errs, "name is required")
	}

	// Triggers: flow never dispatches without at least one condition.
	for i, c := range f.Conditions {
		switch {
		case strings.TrimSpace(c.Field) == "":
			errs = append(errs, fmt.Sprintf("conditions[%d]: field is required", i))
		case strings.TrimSpace(string(c.Operator)) == "":
			errs = append(errs, fmt.Sprintf("conditions[%d]: operator is required", i))
		}
	}
	if len(f.Conditions) == 0 {
		errs = append(errs, "at least one trigger condition is required")
	}

	// Roots must exist, be non-empty and be unique (duplicates double-count
	// AND-join sizing downstream).
	seenRoots := make(map[string]bool, len(f.Roots))
	for _, id := range f.Roots {
		if strings.TrimSpace(id) == "" {
			errs = append(errs, "roots contains an empty id")
			continue
		}
		if seenRoots[id] {
			errs = append(errs, fmt.Sprintf("root %q is listed twice", id))
			continue
		}
		seenRoots[id] = true
		if _, ok := f.Nodes[id]; !ok {
			errs = append(errs, fmt.Sprintf("root %q does not match any node", id))
		}
	}

	for id, n := range f.Nodes {
		if strings.TrimSpace(id) == "" {
			errs = append(errs, "nodes contains an empty id")
			continue
		}
		if !knownExecutors[n.Executor] {
			errs = append(errs, fmt.Sprintf("node %q: unknown executor %q", id, n.Executor))
			continue
		}
		// Edges must point at real nodes.
		for _, t := range n.OnSuccess {
			if t == id {
				errs = append(errs, fmt.Sprintf("node %q: onSuccess points at itself", id))
			} else if _, ok := f.Nodes[t]; !ok {
				errs = append(errs, fmt.Sprintf("node %q: onSuccess target %q does not exist", id, t))
			}
		}
		for _, t := range n.OnError {
			if t == id {
				errs = append(errs, fmt.Sprintf("node %q: onError points at itself", id))
			} else if _, ok := f.Nodes[t]; !ok {
				errs = append(errs, fmt.Sprintf("node %q: onError target %q does not exist", id, t))
			}
		}
		errs = append(errs, validateNodeParams(id, n)...)
	}

	if cyc := findCycle(f); cyc != nil {
		errs = append(errs, "flow contains a cycle: "+strings.Join(cyc, " -> "))
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid flow: %s", strings.Join(errs, "; "))
	}
	return nil
}

// validateNodeParams checks per-executor required fields. http urls are checked
// after stripping $(...) placeholders: interpolation happens at dispatch time,
// so a literal parse would reject every templated endpoint.
func validateNodeParams(id string, n domain.FlowNode) []string {
	var errs []string

	switch n.Executor {
	case "shell":
		if strings.TrimSpace(n.Command) == "" && len(n.Params) == 0 {
			errs = append(errs, fmt.Sprintf("node %q: shell requires a command", id))
		}
	case "http":
		var p struct {
			URL string `json:"url"`
		}
		if len(n.Params) > 0 {
			if err := json.Unmarshal(n.Params, &p); err != nil {
				errs = append(errs, fmt.Sprintf("node %q: params is not valid JSON: %v", id, err))
				return errs
			}
		}
		if u := strings.TrimSpace(p.URL); u == "" {
			errs = append(errs, fmt.Sprintf("node %q: http requires params.url", id))
		} else if err := validFlowURL(u); err != nil {
			errs = append(errs, fmt.Sprintf("node %q: http url invalid: %v", id, err))
		}
	case "conditional":
		var p struct {
			Conditions []domain.FilterType `json:"conditions"`
		}
		if len(n.Params) > 0 {
			if err := json.Unmarshal(n.Params, &p); err != nil {
				errs = append(errs, fmt.Sprintf("node %q: params is not valid JSON: %v", id, err))
				return errs
			}
		}
		if len(p.Conditions) == 0 {
			errs = append(errs, fmt.Sprintf("node %q: conditional requires params.conditions", id))
		}
	case "notify":
		var p struct {
			Message string `json:"message"`
		}
		if len(n.Params) > 0 {
			_ = json.Unmarshal(n.Params, &p)
		}
		if strings.TrimSpace(p.Message) == "" {
			errs = append(errs, fmt.Sprintf("node %q: notify requires params.message", id))
		}
	case "mail":
		var p struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}
		if len(n.Params) > 0 {
			_ = json.Unmarshal(n.Params, &p)
		}
		miss := make([]string, 0, 3)
		if strings.TrimSpace(p.To) == "" {
			miss = append(miss, "to")
		}
		if strings.TrimSpace(p.Subject) == "" {
			miss = append(miss, "subject")
		}
		if strings.TrimSpace(p.Body) == "" {
			miss = append(miss, "body")
		}
		if len(miss) > 0 {
			errs = append(errs, fmt.Sprintf("node %q: mail requires params: %s", id, strings.Join(miss, ", ")))
		}
	case "incident":
		var p struct {
			Name string `json:"name"`
		}
		if len(n.Params) > 0 {
			_ = json.Unmarshal(n.Params, &p)
		}
		if strings.TrimSpace(p.Name) == "" {
			errs = append(errs, fmt.Sprintf("node %q: incident requires params.name", id))
		}
	case "llm_enrich", "llm_action":
		var p struct {
			Prompt string `json:"prompt"`
		}
		if len(n.Params) > 0 {
			_ = json.Unmarshal(n.Params, &p)
		}
		if strings.TrimSpace(p.Prompt) == "" {
			errs = append(errs, fmt.Sprintf("node %q: %s requires params.prompt", id, n.Executor))
		}
	}
	return errs
}

// validFlowURL accepts absolute http(s) urls; $(...) placeholders are blanked
// first because they resolve at dispatch time.
func validFlowURL(u string) error {
	cleaned := placeholderRE.ReplaceAllString(u, "")
	cleaned = strings.TrimSpace(cleaned)
	parsed, err := url.ParseRequestURI(cleaned)
	if err != nil {
		return fmt.Errorf("cannot parse %q", u)
	}
	if !parsed.IsAbs() {
		return fmt.Errorf("url %q must be absolute", u)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url scheme %q must be http or https", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("url %q has no host", u)
	}
	return nil
}

// findCycle returns one cycle path (ids) if the flow has any, else nil.
// Iterative three-color DFS with an explicit parent map so recursion depth
// never tracks user-controlled flow size, and the cycle is rebuilt in O(length).
func findCycle(f domain.Flow) []string {
	const (
		white = 0 // unvisited
		gray  = 1 // on current DFS path
		black = 2 // fully explored
	)
	color := make(map[string]int, len(f.Nodes))
	parent := make(map[string]string, len(f.Nodes))

	for start := range f.Nodes {
		if color[start] != white {
			continue
		}
		color[start] = gray
		stack := []string{start}
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			expanded := false
			for _, t := range append(append([]string{}, f.Nodes[id].OnSuccess...), f.Nodes[id].OnError...) {
				if _, ok := f.Nodes[t]; !ok {
					continue // dangling edges reported by validateFlow
				}
				switch color[t] {
				case white:
					color[t] = gray
					parent[t] = id
					stack = append(stack, t)
					expanded = true
				case gray:
					// id -> t closes the cycle; walk parents from id back to t.
					c := []string{t}
					for cur := id; cur != t; cur = parent[cur] {
						c = append(c, cur)
					}
					c = append(c, t)
					reverse(c)
					return c
				}
			}
			if !expanded {
				stack = stack[:len(stack)-1]
				color[id] = black
			}
		}
	}
	return nil
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
