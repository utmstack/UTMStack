package signatures

import (
	"context"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

func TestPreclassBareAndTriggeredPhase(t *testing.T) {
	bare := callbackWithLabel([]byte("Preclass.Bare"), true)
	bare.kind, bare.logical = 260, ""
	triggered := callbackWithLabel([]byte("Preclass.Triggered"), true)
	triggered.kind, triggered.logical = 260, "Preclass.Trigger;Engine:51-255,Target:0;0;4142"
	d, err := Compile(map[string][]byte{"a.cbc": triggered.encode(), "b.cbc": bare.encode()})
	if err != nil {
		t.Fatal(err)
	}
	if d.DetectionCount() != 2 || len(d.preclassHooks) != 2 {
		t.Fatalf("compile preclass: count=%d hooks=%d err=%v", d.DetectionCount(), len(d.preclassHooks), err)
	}
	flags := make(map[*bytecode.Program]bool)
	res, err := d.MatchObject(context.Background(), []byte("xxAByy"), MatchContext{Type: TypeAny, PreclassMatched: func(p *bytecode.Program) { flags[p] = true }})
	if err != nil || res.Name != "" || len(flags) != 1 {
		t.Fatalf("trigger phase: %+v, %v, flags=%d", res, err, len(flags))
	}
	name, err := d.MatchPreclass(context.Background(), []byte("xxAByy"), flags, MatchContext{Metadata: map[string]any{}, MetadataPartial: true})
	if err != nil || name != "Preclass.Triggered" {
		t.Fatalf("preclass phase: %q, %v", name, err)
	}
	name, err = d.MatchPreclass(context.Background(), []byte("ordinary"), nil, MatchContext{Metadata: map[string]any{}, MetadataPartial: true})
	if err != nil || name != "Preclass.Bare" {
		t.Fatalf("bare preclass: %q, %v", name, err)
	}
	if res, err := d.MatchObject(context.Background(), []byte("xxAByy"), MatchContext{Type: TypeAny}); res.Name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("unhandled trigger: %+v, %v", res, err)
	}
}

func TestLegacyMetadataCancellationPrecedesIncomplete(t *testing.T) {
	d, err := Compile(map[string][]byte{"legacy.ndb": []byte("Legacy.Rule:13:*:414243444546\n")})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if name, err := d.MatchLegacyMetadata(ctx, nil, false, MatchContext{}); name != "" || err != context.Canceled {
		t.Fatalf("cancelled partial metadata: name=%q err=%v", name, err)
	}
}

func TestPreclassWithoutMetadataIsIncomplete(t *testing.T) {
	f := callbackWithLabel([]byte("Preclass.Bare"), true)
	f.kind, f.logical = 260, ""
	d, err := Compile(map[string][]byte{"preclass.cbc": f.encode()})
	if err != nil {
		t.Fatal(err)
	}
	name, err := d.MatchPreclass(context.Background(), []byte("ordinary"), nil, MatchContext{})
	if name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("missing metadata: %q, %v", name, err)
	}
}

func TestPreclassBreakStopsLaterHooks(t *testing.T) {
	brk := callbackFixture{kind: 260, result: 32, blocks: [][]string{{callbackReturn(32, 0xcea5e)}}}
	label := callbackWithLabel([]byte("Preclass.Later"), true)
	label.kind, label.logical = 260, ""
	d, err := Compile(map[string][]byte{"a.cbc": brk.encode(), "b.cbc": label.encode()})
	if err != nil {
		t.Fatal(err)
	}
	name, err := d.MatchPreclass(context.Background(), []byte("ordinary"), nil, MatchContext{Metadata: map[string]any{}})
	if err != nil || name != "" {
		t.Fatalf("break hook continued: %q, %v", name, err)
	}
}

func TestUnsupportedBarePreclassStaysIncomplete(t *testing.T) {
	f := callbackWithLabel([]byte("Preclass.Unsupported"), true)
	f.kind, f.logical = 260, ""
	f.apiMax = 53
	f.apis = append(f.apis, callbackAPI(53, 70, "atoi"))
	d, err := Compile(map[string][]byte{"preclass.cbc": f.encode()})
	if err != nil {
		t.Fatal(err)
	}
	if !d.unknownPreclass {
		t.Fatal("unsupported preclass was not retained as unknown")
	}
	name, err := d.MatchPreclass(context.Background(), []byte("ordinary"), nil, MatchContext{Metadata: map[string]any{}})
	if name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("unknown preclass verdict: %q, %v", name, err)
	}
}

func TestUnsupportedTriggeredPreclassIsConditional(t *testing.T) {
	f := callbackWithLabel([]byte("Preclass.Unsupported"), true)
	f.kind, f.logical = 260, "Preclass.Trigger;Engine:51-255,Target:0;0;4142"
	f.apiMax = 53
	f.apis = append(f.apis, callbackAPI(53, 70, "atoi"))
	d, err := Compile(map[string][]byte{"preclass.cbc": f.encode()})
	if err != nil {
		t.Fatal(err)
	}
	flags := make(map[*bytecode.Program]bool)
	res, err := d.MatchObject(context.Background(), []byte("xxAByy"), MatchContext{Type: TypeAny, PreclassMatched: func(p *bytecode.Program) { flags[p] = true }})
	if err != nil || res.Name != "" || !flags[nil] {
		t.Fatalf("unknown root trigger: %+v, %v, flags=%v", res, err, flags)
	}
	name, err := d.MatchPreclass(context.Background(), []byte("xxAByy"), flags, MatchContext{Metadata: map[string]any{}})
	if name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("reached unsupported hook: %q, %v", name, err)
	}
	name, err = d.MatchPreclass(context.Background(), []byte("ordinary"), nil, MatchContext{Metadata: map[string]any{}})
	if name != "" || err != nil {
		t.Fatalf("unreached unsupported hook: %q, %v", name, err)
	}
}

func TestPreclassRefreshReplacesProgram(t *testing.T) {
	f := callbackWithLabel([]byte("Preclass.Old"), true)
	f.kind, f.logical = 260, ""
	members := []member{{"preclass.cbc", f.encode()}, {"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")}}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	f = callbackWithLabel([]byte("Preclass.New"), true)
	f.kind, f.logical = 260, ""
	members[0].data = f.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	for _, tc := range []struct {
		db   *Database
		want string
	}{{prior, "Preclass.Old"}, {next, "Preclass.New"}} {
		got, err := tc.db.MatchPreclass(context.Background(), []byte("ordinary"), nil, MatchContext{Metadata: map[string]any{}})
		if err != nil || got != tc.want {
			t.Fatalf("refresh preclass: %q, %v; want %q", got, err, tc.want)
		}
	}
}
