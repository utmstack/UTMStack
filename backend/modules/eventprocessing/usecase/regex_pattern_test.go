package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/repository"
)

func newTestUsecase(t *testing.T) (*regexPatternUsecase, connectors.EngineConfigRepository, string) {
	t.Helper()
	dir := t.TempDir()
	w := repository.NewEngineConfig(dir)
	return &regexPatternUsecase{writer: w}, w, dir
}

func find(items []dto.RegexPatternResponse, id string) (dto.RegexPatternResponse, bool) {
	for _, it := range items {
		if it.PatternID == id {
			return it, true
		}
	}
	return dto.RegexPatternResponse{}, false
}

func listAll(t *testing.T, uc *regexPatternUsecase) []dto.RegexPatternResponse {
	t.Helper()
	res, err := uc.List(context.Background(), dto.RegexPatternFilters{Size: 200})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	return res.Items
}

// A fresh install has no patterns.yaml until the pipeline bootstrap runs during
// Module.Start — the API must still answer with the built-in vocabulary.
func TestListWithoutPatternsFile(t *testing.T) {
	uc, w, _ := newTestUsecase(t)

	items := listAll(t, uc)
	if len(items) != len(w.BuiltInPatterns()) {
		t.Fatalf("got %d patterns, want %d built-ins", len(items), len(w.BuiltInPatterns()))
	}
}

// The regression this fixes: patterns that reached patterns.yaml without going
// through this type — the bootstrap migrates them straight from the old DB table
// — used to resolve in filters but never appear in the API.
func TestListIncludesUserPatternsFromDisk(t *testing.T) {
	uc, w, _ := newTestUsecase(t)

	defs := map[string]string{"custom_id": `[A-Z]{3}-\d+`}
	for k, v := range w.BuiltInPatterns() {
		defs[k] = v
	}
	if err := w.WritePatterns(defs); err != nil {
		t.Fatalf("WritePatterns: %v", err)
	}

	got, ok := find(listAll(t, uc), "custom_id")
	if !ok {
		t.Fatal("custom_id from patterns.yaml is missing from List")
	}
	if got.PatternDefinition != `[A-Z]{3}-\d+` {
		t.Errorf("definition = %q, want %q", got.PatternDefinition, `[A-Z]{3}-\d+`)
	}
}

// patterns.yaml is what the engine actually resolves, so where it disagrees with
// the built-in table the file has to win — otherwise the API describes a pipeline
// that does not exist.
func TestDiskOverridesBuiltInDefinition(t *testing.T) {
	uc, w, _ := newTestUsecase(t)

	if err := w.WritePatterns(map[string]string{"greedy": "OVERRIDDEN"}); err != nil {
		t.Fatalf("WritePatterns: %v", err)
	}

	got, ok := find(listAll(t, uc), "greedy")
	if !ok {
		t.Fatal("greedy missing from List")
	}
	if got.PatternDefinition != "OVERRIDDEN" {
		t.Errorf("definition = %q, want the on-disk value", got.PatternDefinition)
	}
}

// A file the bootstrap has not written yet, or one an operator truncated, must
// not take the endpoint down.
func TestReadFailureDegradesToBuiltIns(t *testing.T) {
	uc, w, dir := newTestUsecase(t)

	// A directory where the file is expected makes ReadPatterns fail with
	// something other than "not exists".
	if err := writeUnreadablePatterns(t, dir); err != nil {
		t.Skipf("cannot stage an unreadable patterns.yaml: %v", err)
	}

	items := listAll(t, uc)
	if len(items) != len(w.BuiltInPatterns()) {
		t.Errorf("got %d patterns, want the %d built-ins as a fallback", len(items), len(w.BuiltInPatterns()))
	}
}

func TestGetByID(t *testing.T) {
	uc, w, _ := newTestUsecase(t)
	if err := w.WritePatterns(map[string]string{"custom_id": "abc"}); err != nil {
		t.Fatalf("WritePatterns: %v", err)
	}

	got, err := uc.GetByID(context.Background(), "custom_id")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.PatternDefinition != "abc" {
		t.Errorf("got %+v, want definition abc", *got)
	}

	if _, err := uc.GetByID(context.Background(), "nope"); !errors.Is(err, domain.ErrRegexPatternNotFound) {
		t.Errorf("err = %v, want ErrRegexPatternNotFound", err)
	}
}

func TestListFiltersAndPaging(t *testing.T) {
	uc, w, _ := newTestUsecase(t)
	if err := w.WritePatterns(map[string]string{"custom_a": "1", "custom_b": "2"}); err != nil {
		t.Fatalf("WritePatterns: %v", err)
	}
	ctx := context.Background()

	res, err := uc.List(ctx, dto.RegexPatternFilters{Search: "custom_", Size: 200})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(res.Items) != 2 {
		t.Errorf("user-only filter returned %d items, want 2", len(res.Items))
	}

	res, err = uc.List(ctx, dto.RegexPatternFilters{Search: "custom_a", Size: 200})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(res.Items) != 1 || res.Items[0].PatternID != "custom_a" {
		t.Errorf("search returned %+v, want just custom_a", res.Items)
	}

	res, err = uc.List(ctx, dto.RegexPatternFilters{Page: 0, Size: 3})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(res.Items) != 3 {
		t.Errorf("page size 3 returned %d items", len(res.Items))
	}
	if res.Total != int64(len(w.BuiltInPatterns())+2) {
		t.Errorf("Total = %d, want %d (all matches, not just the page)", res.Total, len(w.BuiltInPatterns())+2)
	}
}

// writeUnreadablePatterns puts a directory where patterns.yaml belongs, so the
// read fails for a reason other than absence.
func writeUnreadablePatterns(t *testing.T, dir string) error {
	t.Helper()
	return os.MkdirAll(filepath.Join(dir, repository.PatternsFileName), 0o755)
}
