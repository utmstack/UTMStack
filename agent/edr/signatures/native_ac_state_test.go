package signatures

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
)

func TestNativeACSharedCountUsesLargestConsumerCap(t *testing.T) {
	files := map[string][]byte{"rules.ldb": []byte("Low;Target:0;0=0;4142*4344\nHigh;Target:0;0>4;4142*4344\n")}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	input := []byte("ABxCDCDCDCDCD")
	assertMatch(t, d, input, "High")
	nativeFixedACControl(t, input, files, "High")
}

func TestNativeACByteCompareRetainsLastTrigger(t *testing.T) {
	d := compileTest(t, "ldb", "Compare;Engine:100-255,Target:0;1;4142{1-2}4344;0(>>5#ib1#=63)\n")
	for _, tc := range []struct{ input, want string }{
		{"xxABxCD!yyABxCD?", "Compare"},
		{"xxABxCD?yyABxCD!", ""},
	} {
		assertMatch(t, d, []byte(tc.input), tc.want)
	}
}

func TestNativeACOffsetPruningAndHardBudget(t *testing.T) {
	// A short, common key belongs to a long body at one absolute offset.
	// Scanning every apparent key through the long suffix is quadratic in
	// this control. Offset clipping must leave enough work for the valid end.
	body := "4141" + "??" + "41414141414141414141414141414141"
	d := compileTest(t, "ndb", "Absolute:0:64000:"+body+"\n")
	input := bytes.Repeat([]byte{'A'}, 65000)
	name, steps, err := d.matchCounted(context.Background(), input, 500_000)
	if err != nil || name != "Absolute" {
		t.Fatalf("offset clipping: %q steps=%d err=%v", name, steps, err)
	}
	if _, steps, err = d.matchCounted(context.Background(), input, 100); !errors.Is(err, ErrMatchLimit) || steps > 100 {
		t.Fatalf("hard work bound: steps=%d err=%v", steps, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err = d.matchCounted(ctx, input, 500_000); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation escaped", err)
	}
}

func TestNativeACMultipartRefreshKeepsOldHistoriesIndependent(t *testing.T) {
	files := map[string][]byte{"rules.ldb": []byte("Old;Target:0;0=3;4142*4344\n"), "change.ndb": []byte("Other:0:*:7879*7a7a\n")}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	input := []byte("ABxCDCDCD")
	assertMatch(t, prior, input, "Old")
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for range 50 {
				name, err := prior.Match(context.Background(), input)
				if err != nil || name != "Old" {
					t.Errorf("old multipart snapshot: %q %v", name, err)
					return
				}
			}
		})
	}
	defer readers.Wait()
	current := prior
	for i := range 10 {
		files["change.ndb"] = []byte("Other:0:*:7879*7a7a\n")
		want := "Old"
		if i%2 == 0 {
			files["change.ndb"] = []byte("New:0:*:4142*4344\n")
			want = "New"
		}
		next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: current})
		if err != nil || !handled {
			t.Fatalf("multipart refresh: handled=%v err=%v", handled, err)
		}
		assertMatch(t, next, input, want)
		assertMatch(t, prior, input, "Old")
		current = next
	}
}

func TestNativeACFallbackKeepsGraphAnchorStarts(t *testing.T) {
	files := map[string][]byte{"rules.ndb": []byte("Character:0:*:41[1-3]42434445\nFallback:0:*:*78797a\n")}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if d.nativeAC[0] != nil {
		t.Fatal("control must use the graph fallback")
	}
	assertMatch(t, d, []byte("AxxBCDE"), "Character")
}
