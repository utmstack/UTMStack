package signatures

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// sig-4: one pattern exhausting its own state bound must not hide a definite
// detection by an independent rule. Detection wins over incomplete, as in the
// archive walkers; with no detection the result stays incomplete.
func TestPatternStateLimitDoesNotHideDetection(t *testing.T) {
	marker := hex.EncodeToString([]byte("EVILMARKER"))
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("Generic.Gap:0:*:616161*(6262|6363)\n"),
		"b.ndb": []byte("Known.Malware:0:*:" + marker + "\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	gapOnly, err := Compile(map[string][]byte{"a.ndb": []byte("Generic.Gap:0:*:616161*(6262|6363)\n")})
	if err != nil {
		t.Fatal(err)
	}
	prefix := []byte("EVILMARKER aaa")
	small := append(append([]byte(nil), prefix...), bytes.Repeat([]byte{0}, 1000)...)
	if got, err := d.Match(context.Background(), small); got != "Known.Malware" || err != nil {
		t.Fatalf("small: got %q err %v", got, err)
	}
	// Every byte after the prefix can begin the alternative, and none completes
	// it, so the gap search reaches its per-pattern state bound; the forward
	// search that then decides the start is given a bound it reaches too.
	defer func(steps int) { maxForwardSteps = steps }(maxForwardSteps)
	maxForwardSteps = 1000
	hostile := append(append([]byte(nil), prefix...), bytes.Repeat([]byte("bc"), 150000)...)
	if got, err := d.Match(context.Background(), hostile); got != "Known.Malware" || err != nil {
		t.Fatalf("hostile suffix turned the detection into %q, %v", got, err)
	}
	got, err := gapOnly.Match(context.Background(), hostile)
	if got != "" || !errors.Is(err, ErrMatchLimit) || !errors.Is(err, ErrRuleIncomplete) {
		t.Fatalf("undecided pattern alone: got %q, %v; want incomplete", got, err)
	}
}

// Benign bytes that cannot begin the continuation after a gap are skipped, so a
// large ordinary file no longer exhausts the per-pattern state bound.
func TestGapSkipsBytesThatCannotContinue(t *testing.T) {
	d, err := Compile(map[string][]byte{"a.ndb": []byte("Generic.Gap:0:*:616161*(6262|6363)\n")})
	if err != nil {
		t.Fatal(err)
	}
	large := append([]byte("aaa"), bytes.Repeat([]byte{0}, 300000)...)
	if got, err := d.Match(context.Background(), large); got != "" || err != nil {
		t.Fatalf("benign large input: got %q, %v", got, err)
	}
	large = append(large, "cc"...)
	if got, err := d.Match(context.Background(), large); got != "Generic.Gap" || err != nil {
		t.Fatalf("late continuation: got %q, %v", got, err)
	}
	for _, tc := range []struct{ sig, data, want string }{
		{"4142*?f", "AB\x00\x00\x1f", "P"},
		{"4142*?f", "AB\x00\x00\x20", ""},
		{"4142*6364::i", "AB\x00CD", "P"},
		{"4142*6364::i", "AB\x00CE", ""},
		{"4142*(6364|4546)::i", "AB\x00ef", "P"},
		{"4142*!(4344|4546)", "ABCD", ""},
		{"4142*!(4344|4546)", "ABCDxy", "P"},
		{"4142*(43|44)::f", "ABxD", "P"},
		{"4142*(43|44)::f", "ABxDz", ""},
	} {
		d := compileTest(t, "ldb", "P;Target:0;0;"+tc.sig)
		assertMatch(t, d, []byte(tc.data), tc.want)
	}
}

// bc-1: a logical bytecode rule that fails at runtime or returns an
// unsupported (PUA/heuristic) label must not stop the evaluation of later
// ordinary rules. A detection by any other rule wins over the bytecode rule's
// incomplete outcome; alone, the failure is still reported.
func TestBytecodeFailureDoesNotHideLaterDetection(t *testing.T) {
	abort := callbackWithLabel([]byte("Win.Test.Aborted"), true)
	abort.blocks[0] = append(abort.blocks[0][:1], callbackInst(0, 0, 45), callbackReturn(32, 0))
	cases := map[string]struct {
		fixture callbackFixture
		cause   error
	}{
		"pua-label":     {callbackWithLabel([]byte("PUA.Win.Packer.Test"), true), bytecode.ErrUnsupported},
		"runtime-abort": {abort, bytecode.ErrRuntime},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d, err := Compile(map[string][]byte{
				"a.cbc": tc.fixture.encode(),
				"b.ndb": []byte("Win.Test.Plain:0:*:4142\n"),
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := d.Match(context.Background(), []byte("xxABxx"))
			if got != "Win.Test.Plain" || err != nil {
				t.Fatalf("Match = %q, %v; want the later plain detection", got, err)
			}
			alone := compileCallback(t, tc.fixture)
			got, err = alone.Match(context.Background(), []byte("xxABxx"))
			if got != "" || !errors.Is(err, tc.cause) || !errors.Is(err, ErrRuleIncomplete) {
				t.Fatalf("failure alone: got %q, %v", got, err)
			}
		})
	}
	d, err := Compile(map[string][]byte{"b.ndb": []byte("Win.Test.Plain:0:*:4142\n")})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := d.Match(context.Background(), []byte("xxABxx")); got != "Win.Test.Plain" || err != nil {
		t.Fatalf("control Match = %q, %v", got, err)
	}
}

// The shared work budget still bounds the whole call, including failed
// bytecode runs, and cancellation still stops it: every budget gives either the
// detection or a work-limit error, never a clean verdict.
func TestSharedBudgetAndCancellationStillAbort(t *testing.T) {
	files := map[string][]byte{
		"a.cbc": callbackWithLabel([]byte("PUA.Win.Packer.Test"), true).encode(),
		"b.ndb": []byte("Win.Test.Plain:0:*:4142\n"),
	}
	limited, detected := 0, 0
	for steps := int64(1); steps <= 400; steps++ {
		d, err := CompileWithLimits(files, Limits{MaxMatchSteps: steps})
		if err != nil {
			t.Fatal(err)
		}
		got, err := d.Match(context.Background(), []byte("xxABxx"))
		switch {
		case got == "Win.Test.Plain" && err == nil:
			detected++
		case got == "" && errors.Is(err, ErrMatchLimit):
			limited++
		default:
			t.Fatalf("budget %d: got %q, %v", steps, got, err)
		}
	}
	if limited == 0 || detected == 0 {
		t.Fatalf("budget sweep did not cover both outcomes: %d limited, %d detected", limited, detected)
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := d.Match(ctx, []byte("xxABxx")); got != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled: got %q, %v", got, err)
	}
}
