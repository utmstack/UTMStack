package recovery

import (
	"fmt"
	"strings"
)

// validShells is the set of accepted shell values (all lowercase for comparison).
var validShells = map[string]bool{
	"powershell": true,
	"cmd":        true,
	"bash":       true,
	"sh":         true,
}

// ValidateSchema checks that all required fields of a RecoveryYAML are present
// and have valid values. It also applies sane defaults to optional fields that
// have a zero value. Returns nil on success, a descriptive error on failure.
//
// Required: yaml_id, name, shell ∈ {powershell,cmd,bash,sh},
// target.os, success_pattern, script.
// Optional defaults: max_attempts=3, retry_after_seconds=1800 (30m),
// ack_timeout_seconds=300 (5m), max_concurrency=100,
// success_pattern=DefaultSuccessPattern.
func ValidateSchema(y *RecoveryYAML) error {
	if strings.TrimSpace(y.YamlID) == "" {
		return fmt.Errorf("yaml_id is required")
	}
	if strings.TrimSpace(y.Name) == "" {
		return fmt.Errorf("name is required (yaml_id=%q)", y.YamlID)
	}
	if !validShells[strings.ToLower(y.Shell)] {
		return fmt.Errorf("shell %q is invalid: must be one of powershell, cmd, bash, sh (yaml_id=%q)", y.Shell, y.YamlID)
	}
	if strings.TrimSpace(y.Target.OS) == "" {
		return fmt.Errorf("target.os is required (yaml_id=%q)", y.YamlID)
	}
	if strings.TrimSpace(y.Script) == "" {
		return fmt.Errorf("script is required (yaml_id=%q)", y.YamlID)
	}

	// Apply optional defaults.
	if y.SuccessPattern == "" {
		y.SuccessPattern = DefaultSuccessPattern
	}
	if y.MaxAttempts == 0 {
		y.MaxAttempts = 3
	}
	if y.RetryAfter == 0 {
		y.RetryAfter = 1800 // 30 minutes in seconds
	}
	if y.AckTimeout == 0 {
		y.AckTimeout = 300 // 5 minutes in seconds
	}
	if y.MaxConcurrency == 0 {
		y.MaxConcurrency = 100
	}

	return nil
}

// ValidateSentinel ensures that the configured success_pattern appears as a
// literal substring of script. The match is case-sensitive because the script
// is executed verbatim and output comparison will be exact.
func ValidateSentinel(y *RecoveryYAML) error {
	if !strings.Contains(y.Script, y.SuccessPattern) {
		return fmt.Errorf(
			"success_pattern %q not found as literal substring of script (yaml_id=%q): script must print the sentinel to stdout",
			y.SuccessPattern, y.YamlID,
		)
	}
	return nil
}

// DetectCycles runs a DFS over the depends_on graph formed by the provided
// valid YAMLs and returns a map of yaml_id → cycle path string (e.g. "A->B->A")
// for every recovery that participates in a cycle.
//
// Recoveries NOT in any cycle are absent from the returned map.
// depends_on references pointing to yaml_ids outside the valid set are
// ignored — that is "dependency_missing", a separate concern handled in service.go.
//
// Edge cases handled:
//   - Empty input → empty map
//   - Self-loop (A.depends_on = A) → {"A": "A->A"}
//   - Mutual cycle (A↔B) → both included
//   - 3+ node cycle (A→B→C→A) → all three included
//   - depends_on pointing to unknown yaml_id → NOT a cycle
//   - Mixed graph (one cyclic component + one acyclic) → only cyclic component returned
func DetectCycles(valid []RecoveryYAML) map[string]string {
	// Build adjacency: yaml_id → depends_on yaml_id (only if target is in valid set).
	validSet := make(map[string]struct{}, len(valid))
	for _, y := range valid {
		validSet[y.YamlID] = struct{}{}
	}

	dep := make(map[string]string, len(valid)) // yaml_id → depends_on
	for _, y := range valid {
		if y.DependsOn != "" {
			if _, ok := validSet[y.DependsOn]; ok {
				dep[y.YamlID] = y.DependsOn
			}
		}
	}

	// DFS state per node.
	const (
		unvisited = 0
		inStack   = 1 // currently on the DFS path
		done      = 2
	)
	state := make(map[string]int, len(valid))
	result := make(map[string]string)

	// path is the ordered list of nodes on the current DFS stack.
	// pathIdx maps node → its index in path for O(1) lookup.
	path := make([]string, 0, len(valid))
	pathIdx := make(map[string]int, len(valid))

	// visit performs recursive DFS. Using a local recursive closure is safe
	// for the expected input size (< 100 recoveries); the stack depth equals
	// the longest simple path in the graph, which is bounded by len(valid).
	var visit func(node string)
	visit = func(node string) {
		if state[node] == done {
			return
		}
		if state[node] == inStack {
			// Back edge detected — we found a cycle.
			// Collect the cycle starting from the first occurrence of node in path.
			start := pathIdx[node]
			cycle := make([]string, 0, len(path)-start+1)
			cycle = append(cycle, path[start:]...)
			cycle = append(cycle, node) // close the cycle
			cycleStr := strings.Join(cycle, "->")
			for _, n := range path[start:] {
				result[n] = cycleStr
			}
			return
		}

		// Mark as in-stack and push to path.
		state[node] = inStack
		pathIdx[node] = len(path)
		path = append(path, node)

		if next, ok := dep[node]; ok {
			visit(next)
		}

		// Pop from path and mark done.
		path = path[:len(path)-1]
		delete(pathIdx, node)
		state[node] = done
	}

	for _, y := range valid {
		visit(y.YamlID)
	}

	return result
}
