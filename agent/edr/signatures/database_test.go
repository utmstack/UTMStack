package signatures

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func compileTest(t *testing.T, ext, text string) *Database {
	t.Helper()
	d, e := Compile(map[string][]byte{"test." + ext: []byte(text)})
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func assertMatch(t *testing.T, d *Database, data []byte, want string) {
	t.Helper()
	got, e := d.Match(context.Background(), data)
	if e != nil || got != want {
		t.Fatalf("Match(%q)=(%q,%v), want %q", data, got, e, want)
	}
}
func TestHashSizeAndAllowlist(t *testing.T) {
	payload := []byte("malicious fixture")
	hash := sha256.Sum256(payload)
	d := compileTest(t, "hsb", fmt.Sprintf("%x:%d:Exact\n%x:*:Any:73\n", hash, len(payload)+1, hash))
	assertMatch(t, d, payload, "Any")
	assertMatch(t, d, []byte("clean"), "")
	h := md5.Sum(payload)
	d, e := Compile(map[string][]byte{"test.hdb": []byte(fmt.Sprintf("%x:*:Bad:73", h)), "test.fp": []byte(fmt.Sprintf("%x:%d:Allowed", h, len(payload)))})
	if e != nil {
		t.Fatal(e)
	}
	assertMatch(t, d, payload, "")
}
func TestExtendedPatternsAndOffsets(t *testing.T) {
	tests := []struct {
		sig, data string
		want      bool
	}{
		{"*:4142??4?", "ABxO", true}, {"*:4142??4?", "ABxP", false},
		{"*:4142?f", "AB\x9f", true}, {"*:4142{2-4}4344", "ABxyzCD", true},
		{"*:4142{2-4}4344", "ABxCD", false}, {"*:4142*4344", "ABCD", true},
		{"*:4142{-2}4344", "ABxCD", true}, {"*:4142{2-}4344", "ABxxxxxCD", true},
		{"*:4142(4344|4546)", "xxABEFzz", true},
		{"*:4142!(4344|4546)", "ABGH", true}, {"*:4142!(4344|4546)", "ABCD", false},
		{"0:41424344", "xABCD", false}, {"1,3:41424344", "xxxABCD", true},
		{"EOF-4:41424344", "xxxABCD", true}, {"EOF-4:41424344", "ABCDxxx", false},
		{"*:4142(4344|4546??)", "ABEFX", true},
	}
	for _, tt := range tests {
		t.Run(tt.sig+tt.data, func(t *testing.T) {
			d := compileTest(t, "ndb", "Pattern:0:"+tt.sig)
			want := ""
			if tt.want {
				want = "Pattern"
			}
			assertMatch(t, d, []byte(tt.data), want)
		})
	}
}
func TestLogical(t *testing.T) {
	d := compileTest(t, "ldb", "Logic;Engine:51-255,Target:0,FileSize:1-100;(0&1)&2=0;4142;4344;4546")
	assertMatch(t, d, []byte("AB CD"), "Logic")
	assertMatch(t, d, []byte("AB CD EF"), "")
	assertMatch(t, d, []byte("AB"), "")
	d = compileTest(t, "ldb", "Repeated;Target:0;(0|1)>2,2;4142;4344")
	assertMatch(t, d, []byte("ABABCD"), "Repeated")
	assertMatch(t, d, []byte("ABABAB"), "")
	d = compileTest(t, "ldb", "Overlap;Target:0;0=2;414141")
	assertMatch(t, d, []byte("AAAA"), "Overlap")
	d = compileTest(t, "ldb", "Words;Target:0;0;616263::if")
	assertMatch(t, d, []byte(" ABC!"), "Words")
	assertMatch(t, d, []byte("xABC!"), "")
}
func TestStrictRejection(t *testing.T) {
	tests := map[string]string{"cbc": "unknown", "yar": "rule foo {}", "ndb": "Unsafe:0:$0$:4142", "ldb": "Unsafe;Target:0,Container:CL_TYPE_ZIP;0;4142"}
	for ext, sig := range tests {
		t.Run(ext, func(t *testing.T) {
			if _, e := Compile(map[string][]byte{"bad." + ext: []byte(sig)}); e == nil {
				t.Fatal("unsupported signature accepted")
			}
		})
	}
	for _, sig := range []string{"Bad:0:*:??", "Bad:0:*:4142{9-1}4344", "Bad:0:*:4142(4|55)", "Bad:0:BOGUS:4142", "Bad:0:*:4142:999", "Bad:0:*:4142:0:20"} {
		if _, e := Compile(map[string][]byte{"bad.ndb": []byte(sig)}); e == nil {
			t.Errorf("accepted %s", sig)
		}
	}
}
func testPE() []byte {
	data := make([]byte, 1024)
	copy(data, "MZ")
	binary.LittleEndian.PutUint32(data[0x3c:], 128)
	copy(data[128:], "PE\x00\x00")
	binary.LittleEndian.PutUint16(data[132:], 0x14c)
	binary.LittleEndian.PutUint16(data[134:], 1)
	binary.LittleEndian.PutUint16(data[148:], 224)
	binary.LittleEndian.PutUint16(data[152:], 0x10b)
	binary.LittleEndian.PutUint32(data[168:], 0x1000)
	binary.LittleEndian.PutUint32(data[184:], 4096)
	binary.LittleEndian.PutUint32(data[188:], 512)
	binary.LittleEndian.PutUint32(data[212:], 512)
	copy(data[376:], ".text")
	binary.LittleEndian.PutUint32(data[384:], 512)
	binary.LittleEndian.PutUint32(data[388:], 0x1000)
	binary.LittleEndian.PutUint32(data[392:], 512)
	binary.LittleEndian.PutUint32(data[396:], 512)
	copy(data[512:], "ABCD")
	return data
}
func TestPEOffsetsAndSectionHash(t *testing.T) {
	pe := testPE()
	for _, offset := range []string{"EP+0", "S0+0", "SL+0", "SE0"} {
		d := compileTest(t, "ndb", "PE:1:"+offset+":41424344")
		assertMatch(t, d, pe, "PE")
		assertMatch(t, d, []byte("ABCD"), "")
		// An unmapped entry point makes the reference reject PE offsets.
		if got, e := d.Match(context.Background(), append(brokenPE(), "ABCD"...)); got != "" || e != nil {
			t.Fatalf("Match(broken header)=(%q,%v), want clean", got, e)
		}
		// Without a PE header the reference fails the header: no offsets.
		if got, e := d.Match(context.Background(), []byte("MZABCD")); got != "" || e != nil {
			t.Fatalf("Match(MZABCD)=(%q,%v), want clean", got, e)
		}
	}
	hash := md5.Sum(pe[512:])
	d := compileTest(t, "mdb", fmt.Sprintf("512:%x:Section", hash))
	assertMatch(t, d, pe, "Section")
}
func cvdFixture(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var payload bytes.Buffer
	z := gzip.NewWriter(&payload)
	tw := tar.NewWriter(z)
	for name, b := range files {
		if e := tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(b)), Mode: 0600}); e != nil {
			t.Fatal(e)
		}
		if _, e := tw.Write(b); e != nil {
			t.Fatal(e)
		}
	}
	if e := tw.Close(); e != nil {
		t.Fatal(e)
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	header := fmt.Sprintf("ClamAV-VDB:now:123:1:90:%x:unverified:builder:1", md5.Sum(payload.Bytes()))
	return append([]byte(header+strings.Repeat(" ", 512-len(header))), payload.Bytes()...)
}

