package mcp

import (
	"strings"
	"testing"
)

var logFields = []string{"@timestamp", "dataType", "dataSource", "id", "origin.user", "origin.ip", "target.user", "target.ip", "log", "raw"}

// The agent guessed "user" for the logs dataset and the store answered with a column error. The guess is
// refused up front, naming the fields it probably meant.
func TestAGuessedFieldIsRefusedWithTheRealOnes(t *testing.T) {
	msg := unknownRootMessage("logs", "user", logFields)

	for _, want := range []string{`"user" is not a field of the logs dataset`, `"origin.user"`, `"target.user"`, "store.dataset.fields"} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in %q", want, msg)
		}
	}
}

func TestRealAndDynamicFieldsPass(t *testing.T) {
	for _, f := range []string{"origin.user", "dataType", "log.eventCode", "log", "origin.anything.deep"} {
		if msg := unknownRootMessage("logs", f, logFields); msg != "" {
			t.Errorf("%s must pass: %s", f, msg)
		}
	}
}

func TestSuggestionsIgnoreCaseAndSeparators(t *testing.T) {
	if got := suggestFields("data_source", logFields); len(got) != 1 || got[0] != `"dataSource"` {
		t.Fatalf("%v", got)
	}

	if got := suggestFields("nonsense", logFields); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}
