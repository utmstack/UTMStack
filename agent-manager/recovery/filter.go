package recovery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/utmstack/UTMStack/agent-manager/models"
)

// Resolve returns the subset of agents that match the recovery's target
// criteria. It is a pure function — callers pass the full agent list,
// no DB access happens here.
//
// Matching rules (ALL must hold):
//   - r.TargetOS matches agent.Os (case-insensitive, required)
//   - r.TargetVersionLte is empty OR agent.Version ≤ r.TargetVersionLte (semver)
//
// Field mapping (confirmed from models/agent.go):
//   - OS      → models.Agent.Os       (e.g. "windows", "linux", "darwin")
//   - Version → models.Agent.Version
//
// Architecture filtering is intentionally NOT supported in v1: the agent
// does not report CPU architecture over the wire. See YAMLTarget docs in
// recovery/types.go.
//
// If version comparison fails for either side (unparseable), the agent is
// EXCLUDED as a defensive default — better to skip than to wrongly include.
func Resolve(r *models.Recovery, agents []models.Agent) []models.Agent {
	out := make([]models.Agent, 0, len(agents))
	for _, a := range agents {
		if !matchesOS(r.TargetOS, a.Os) {
			continue
		}
		if r.TargetVersionLte != "" {
			cmp, err := compareVersions(a.Version, r.TargetVersionLte)
			if err != nil {
				// Unparseable version — exclude defensively.
				continue
			}
			if cmp > 0 {
				// agent.Version > TargetVersionLte → not a match.
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func matchesOS(target, agentOS string) bool {
	return strings.EqualFold(strings.TrimSpace(target), strings.TrimSpace(agentOS))
}

// compareVersions returns -1 if a < b, 0 if equal, 1 if a > b.
// Returns error if either version string is unparseable.
//
// Version format: optional leading "v", then "major.minor.patch".
// Pre-release tags (e.g., -rc1, -beta) are stripped before comparison.
func compareVersions(a, b string) (int, error) {
	aMaj, aMin, aPatch, err := parseVersion(a)
	if err != nil {
		return 0, fmt.Errorf("parse version %q: %w", a, err)
	}
	bMaj, bMin, bPatch, err := parseVersion(b)
	if err != nil {
		return 0, fmt.Errorf("parse version %q: %w", b, err)
	}

	switch {
	case aMaj != bMaj:
		return cmpInt(aMaj, bMaj), nil
	case aMin != bMin:
		return cmpInt(aMin, bMin), nil
	default:
		return cmpInt(aPatch, bPatch), nil
	}
}

// parseVersion strips an optional leading "v", strips pre-release tags, and
// splits into major/minor/patch integers.
func parseVersion(v string) (major, minor, patch int, err error) {
	s := strings.TrimPrefix(strings.TrimSpace(v), "v")
	// Strip pre-release suffix (e.g., -rc1, -beta, -alpha.1).
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		s = s[:idx]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("expected major.minor.patch, got %q", v)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("major not an integer in %q: %w", v, err)
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("minor not an integer in %q: %w", v, err)
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("patch not an integer in %q: %w", v, err)
	}
	return major, minor, patch, nil
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
