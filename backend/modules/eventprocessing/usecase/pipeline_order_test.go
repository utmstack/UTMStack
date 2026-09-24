package usecase

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

func pipelinesFileOrdered(names ...string) []dto.PipelineResponse {
	out := make([]dto.PipelineResponse, len(names))
	for i, n := range names {
		out[i] = dto.PipelineResponse{RelPath: n + ".yaml", Order: int32(i)}
	}
	return out
}

func namesOf(items []dto.PipelineResponse) []string {
	out := make([]string, len(items))
	for i, p := range items {
		out[i] = pipelineIdentity(p.RelPath)
	}
	return out
}

// The saved sequence is written to tenants.yaml and read by the engine, so the
// listing has to show that sequence. It showed the files' own order instead: a
// pipeline moved on the page was back where it started after a reload.
func TestTheSavedOrderComesFirstAndTheRestKeepTheirFilesOrder(t *testing.T) {
	items := pipelinesFileOrdered("aws", "azure", "linux", "macos", "windows")

	sequencePipelines(items, []string{"linux", "aws"})

	want := []string{"linux", "aws", "azure", "macos", "windows"}
	if got := namesOf(items); !reflect.DeepEqual(got, want) {
		t.Errorf("sequence = %v, want %v", got, want)
	}
	for i, p := range items {
		if int(p.Position) != i {
			t.Errorf("%s position = %d, want %d", p.RelPath, p.Position, i)
		}
	}
}

func TestNothingSavedLeavesTheFilesOrder(t *testing.T) {
	items := pipelinesFileOrdered("aws", "azure", "linux")

	sequencePipelines(items, nil)

	if got, want := namesOf(items), []string{"aws", "azure", "linux"}; !reflect.DeepEqual(got, want) {
		t.Errorf("sequence = %v, want %v", got, want)
	}
}

// A saved order outlives the pipelines it names: one deleted since must not
// leave a hole or move anything.
func TestSavedNamesThatNoLongerExistAreIgnored(t *testing.T) {
	items := pipelinesFileOrdered("aws", "azure", "linux")

	sequencePipelines(items, []string{"gone", "linux", "also-gone"})

	if got, want := namesOf(items), []string{"linux", "aws", "azure"}; !reflect.DeepEqual(got, want) {
		t.Errorf("sequence = %v, want %v", got, want)
	}
	if items[2].Position != 2 {
		t.Errorf("last position = %d, want 2", items[2].Position)
	}
}

func TestTheFilesOwnOrderIsKept(t *testing.T) {
	items := pipelinesFileOrdered("aws", "azure", "linux")

	sequencePipelines(items, []string{"linux"})

	for _, p := range items {
		if p.Order != map[string]int32{"aws": 0, "azure": 1, "linux": 2}[pipelineIdentity(p.RelPath)] {
			t.Errorf("%s order = %d: it is the file's, and creating a pipeline reads it", p.RelPath, p.Order)
		}
	}
}

type fakePipelineStore struct {
	connectors.PipelineRepository
	all []domain.Pipeline
}

func (f fakePipelineStore) List(string) []domain.Pipeline { return f.all }

type fakeEngineConfig struct {
	connectors.EngineConfigRepository
	order []string
}

func (f fakeEngineConfig) PipelineOrder(string) []string { return f.order }

func shipped(name string, order int, system bool) domain.Pipeline {
	return domain.Pipeline{
		RelPath: name + ".yaml",
		System:  system,
		Active:  true,
		Content: []byte(fmt.Sprintf("pipeline:\n  - dataTypes: [%s]\n    order: %d\n    steps:\n      - drop: {}\n", name, order)),
	}
}

func listPipelines(t *testing.T, saved []string, f dto.PipelineFilters) []dto.PipelineResponse {
	t.Helper()
	uc := &pipelineUsecase{
		store: fakePipelineStore{all: []domain.Pipeline{
			shipped("aws", 0, true), shipped("azure", 1, true), shipped("custom", 2, false), shipped("linux", 3, true),
		}},
		config: fakeEngineConfig{order: saved},
	}
	res, err := uc.List(context.Background(), f)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	return res.Items
}

func TestListShowsTheTenantsSavedOrder(t *testing.T) {
	got := listPipelines(t, []string{"linux", "custom"}, dto.PipelineFilters{})

	if want := []string{"linux", "custom", "aws", "azure"}; !reflect.DeepEqual(namesOf(got), want) {
		t.Errorf("listing = %v, want %v", namesOf(got), want)
	}
}

// A view narrowed to some of the pipelines still has to say where each runs
// among all of them, not renumber the few it shows.
func TestPositionSurvivesFiltering(t *testing.T) {
	custom := false
	got := listPipelines(t, []string{"linux", "custom"}, dto.PipelineFilters{SystemEq: &custom})

	if len(got) != 1 || got[0].RelPath != "custom.yaml" {
		t.Fatalf("listing = %v, want only custom", namesOf(got))
	}
	if got[0].Position != 1 {
		t.Errorf("position = %d, want 1: linux runs before it", got[0].Position)
	}
}
