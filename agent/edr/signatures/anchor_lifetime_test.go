package signatures

import (
	"context"
	"testing"
)

func TestPublishedIndexReleasesCompilerAnchorMetadata(t *testing.T) {
	d, err := Compile(map[string][]byte{"rules.ldb": []byte("Twice;Target:0;0=2;(4141|4242)\nGap;Target:0;0;43{1-4}44\n")})
	if err != nil {
		t.Fatal(err)
	}
	for id := range d.patterns {
		if p := &d.patterns[id]; p.graph != nil && len(p.graph.anchors) != 0 {
			t.Fatalf("pattern %d retains compile-only anchor metadata after index publication", id)
		}
		if d.patterns[id].everywhere() {
			t.Fatalf("anchored pattern %d became an unfiltered candidate", id)
		}
	}
	for _, tc := range []struct{ data, want string }{
		{"AAxxxBB", "Twice"}, {"BBxxxAA", "Twice"},
		{"AA", ""}, {"plain benign input", ""},
		{"C-D", "Gap"}, {"CD", ""}, {"C-----D", ""},
	} {
		got, err := d.Match(context.Background(), []byte(tc.data))
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %q, %v; want %q", tc.data, got, err, tc.want)
		}
	}
}

func TestDerivedFoldedLiteralAnchorAcrossGenerations(t *testing.T) {
	files := map[string][]byte{
		"a.ldb": []byte("Fold.Old;Target:0;0;414243::i\n"),
		"b.ldb": []byte("Plain.Kept;Target:0;0;4d4e4f\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ldb"] = []byte("Fold.New;Target:0;0;58595a::i\n")
	next, err := compileWithLimits(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil {
		t.Fatalf("new generation: %v", err)
	}
	for _, tc := range []struct {
		db   *Database
		data string
		want string
	}{
		{prior, "aBc", "Fold.Old"},
		{next, "aBc", ""},
		{next, "xYz", "Fold.New"},
		{next, "MNO", "Plain.Kept"},
	} {
		assertMatch(t, tc.db, []byte(tc.data), tc.want)
	}
	for _, d := range []*Database{prior, next} {
		for _, p := range d.patterns {
			if p.simple && p.graph != nil {
				t.Fatal("literal retained graph after generation update")
			}
		}
	}
}

// Inspect the immutable index rather than retaining a second per-pattern copy
// just for test coverage accounting. Folded keys can reference the same group.
func indexedPatternAnchors(d *Database) map[int][]patternAnchor {
	result := map[int][]patternAnchor{}
	seen := map[*anchorGroup]bool{}
	var walk func(*anchorBucket)
	walk = func(b *anchorBucket) {
		for _, g := range b.groups {
			if seen[g] {
				continue
			}
			seen[g] = true
			for _, m := range g.members {
				result[int(m.id)] = append(result[int(m.id)], patternAnchor{
					literal: g.literal, fold: g.fold, minPre: int(m.minPre), maxPre: int(m.maxPre),
				})
			}
		}
		for _, child := range b.chunk {
			walk(child)
		}
		for _, child := range b.single {
			walk(child)
		}
	}
	for _, b := range d.index.long {
		walk(b)
	}
	for _, table := range d.index.short {
		for _, b := range table {
			walk(b)
		}
	}
	return result
}
