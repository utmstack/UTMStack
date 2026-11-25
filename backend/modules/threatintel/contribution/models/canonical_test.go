package models

import "testing"

func TestCanonicalJSONIgnoresTheOrderOfListsAndKeys(t *testing.T) {
	a := map[string]any{"type": "domain", "associations": []any{map[string]any{"x": 1}, map[string]any{"x": 2}}, "tags": []any{"b", "a"}}
	b := map[string]any{"tags": []any{"a", "b"}, "associations": []any{map[string]any{"x": 2}, map[string]any{"x": 1}}, "type": "domain"}

	ca, err := CanonicalJSON(a)
	if err != nil {
		t.Fatal(err)
	}

	cb, err := CanonicalJSON(b)
	if err != nil {
		t.Fatal(err)
	}

	if ca != cb {
		t.Errorf("same content, different canonical form:\n%s\n%s", ca, cb)
	}
}

func TestCanonicalJSONTellsDifferentContentApart(t *testing.T) {
	a, _ := CanonicalJSON(map[string]any{"associations": []any{"x", "y"}})
	b, _ := CanonicalJSON(map[string]any{"associations": []any{"x", "z"}})
	c, _ := CanonicalJSON(map[string]any{"associations": []any{"x"}})

	if a == b || a == c {
		t.Error("a changed or removed association must give another form")
	}
}
