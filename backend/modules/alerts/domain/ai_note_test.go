package domain

import "testing"

const assessment = "[AI SOC Agent] Score: 80/100 - Open - High Risk | Threat Assessment: x | Action: isolate"

func TestMergeAssessmentKeepsWhatTheAnalystWrote(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		want     string
	}{
		{"nothing written yet", "", assessment},
		{"analyst text only", "Checked with the owner, expected.", "Checked with the owner, expected.\n\n" + assessment},
		{"replaces an earlier assessment, keeps the analyst", "Checked with the owner.\n\n[AI SOC Agent] Score: 10/100 - old", "Checked with the owner.\n\n" + assessment},
		{"earlier assessment only", "[AI SOC Agent] Score: 10/100 - old", assessment},
		{"quoted by the pipeline", `"Analyst note.\n\n[AI SOC Agent] old"`, "Analyst note.\n\n" + assessment},
		{"marker in another case", "Mine.\n\n[ai soc agent] old", "Mine.\n\n" + assessment},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeAssessment(tt.existing, assessment); got != tt.want {
				t.Fatalf("MergeAssessment(%q)\n got: %q\nwant: %q", tt.existing, got, tt.want)
			}
		})
	}
}

func TestIsAIAssessment(t *testing.T) {
	tests := map[string]bool{
		assessment:                       true,
		"  " + assessment:                true,
		`"` + assessment + `"`:           true,
		"":                               false,
		"An analyst wrote this":          false,
		"Mine.\n\n" + assessment:         false,
		"Mentions [AI SOC Agent] midway": false,
	}
	for notes, want := range tests {
		if got := IsAIAssessment(notes); got != want {
			t.Errorf("IsAIAssessment(%q) = %v, want %v", notes, got, want)
		}
	}
}
