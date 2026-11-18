package signatures

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestPreparedHashListMatchesDurableCompiler(t *testing.T) {
	data := []byte("an inert indicator-listed fixture body")
	m, s1, s256 := md5.Sum(data), sha1.Sum(data), sha256.Sum256(data)
	for _, tc := range []struct{ algorithm, ext, digest, empty string }{
		{"MD5", ".hdb", hex.EncodeToString(m[:]), "d41d8cd98f00b204e9800998ecf8427e"},
		{"SHA1", ".hsb", hex.EncodeToString(s1[:]), "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"SHA256", ".hsb", hex.EncodeToString(s256[:]), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	} {
		t.Run(tc.algorithm, func(t *testing.T) {
			wire, prepared, err := PrepareHashList([]byte("# source\n"+tc.empty+"\n"+strings.ToUpper(tc.digest)+"\r\n"+tc.digest+"\n"), tc.algorithm, 4096)
			if err != nil {
				t.Fatal(err)
			}
			row := fmt.Sprintf("%s:*:UTMStack.Intel.%s.%s:73\n", tc.digest, tc.algorithm, tc.digest)
			if string(wire) != row+row {
				t.Fatalf("durable compatibility changed: %q", wire)
			}
			files := map[string][]byte{"intel" + tc.ext: wire}
			legacy, err := Compile(files)
			if err != nil {
				t.Fatal(err)
			}
			native, localErr, fatalErr := CompileWithPreparedLocalFiles(files, nil,
				map[string]*PreparedHashList{"intel" + tc.ext: prepared}, Limits{})
			if native == nil || localErr != nil || fatalErr != nil {
				t.Fatalf("local=%v fatal=%v", localErr, fatalErr)
			}
			if native.Version() != legacy.Version() || native.DetectionCount() != legacy.DetectionCount() || !reflect.DeepEqual(native.Report(), legacy.Report()) {
				t.Fatalf("typed/text semantics differ: native=%+v legacy=%+v", native.Report(), legacy.Report())
			}
			assertMatch(t, native, data, "UTMStack.Intel."+tc.algorithm+"."+tc.digest)
		})
	}
}

func TestPreparedHashListRejectsUnboundOrMalformedSources(t *testing.T) {
	wire, prepared, err := PrepareHashList([]byte(strings.Repeat("a", 64)+"\n"), "SHA256", 4096)
	if err != nil {
		t.Fatal(err)
	}
	for name, alter := range map[string]func(map[string][]byte, map[string]*PreparedHashList){
		"stale bytes": func(files map[string][]byte, _ map[string]*PreparedHashList) {
			files["intel.hsb"] = append([]byte(nil), wire...)
			files["intel.hsb"][0] = 'b'
		},
		"truncated source": func(files map[string][]byte, _ map[string]*PreparedHashList) { files["intel.hsb"] = wire[:len(wire)-1] },
		"missing source":   func(_ map[string][]byte, p map[string]*PreparedHashList) { p["other.hsb"] = prepared },
		"nil prepared":     func(_ map[string][]byte, p map[string]*PreparedHashList) { p["intel.hsb"] = nil },
		"zero value":       func(_ map[string][]byte, p map[string]*PreparedHashList) { p["intel.hsb"] = &PreparedHashList{} },
		"truncated digests": func(_ map[string][]byte, p map[string]*PreparedHashList) {
			copy := *prepared
			copy.digests = copy.digests[:len(copy.digests)-1]
			p["intel.hsb"] = &copy
		},
		"invalid row bounds": func(_ map[string][]byte, p map[string]*PreparedHashList) {
			copy := *prepared
			copy.rowBytes++
			p["intel.hsb"] = &copy
		},
		"wrong source type": func(files map[string][]byte, p map[string]*PreparedHashList) {
			delete(files, "intel.hsb")
			delete(p, "intel.hsb")
			files["intel.ndb"] = wire
			p["intel.ndb"] = prepared
		},
	} {
		t.Run(name, func(t *testing.T) {
			files := map[string][]byte{"intel.hsb": wire}
			preparedMap := map[string]*PreparedHashList{"intel.hsb": prepared}
			alter(files, preparedMap)
			db, localErr, fatalErr := CompileWithPreparedLocalFiles(files, nil, preparedMap, Limits{})
			if db != nil || localErr != nil || fatalErr == nil {
				t.Fatalf("unbound native rules accepted: db=%v local=%v fatal=%v", db, localErr, fatalErr)
			}
		})
	}
	for _, tc := range []struct {
		data, algorithm string
		maxBytes        int
	}{
		{strings.Repeat("a", 64), "SHA256", 8},
		{strings.Repeat("a", 64), "other", 4096},
		{"# only comments\n", "SHA256", 4096},
		{strings.Repeat("a", 4097), "SHA256", 8192},
		{"not a digest\n", "SHA256", 4096},
	} {
		if _, _, err := PrepareHashList([]byte(tc.data), tc.algorithm, tc.maxBytes); err == nil {
			t.Fatalf("invalid constructor input accepted: %+v", tc)
		}
	}
	for name, limits := range map[string]Limits{
		"file bytes":  {MaxFileBytes: int64(len(wire) - 1)},
		"total bytes": {MaxTotalBytes: int64(len(wire) - 1)},
	} {
		t.Run(name, func(t *testing.T) {
			if db, _, err := CompileWithPreparedLocalFiles(map[string][]byte{"intel.hsb": wire}, nil,
				map[string]*PreparedHashList{"intel.hsb": prepared}, limits); err == nil || db != nil {
				t.Fatal("prepared source bypassed byte bounds")
			}
		})
	}
}

func TestPreparedHashListCanBeSharedDuringCompilation(t *testing.T) {
	wire, prepared, err := PrepareHashList([]byte(strings.Repeat("a", 64)+"\n"), "SHA256", 4096)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"intel.hsb": wire}
	preparedFiles := map[string]*PreparedHashList{"intel.hsb": prepared}
	wantVersion, err := ContentVersion(files, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 8)
	for i := 0; i < cap(results); i++ {
		go func() {
			db, localErr, fatalErr := CompileWithPreparedLocalFiles(files, nil, preparedFiles, Limits{})
			if db == nil || localErr != nil || fatalErr != nil {
				results <- fmt.Errorf("local=%v fatal=%v", localErr, fatalErr)
			} else if db.Version() != wantVersion || db.DetectionCount() != 1 {
				results <- fmt.Errorf("version=%q count=%d", db.Version(), db.DetectionCount())
			} else {
				results <- nil
			}
		}()
	}
	for i := 0; i < cap(results); i++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestPreparedHashListPreservesIgnoreMD5FunctionalityAndPrecedence(t *testing.T) {
	data := []byte("an inert indicator-listed fixture body")
	h := sha256.Sum256(data)
	digest := hex.EncodeToString(h[:])
	wire, prepared, err := PrepareHashList([]byte(digest+"\n"+digest+"\n"), "SHA256", 4096)
	if err != nil {
		t.Fatal(err)
	}
	base := map[string][]byte{"intel.hsb": wire, "other.ndb": []byte("Other.Test:0:*:414243444546\n")}
	for name, local := range map[string]map[string][]byte{
		"matching line MD5":  {"local.ign2": []byte(fmt.Sprintf("UTMStack.Intel.SHA256.%s:%x\n", digest, md5.Sum(bytes.SplitN(wire, []byte("\n"), 2)[0])))},
		"different line MD5": {"local.ign2": []byte(fmt.Sprintf("UTMStack.Intel.SHA256.%s:%x\n", digest, md5.Sum([]byte("different source line"))))},
		"earlier local hash": {"a.hsb": []byte(digest + ":*:Local.First:73\n")},
	} {
		t.Run(name, func(t *testing.T) {
			legacy, legacyLocal, legacyErr := CompileWithLocalFiles(base, local, Limits{})
			native, nativeLocal, nativeErr := CompileWithPreparedLocalFiles(base, local, map[string]*PreparedHashList{"intel.hsb": prepared}, Limits{})
			if legacyErr != nil || nativeErr != nil || legacyLocal != nil || nativeLocal != nil || native == nil || legacy == nil {
				t.Fatalf("legacy=%v/%v native=%v/%v", legacyLocal, legacyErr, nativeLocal, nativeErr)
			}
			if native.Version() != legacy.Version() || !reflect.DeepEqual(native.Report(), legacy.Report()) {
				t.Fatalf("typed/text selection differs: native=%+v legacy=%+v", native.Report(), legacy.Report())
			}
			want := "UTMStack.Intel.SHA256." + digest
			if name == "matching line MD5" {
				want = ""
			}
			if name == "earlier local hash" {
				want = "Local.First"
			}
			assertMatch(t, native, data, want)
		})
	}
	for _, level := range []int{72, 73, FunctionalityLevel} {
		t.Run(fmt.Sprintf("functionality %d", level), func(t *testing.T) {
			legacy, native := newDatabase(Limits{}.defaults(), level), newDatabase(Limits{}.defaults(), level)
			src := source{name: "intel.hsb", ext: ".hsb", data: wire}
			var legacyErrors, nativeErrors []string
			legacy.compileSource(src, func(_ string, _ int, err error) { legacyErrors = append(legacyErrors, err.Error()) })
			native.compilePreparedHashList(src, prepared, func(_ string, _ int, err error) { nativeErrors = append(nativeErrors, err.Error()) })
			if !reflect.DeepEqual(legacyErrors, nativeErrors) || !reflect.DeepEqual(legacy.Report(), native.Report()) || legacy.DetectionCount() != native.DetectionCount() {
				t.Fatalf("functionality changed: legacy=%+v/%v native=%+v/%v", legacy.Report(), legacyErrors, native.Report(), nativeErrors)
			}
		})
	}
	t.Run("invalid locals retain the prepared base only", func(t *testing.T) {
		db, localErr, fatalErr := CompileWithPreparedLocalFiles(base,
			map[string][]byte{"bad.ndb": []byte("Bad.Test:0:*:zz\n"), "local.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")},
			map[string]*PreparedHashList{"intel.hsb": prepared}, Limits{})
		if db == nil || localErr == nil || fatalErr != nil {
			t.Fatalf("local=%v fatal=%v", localErr, fatalErr)
		}
		assertMatch(t, db, data, "UTMStack.Intel.SHA256."+digest)
		assertMatch(t, db, []byte("GHIJKL"), "")
	})
	if db, _, err := CompileWithPreparedLocalFiles(base, nil, map[string]*PreparedHashList{"intel.hsb": prepared}, Limits{MaxSignatures: 1}); err == nil || db != nil {
		t.Fatal("prepared duplicate rows bypassed signature count limit")
	}
}

func BenchmarkPreparedHashListCompile(b *testing.B) {
	var list strings.Builder
	for i := 0; i < 4096; i++ {
		fmt.Fprintf(&list, "%x\n", sha256.Sum256([]byte(fmt.Sprintf("inert hash fixture %d", i))))
	}
	wire, prepared, err := PrepareHashList([]byte(list.String()), "SHA256", 2<<20)
	if err != nil {
		b.Fatal(err)
	}
	files := map[string][]byte{"intel.hsb": wire}
	for _, mode := range []string{"durable-text", "prepared-native"} {
		b.Run(mode, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(wire)))
			for i := 0; i < b.N; i++ {
				var native map[string]*PreparedHashList
				if mode == "prepared-native" {
					native = map[string]*PreparedHashList{"intel.hsb": prepared}
				}
				if _, _, err := CompileWithPreparedLocalFiles(files, nil, native, Limits{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
