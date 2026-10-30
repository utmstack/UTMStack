package domain

import "testing"

func TestValidIdentityName(t *testing.T) {
	valid := []string{"  ops  ", "ops-team", "ab"}
	invalid := []string{"", "   ", "a ", "a", "b"}
	for _, name := range valid {
		if !ValidIdentityName(name) {
			t.Errorf("ValidIdentityName(%q) = false, want true", name)
		}
	}
	for _, name := range invalid {
		if ValidIdentityName(name) {
			t.Errorf("ValidIdentityName(%q) = true, want false", name)
		}
	}
	if ValidIdentityName(string(make([]rune, 65))) {
		t.Error("a 65-char name is too long")
	}
}