// cvdFixtureOrdered builds a container whose members keep the given order.
func cvdFixtureOrdered(t *testing.T, members []member) []byte {
	t.Helper()
	var payload bytes.Buffer
	z := gzip.NewWriter(&payload)
	tw := tar.NewWriter(z)
	for _, m := range members {
		if e := tw.WriteHeader(&tar.Header{Name: m.name, Size: int64(len(m.data)), Mode: 0600}); e != nil {
			t.Fatal(e)
		}
		if _, e := tw.Write(m.data); e != nil {
			t.Fatal(e)
		}
	}
	if e := tw.Close(); e != nil {
		t.Fatal(e)
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	header := fmt.Sprintf("ClamAV-VDB:now:123:1:90:%x:unverified:builder:1", md5.Sum(payload.Bytes()))
	return append([]byte(header+strings.Repeat(" ", 512-len(header))), payload.Bytes()...)
}
func TestCVDIntegrityAndUntrustedMetadata(t *testing.T) {
	input := cvdFixture(t, map[string][]byte{"daily.ndb": []byte("Test:0:*:41424344")})
	d, e := Compile(map[string][]byte{"daily.cvd": input})
	if e != nil {
		t.Fatal(e)
	}
	assertMatch(t, d, []byte("ABCD"), "Test")
	report := d.Report()
	if len(report.Containers) != 1 || report.Containers[0].Authenticated || !report.Containers[0].IntegrityVerified {
		t.Fatalf("bad trust metadata: %+v", report)
	}
	input[len(input)-1] ^= 1
	if _, e := Compile(map[string][]byte{"daily.cvd": input}); e == nil {
		t.Fatal("modified CVD accepted")
	}
}
func TestCVDSafePaths(t *testing.T) {
	for _, path := range []string{"../evil.ndb", "/evil.ndb", "a/evil.ndb"} {
		if _, e := Compile(map[string][]byte{"daily.cvd": cvdFixture(t, map[string][]byte{path: []byte("Test:0:*:4142")})}); e == nil {
			t.Fatalf("unsafe path %s accepted", path)
		}
	}
}
func TestCancellationAndStableVersion(t *testing.T) {
	d := compileTest(t, "ndb", "Test:0:*:41424344")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := d.Match(ctx, []byte("ABCD")); e != context.Canceled {
		t.Fatal(e)
	}
	d2 := compileTest(t, "ndb", "Test:0:*:41424344")
	if d.Version() != d2.Version() || d.Count() != 1 {
		t.Fatal("unstable database identity")
	}
}

func TestHashAlgorithmsAndExactSizes(t *testing.T) {
	data := []byte("abc")
	for _, alg := range []struct{ extension, hash string }{{"hdb", "900150983cd24fb0d6963f7d28e17f72"}, {"hsb", "a9993e364706816aba3e25717850c26c9cd0d89d"}, {"hsb", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}} {
		d := compileTest(t, alg.extension, alg.hash+":3:Known")
		assertMatch(t, d, data, "Known")
		d = compileTest(t, alg.extension, alg.hash+":4:WrongSize")
		assertMatch(t, d, data, "")
	}
}
func TestNumericAndStateBudgets(t *testing.T) {
	d, e := CompileWithLimits(map[string][]byte{"test.ndb": []byte("Test:0:*:41424344")}, Limits{MaxMatchSteps: 8})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = d.Match(context.Background(), []byte(strings.Repeat("x", 100))); e != ErrMatchLimit {
		t.Fatalf("expected match limit, got %v", e)
	}
	_, e = CompileWithLimits(map[string][]byte{"test.ndb": []byte("Test:0:*:41424344")}, Limits{MaxFileBytes: 2})
	if e == nil {
		t.Fatal("oversized file accepted")
	}
	_, e = CompileWithLimits(map[string][]byte{"test.ndb": []byte("One:0:*:4142\nTwo:0:*:4344")}, Limits{MaxSignatures: 1})
	if e == nil {
		t.Fatal("signature limit ignored")
	}
}
func TestLogicalModifiersAndOffsets(t *testing.T) {
	tests := []struct{ sig, data, want string }{
		{"Word;Target:0;0;616263::i", "aBc", "Word"},
		{"Word;Target:0;0;616263::w", "a\x00b\x00c\x00", "Word"},
		{"Word;Target:0;0;616263::wa", "abc", "Word"},
		{"Word;Target:0;0;1:616263::i", "xaBc", "Word"},
		{"Word;Target:0;0;1:616263::i", "aBc", ""},
		{"Word;Target:0;0;616263::f", "abc_", "Word"},
		{"Word;Target:0;0;616263::f", "abc1", ""},
		{"Absent;Target:0;0=0;616263", "def", "Absent"},
		{"Absent;Target:0;0=0;616263", "abc", ""},
	}
	for _, tt := range tests {
		d := compileTest(t, "ldb", tt.sig)
		assertMatch(t, d, []byte(tt.data), tt.want)
	}
}
func TestPatternEdges(t *testing.T) {
	tests := []struct {
		sig, data string
		want      bool
	}{
		{"4142*4344", "ABxxxCDzzz", true}, {"4142*4344", "ABxxxC", false},
		{"4142{0}4344", "ABCD", true}, {"4142[0-1]43", "ABxC", true},
		{"(4142|4344)4546", "xxCDEF", true}, {"4142(43(44|45)|46)", "ABCE", true},
		{"!(4142|4344)", "EF", true}, {"!(4142|4344)", "AB", false},
		{"?f?f", "\x1f\x2f", true}, {"?f?f", "\x1f\x20", false},
	}
	for _, tt := range tests {
		d := compileTest(t, "ndb", "Pattern:0:*:"+tt.sig)
		want := ""
		if tt.want {
			want = "Pattern"
		}
		assertMatch(t, d, []byte(tt.data), want)
	}
}
func TestStrictLogicalAndNameRejection(t *testing.T) {
	for _, sig := range []string{
		"Bad;Target:0;1;4142", "Bad;Target:0;0&;4142", "Bad;Target:0;((0);4142",
		"Bad;Target:0;0;0/dead/i", "Bad;Target:0;0;${6-7}12$",
		"Bad;Target:0;0;4142::x", "Bad;Target:0,EntryPoint:0-1;0;4142",
		"Bad;Engine:999-1000,Target:0;0;4142", "Bad;Target:0,Target:0;0;4142",
	} {
		if _, e := Compile(map[string][]byte{"test.ldb": []byte(sig)}); e == nil {
			t.Errorf("accepted unsupported signature: %s", sig)
		}
	}
	for _, name := range []string{"bad\x00name", "bad\x1bname", "bad\u2028name"} {
		if _, e := Compile(map[string][]byte{"bad.ndb": []byte(name + ":0:*:4142")}); e == nil {
			t.Errorf("accepted unsafe name %q", name)
		}
	}
}
func TestContainerResourceLimits(t *testing.T) {
	fixture := cvdFixture(t, map[string][]byte{"many.ndb": []byte(strings.Repeat("Test:0:*:4142\n", 1000))})
	if _, e := CompileWithLimits(map[string][]byte{"daily.cvd": fixture}, Limits{MaxFileBytes: 4096}); e == nil {
		t.Fatal("expanded member limit ignored")
	}
	if _, e := CompileWithLimits(map[string][]byte{"daily.cvd": fixture}, Limits{MaxFiles: 1}); e == nil {
		t.Fatal("archive member count limit ignored")
	}
}
func TestDeterministicOrderingAndReportIsolation(t *testing.T) {
	files := map[string][]byte{"b.ndb": []byte("Second:0:*:4142"), "a.ndb": []byte("First:0:*:4142")}
	d, e := Compile(files)
	if e != nil {
		t.Fatal(e)
	}
	assertMatch(t, d, []byte("AB"), "First")
	report := d.Report()
	report.Files[0] = "mutation"
	if d.Report().Files[0] == "mutation" {
		t.Fatal("report exposes mutable storage")
	}
	if len(d.Version()) != 70 {
		t.Fatalf("expected complete SHA256 identity, got %q", d.Version())
	}
}
func FuzzMalformedSignatures(f *testing.F) {
	for _, s := range []string{"Test:0:*:41424344", "Test:0:*:4142{2-4}4344", "bad", "A:1:EP+0:4142"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 4096 {
			t.Skip()
		}
		d, e := CompileWithLimits(map[string][]byte{"fuzz.ndb": []byte(s)}, Limits{MaxFileBytes: 4096, MaxMatchSteps: 10000})
		if e == nil {
			_, _ = d.Match(context.Background(), []byte(s))
		}
	})
}
func FuzzMalformedPE(f *testing.F) {
	f.Add(testPE())
	f.Add([]byte("MZ"))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 1<<20 {
			t.Skip()
		}
		_ = parsePE(b)
	})
}

func TestPortableSignatureFilenames(t *testing.T) {
	for _, name := range []string{"con.hdb", "NUL.ndb", "COM1.ldb", "LPT9.hsb", "../x.ndb", "a\\b.hdb", "x.ndb.", ".hidden.ndb"} {
		if err := ValidateFileName(name); err == nil {
			t.Errorf("accepted %q", name)
		}
	}
	if err := ValidateFileName("daily-approved_1.ndb"); err != nil {
		t.Fatal(err)
	}
}
