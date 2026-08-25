package usecase

import (
	"context"
	"encoding/json"
	"testing"
)

func TestInterpolate_AlertLookup(t *testing.T) {
	bag := NewRootContext(json.RawMessage(`{"src_ip":"10.0.0.5","user":{"name":"alice"}}`))
	got, err := Interpolate(context.Background(), nil, bag, "block $(alert.src_ip) for $(alert.user.name)")
	if err != nil {
		t.Fatal(err)
	}
	if want := "block 10.0.0.5 for alice"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolate_UnknownPathStaysLiteral(t *testing.T) {
	bag := NewRootContext(json.RawMessage(`{"src_ip":"10.0.0.5"}`))
	got, err := Interpolate(context.Background(), nil, bag, "$(alert.missing) $(alert.src_ip)")
	if err != nil {
		t.Fatal(err)
	}
	// Missing paths stay verbatim so an operator can spot the typo instead of
	// running a command with a silently blank field.
	if want := "$(alert.missing) 10.0.0.5"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeContexts_EnrichmentOutputPropagates(t *testing.T) {
	alertBag := NewRootContext(json.RawMessage(`{"id":"a-1"}`))
	geoOutput := json.RawMessage(`{"country":"US","asn":15169}`)
	merged := MergeContexts([]ParentContribution{
		{Context: json.RawMessage(alertBag)},
		{EnrichmentNodeID: "geoip", Output: geoOutput, Context: json.RawMessage(alertBag)},
	})
	got, err := Interpolate(context.Background(), nil, merged, "hit from $(geoip.country) alert=$(alert.id)")
	if err != nil {
		t.Fatal(err)
	}
	if want := "hit from US alert=a-1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeContexts_ExecutorParentContributesNothing(t *testing.T) {
	base := NewRootContext(json.RawMessage(`{"id":"a-1"}`))
	// Parent is an executor (no EnrichmentNodeID) — its Output must not surface
	// in the merged bag even if supplied by mistake.
	merged := MergeContexts([]ParentContribution{
		{Context: json.RawMessage(base), Output: json.RawMessage(`{"leaked":true}`)},
	})
	got, err := Interpolate(context.Background(), nil, merged, "$(leaked.something)")
	if err != nil {
		t.Fatal(err)
	}
	if got != "$(leaked.something)" {
		t.Errorf("executor output leaked into context: %q", got)
	}
}
