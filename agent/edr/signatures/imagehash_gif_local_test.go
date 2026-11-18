package signatures

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestGIFLocalTransparentColorMatchesClamAV(t *testing.T) {
	// The pinned ClamAV 1.5.4 sigtool --fuzzy-img results for generated
	// GIFs with a colored transparent local entry, with and without a
	// conflicting global palette.
	for _, tc := range []struct {
		name string
		hash uint64
	}{
		{"gif_local_transparent.gif", 0xbe2cc98afa829f22},
		{"gif_local_exact.gif", 0x8c5e46ec6238fb51},
		{"gif_local_override.gif", 0x8c5e46ec6238fb51},
	} {
		data, err := os.ReadFile("testdata/" + tc.name)
		if err != nil {
			t.Fatal(err)
		}
		got := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !got.ok || got.hash != tc.hash || (tc.name != "gif_local_transparent.gif" && got.uncertain != 0) {
			t.Fatalf("%s: got=%016x uncertain=%016x ok=%v, want=%016x", tc.name, got.hash, got.uncertain, got.ok, tc.hash)
		}
	}
	data, err := os.ReadFile("testdata/gif_local_exact.gif")
	if err != nil {
		t.Fatal(err)
	}
	const referenceHash uint64 = 0x8c5e46ec6238fb51
	rule := func(hash uint64) *Database {
		return compileTest(t, "ldb", fmt.Sprintf("GIF.Local;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
	}
	ctx := MatchContext{Type: TypeGIF}
	expectOutcome(t, rule(referenceHash), string(data), ctx, "GIF.Local", "")
	expectOutcome(t, rule(referenceHash^1<<40), string(data), ctx, "", "")
}
