package domain

import (
	"encoding/json"
	"regexp"
	"strings"
)

// AIAssessmentMarker opens the note the SOC-AI agent writes; the UI reads it
// back by this marker (frontend/src/features/alerts/lib/ai-note.ts).
const AIAssessmentMarker = "[AI SOC Agent]"

var aiMarker = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(AIAssessmentMarker))

// unquoteNotes undoes the JSON quoting some writers left around a stored note.
func unquoteNotes(s string) string {
	v := strings.TrimSpace(s)
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		var parsed string
		if json.Unmarshal([]byte(v), &parsed) == nil {
			return parsed
		}
		return v[1 : len(v)-1]
	}
	return s
}

// IsAIAssessment reports whether notes is an assessment on its own: text that
// opens with the agent's marker and carries nothing of an analyst's before it.
func IsAIAssessment(notes string) bool {
	loc := aiMarker.FindStringIndex(strings.TrimSpace(unquoteNotes(notes)))
	return loc != nil && loc[0] == 0
}

// MergeAssessment writes an assessment beside what the analyst wrote. Notes are
// one string holding the analyst's text followed by the assessment block, so a
// plain replace — what the notes endpoint does — would erase the analyst's
// text whenever the agent records its assessment.
func MergeAssessment(existing, assessment string) string {
	analyst := unquoteNotes(existing)
	if loc := aiMarker.FindStringIndex(analyst); loc != nil {
		analyst = analyst[:loc[0]]
	}
	analyst = strings.TrimSpace(analyst)
	block := strings.TrimSpace(unquoteNotes(assessment))
	if analyst == "" {
		return block
	}
	return analyst + "\n\n" + block
}
