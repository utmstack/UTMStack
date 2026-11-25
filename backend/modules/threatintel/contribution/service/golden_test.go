package service

import (
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/association"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/extractor"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/mapper"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
)

type sample struct {
	Incident *models.Incident `json:"incident"`
	Alerts   []*models.Alert  `json:"alerts"`
}

func loadSample(t *testing.T) sample {
	t.Helper()

	raw, err := os.ReadFile("../testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}

	var s sample
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}

	return s
}

// entitiesOf runs the extraction, mapping and association steps over the sample, as the incident
// processor does for each alert, and returns the entities as sorted JSON.
func entitiesOf(s sample) []string {
	assoc := association.NewAssociationBuilder()
	builder := NewEntityBuilder(mapper.NewEntityMapper())
	fx := extractor.NewFieldExtractor()

	for _, a := range s.Alerts {
		for _, e := range a.Events {
			builder.MapAndRegisterFieldsToEntities(fx.ExtractFromEvent(e), s.Incident, a, e, assoc)
		}
	}

	var out []string

	for _, e := range assoc.BuildAssociations() {
		text, err := models.CanonicalJSON(e)
		if err != nil {
			panic(err)
		}

		out = append(out, text)
	}

	sort.Strings(out)

	return out
}

// TestEntitiesMatchWhatThePluginProduced guards the port. golden_entities.json holds what the plugin
// this code came from produced for the same sample, so any difference means the behaviour changed.
func TestEntitiesMatchWhatThePluginProduced(t *testing.T) {
	raw, err := os.ReadFile("../testdata/golden_entities.json")
	if err != nil {
		t.Fatal(err)
	}

	var recorded []string
	if err := json.Unmarshal(raw, &recorded); err != nil {
		t.Fatal(err)
	}

	// The reference was recorded as the plugin printed it, with the associations in whatever order that
	// run produced; both sides are compared in canonical form.
	want := make([]string, 0, len(recorded))

	for _, r := range recorded {
		var generic any
		if err := json.Unmarshal([]byte(r), &generic); err != nil {
			t.Fatal(err)
		}

		text, err := models.CanonicalJSON(generic)
		if err != nil {
			t.Fatal(err)
		}

		want = append(want, text)
	}

	sort.Strings(want)

	got := entitiesOf(loadSample(t))

	if len(got) != len(want) {
		t.Fatalf("got %d entities, the plugin produced %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entity %d differs:\n got:  %s\n want: %s", i, got[i], want[i])
		}
	}
}

func TestTheSameInputGivesTheSameEntitiesEveryTime(t *testing.T) {
	s := loadSample(t)

	first := entitiesOf(s)

	for i := 0; i < 5; i++ {
		again := entitiesOf(s)
		if len(again) != len(first) {
			t.Fatalf("run %d: %d entities, first run had %d", i, len(again), len(first))
		}

		for j := range first {
			if again[j] != first[j] {
				t.Fatalf("run %d: entity %d changed between runs; the registry of sent entities depends on this", i, j)
			}
		}
	}
}

func TestNoEntitiesFromAnAlertWithoutEvents(t *testing.T) {
	s := loadSample(t)
	s.Alerts = []*models.Alert{{ID: "empty", Name: "n", DataType: "d"}}

	if got := entitiesOf(s); len(got) != 0 {
		t.Errorf("got %d entities from an alert with no events", len(got))
	}
}
