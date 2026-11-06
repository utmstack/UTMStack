package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/threatwinds/go-sdk/store"
)

// alrF18FakeReader lets FetchPage behave differently per call: the first call
// is the normal time-bounded page, the second (if any) is the unbounded
// fetch-by-id for roots whose own timestamp fell outside the window.
type alrF18FakeReader struct {
	pageDocs    []json.RawMessage
	pageTotal   int64
	echoBuckets []store.Bucket
	rootsByID   map[string]json.RawMessage

	fetchCalls      int
	topValuesCalled bool
	lastUnbounded   store.Scope
}

func (f *alrF18FakeReader) DescribeFields(context.Context, store.Scope) ([]store.Field, error) {
	return nil, nil
}

func (f *alrF18FakeReader) FetchPage(_ context.Context, scope store.Scope, filters []store.Filter, _ store.Page) ([]json.RawMessage, int64, error) {
	// includeRootsWithRecentEchoes is tested directly (not through Search), so
	// the only FetchPage call it can make is its own unbounded fetch-by-id for
	// roots missing from the page.
	f.fetchCalls++
	f.lastUnbounded = scope
	var ids []string
	for _, flt := range filters {
		if flt.Field == "id" && flt.Op == store.OpIn {
			ids, _ = flt.Value.([]string)
		}
	}
	out := make([]json.RawMessage, 0, len(ids))
	for _, id := range ids {
		if doc, ok := f.rootsByID[id]; ok {
			out = append(out, doc)
		}
	}
	return out, int64(len(out)), nil
}

func (f *alrF18FakeReader) TopValues(_ context.Context, _ store.Scope, _ string, _ []store.Filter, _ int) ([]store.Bucket, error) {
	f.topValuesCalled = true
	return f.echoBuckets, nil
}

func (f *alrF18FakeReader) Timeline(context.Context, store.Scope, []store.Filter, store.Interval) ([]store.Point, error) {
	return nil, nil
}

func (f *alrF18FakeReader) Count(context.Context, store.Scope, []store.Filter) (int64, error) {
	return 0, nil
}

// ALR-F18: a root alert whose own timestamp is older than the requested
// window, but which has a fresh echo inside that window, must still appear
// in the result.
func TestIncludeRootsWithRecentEchoesAddsAnOldRootWithAFreshEcho(t *testing.T) {
	oldRoot := mustJSON(t, map[string]any{"id": "old-root", "parentId": ""})
	visible := mustJSON(t, map[string]any{"id": "visible", "parentId": ""})

	reader := &alrF18FakeReader{
		pageDocs:    []json.RawMessage{visible},
		pageTotal:   1,
		echoBuckets: []store.Bucket{{Key: "old-root", Count: 3}},
		rootsByID:   map[string]json.RawMessage{"old-root": oldRoot},
	}
	repo := &chAnalyzerRepository{store: reader}
	scope := store.Scope{From: time.Now().Add(-time.Hour), To: time.Now()}

	docs, total, err := repo.includeRootsWithRecentEchoes(context.Background(), scope, reader.pageDocs, reader.pageTotal, 10, "@timestamp", store.Desc)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if len(docs) != 2 {
		t.Fatalf("docs = %d, want 2", len(docs))
	}

	var ids []string
	for _, raw := range docs {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, m["id"].(string))
	}
	if ids[0] != "visible" || ids[1] != "old-root" {
		t.Fatalf("docs ids = %v, want [visible old-root]", ids)
	}

	// The extra fetch for the missing root must not be time-bounded — that's
	// the whole point of this path.
	if !reader.lastUnbounded.From.IsZero() || !reader.lastUnbounded.To.IsZero() {
		t.Fatalf("extra fetch was time-bounded: %+v, want unbounded", reader.lastUnbounded)
	}
}

// A root already present on the page must not be re-fetched or duplicated.
func TestIncludeRootsWithRecentEchoesSkipsRootsAlreadyOnThePage(t *testing.T) {
	present := mustJSON(t, map[string]any{"id": "present", "parentId": ""})
	reader := &alrF18FakeReader{
		pageDocs:    []json.RawMessage{present},
		pageTotal:   1,
		echoBuckets: []store.Bucket{{Key: "present", Count: 7}},
	}
	repo := &chAnalyzerRepository{store: reader}
	scope := store.Scope{From: time.Now().Add(-time.Hour), To: time.Now()}

	docs, total, err := repo.includeRootsWithRecentEchoes(context.Background(), scope, reader.pageDocs, reader.pageTotal, 10, "@timestamp", store.Desc)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if total != 1 || len(docs) != 1 {
		t.Fatalf("total=%d len(docs)=%d, want 1/1 (no duplicate fetch of an already-present root)", total, len(docs))
	}
	if reader.fetchCalls != 0 {
		t.Fatalf("FetchPage called %d times, want 0 (nothing missing, no extra fetch needed)", reader.fetchCalls)
	}
}

// An ordinary log-search (no time window requested at all) must be completely
// unaffected — no extra TopValues/FetchPage round trip.
func TestIncludeRootsWithRecentEchoesIsInertWithoutATimeWindow(t *testing.T) {
	docs := []json.RawMessage{mustJSON(t, map[string]any{"id": "a", "parentId": ""})}
	reader := &alrF18FakeReader{pageDocs: docs, pageTotal: 1}
	repo := &chAnalyzerRepository{store: reader}

	got, total, err := repo.includeRootsWithRecentEchoes(context.Background(), store.Scope{}, docs, 1, 10, "@timestamp", store.Desc)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if total != 1 || len(got) != 1 {
		t.Fatalf("total=%d len=%d, want unchanged 1/1", total, len(got))
	}
	if reader.topValuesCalled {
		t.Fatal("TopValues was called for a request with no time window — plain log-search must be untouched")
	}
}

// Once the page is already full (no room left), no extra root is spliced in.
func TestIncludeRootsWithRecentEchoesRespectsNoRoomLeft(t *testing.T) {
	docs := []json.RawMessage{mustJSON(t, map[string]any{"id": "a", "parentId": ""})}
	reader := &alrF18FakeReader{pageDocs: docs, pageTotal: 1, echoBuckets: []store.Bucket{{Key: "old-root", Count: 1}}}
	repo := &chAnalyzerRepository{store: reader}
	scope := store.Scope{From: time.Now().Add(-time.Hour), To: time.Now()}

	got, total, err := repo.includeRootsWithRecentEchoes(context.Background(), scope, docs, 1, 0, "@timestamp", store.Desc)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if total != 1 || len(got) != 1 {
		t.Fatalf("total=%d len=%d, want unchanged 1/1 when room=0", total, len(got))
	}
}
