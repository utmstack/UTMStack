package signatures

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/unpack"
)

// TestUPXOracle replays the reference's UPX step on the programs its debug
// log names (EDR_UPX_ORACLE, a list of logs separated by commas; each log
// holds the "Checking realpath", "cli_scanpe: UPX" and "Decompressed data
// saved in" lines of a run with --leave-temps; relative paths are relative
// to the log's folder). A decided outcome must be the reference's, and a
// rebuilt program must equal the one the reference saved, byte for byte.
func TestUPXOracle(t *testing.T) {
	logs := os.Getenv("EDR_UPX_ORACLE")
	if logs == "" {
		t.Skip("EDR_UPX_ORACLE not set")
	}
	reFile := regexp.MustCompile(`Checking realpath of (\S+)`)
	reSaved := regexp.MustCompile(`Decompressed data saved in (\S+)`)
	type record struct {
		path, saved             string
		failed, mismatch, tried bool
		routine                 string // the decompressor routine the reference recognized
	}
	reRoutine := regexp.MustCompile(`Looks like a (NRV2[BDE]) decompression routine`)
	counts := map[string]int{}
	var unknown []string
	for _, logPath := range strings.Split(logs, ",") {
		f, err := os.Open(logPath)
		if err != nil {
			t.Fatal(err)
		}
		var recs []*record
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if m := reFile.FindStringSubmatch(line); m != nil {
				p := m[1]
				if !filepath.IsAbs(p) {
					p = filepath.Join(filepath.Dir(logPath), p)
				}
				recs = append(recs, &record{path: p})
				continue
			}
			if len(recs) == 0 {
				continue
			}
			r := recs[len(recs)-1]
			switch {
			case reSaved.MatchString(line):
				r.saved = reSaved.FindStringSubmatch(line)[1]
			case strings.Contains(line, "All decompressors failed"):
				r.failed = true
			case strings.Contains(line, "Size mismatch"):
				r.mismatch = true
			case reRoutine.MatchString(line):
				r.routine = reRoutine.FindStringSubmatch(line)[1]
			}
			if strings.Contains(line, "empty section found") {
				r.tried = true
			}
		}
		f.Close()
		for _, r := range recs {
			if !r.tried {
				continue
			}
			data, err := os.ReadFile(r.path)
			if err != nil {
				t.Fatal(err)
			}
			h, st := readHookHeader(data, true)
			if st != hookParsed {
				counts["header not parsed"]++
				continue
			}
			index := -1
			for i := 0; i+1 < len(h.sections); i++ {
				s, next := h.sections[i], h.sections[i+1]
				if s.size == 0 && s.vsz != 0 && next.size != 0 && next.vsz != 0 {
					index = i
					break
				}
			}
			if index < 0 {
				t.Errorf("%s: no empty section here, the reference found one", r.path)
				continue
			}
			routine := ""
			if m, ok := unpack.UPXRoutine(data[h.entry : h.entry+h.epBytes]); ok {
				routine = m.String()
			}
			if routine != r.routine {
				t.Errorf("%s: routine %q, the reference recognized %q", r.path, routine, r.routine)
			}
			outcome, out := h.upx(data, index)
			switch {
			case outcome == unpack.UPXUnknown:
				counts["unknown"]++
				unknown = append(unknown, r.path)
			case r.saved != "":
				ref, err := os.ReadFile(r.saved)
				if err != nil {
					t.Fatal(err)
				}
				if outcome != unpack.UPXUnpacked || !bytes.Equal(out, ref) {
					t.Errorf("%s: outcome %d, %d bytes; the reference rebuilt %d bytes", r.path, outcome, len(out), len(ref))
					continue
				}
				counts["rebuilt, identical"]++
			case r.failed:
				if outcome != unpack.UPXFailed {
					t.Errorf("%s: outcome %d, the reference failed", r.path, outcome)
					continue
				}
				counts["failed"]++
			case r.mismatch:
				if outcome != unpack.UPXReturns {
					t.Errorf("%s: outcome %d, the reference reported a size mismatch", r.path, outcome)
					continue
				}
				counts["size mismatch"]++
			default:
				counts["reference outcome not in the log"]++
			}
		}
	}
	t.Logf("%v", counts)
	if len(unknown) > 0 {
		t.Logf("unknown: %s", strings.Join(unknown[:min(len(unknown), 20)], " "))
	}
}

// TestMEWOracle replays the reference's MEW step through the parser model
// on the programs in EDR_MEW_ORACLE (a folder: each X.exe next to
// X.exe.rebuilt, the program the reference rebuilt, or X.exe.log, its debug
// log, when it rebuilt none; docs/2026-10-14-unpackers/fsg2_pack.py builds
// them). A rebuilt program must equal the reference's byte for byte; where
// the reference's unpacking failed, the parser must end without a program.
func TestMEWOracle(t *testing.T) {
	dir := os.Getenv("EDR_MEW_ORACLE")
	if dir == "" {
		t.Skip("EDR_MEW_ORACLE not set")
	}
	programs, err := filepath.Glob(filepath.Join(dir, "*.exe"))
	if err != nil || len(programs) == 0 {
		t.Fatal(err, len(programs))
	}
	counts := map[string]int{}
	for _, path := range programs {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		h, st := readHookHeader(data, true)
		if st != hookParsed {
			t.Errorf("%s: header %v", path, st)
			continue
		}
		reach, res := h.unpackerHookReach(data, 0xfffff&^0x80)
		want, err := os.ReadFile(path + ".rebuilt")
		switch {
		case err == nil:
			if reach != reachNever || !res.tried || !bytes.Equal(res.child, want) {
				t.Errorf("%s: reach %v tried %v, %d bytes rebuilt, the reference %d", path, reach, res.tried, len(res.child), len(want))
			}
			counts["rebuilt"]++
		case reach == reachUnknown:
			counts["unknown"]++
		default:
			if reach != reachNever || res.tried {
				t.Errorf("%s: reach %v tried %v where the reference's unpacking failed", path, reach, res.tried)
			}
			counts["failed"]++
		}
	}
	t.Logf("%d programs: %v", len(programs), counts)
}

// TestWWPackOracle is TestMEWOracle for the WWPack step (EDR_WWPACK_ORACLE;
// docs/2026-10-14-unpackers/wwpack_pack.py builds the variants).
func TestWWPackOracle(t *testing.T) {
	dir := os.Getenv("EDR_WWPACK_ORACLE")
	if dir == "" {
		t.Skip("EDR_WWPACK_ORACLE not set")
	}
	programs, err := filepath.Glob(filepath.Join(dir, "*.exe"))
	if err != nil || len(programs) == 0 {
		t.Fatal(err, len(programs))
	}
	counts := map[string]int{}
	for _, path := range programs {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		h, st := readHookHeader(data, true)
		if st != hookParsed {
			t.Errorf("%s: header %v", path, st)
			continue
		}
		reach, res := h.unpackerHookReach(data, 0xfffff&^0x80)
		want, err := os.ReadFile(path + ".rebuilt")
		switch {
		case reach == reachUnknown:
			counts["unknown"]++
		case err == nil:
			if reach != reachNever || !res.tried || !bytes.Equal(res.child, want) {
				t.Errorf("%s: reach %v tried %v, %d bytes rebuilt, the reference %d", path, reach, res.tried, len(res.child), len(want))
			}
			counts["rebuilt"]++
		default:
			if reach != reachNever || res.tried {
				t.Errorf("%s: reach %v tried %v where the reference's unpacking failed", path, reach, res.tried)
			}
			counts["failed"]++
		}
	}
	t.Logf("%d programs: %v", len(programs), counts)
}

// TestPetiteOracle checks the PEtite step against the reference's runs over
// EDR_PETITE_ORACLE: each program with a .rebuilt file (the reference's
// rebuilt program) must give that program, each whose .log says unpacking
// failed must go on to the hooks, and the rest of the reference's lines are
// in .log for the lab model (docs/2026-10-14-unpackers/petite_model.py).
func TestPetiteOracle(t *testing.T) {
	dir := os.Getenv("EDR_PETITE_ORACLE")
	if dir == "" {
		t.Skip("EDR_PETITE_ORACLE not set")
	}
	programs, err := filepath.Glob(filepath.Join(dir, "*.exe"))
	if err != nil || len(programs) == 0 {
		t.Fatal(err, len(programs))
	}
	counts := map[string]int{}
	var unknown []string
	for _, path := range programs {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		h, st := readHookHeader(data, false)
		if st != hookParsed {
			t.Errorf("%s: header %v", path, st)
			continue
		}
		reach, res := h.unpackerHookReach(data, 0xfffff&^0x80)
		want, err := os.ReadFile(path + ".rebuilt")
		log, _ := os.ReadFile(path + ".log")
		switch {
		case reach == reachUnknown:
			counts["unknown"]++
			unknown = append(unknown, filepath.Base(path))
		case err == nil:
			if reach != reachNever || res.outcome != unpack.UPXUnpacked || !bytes.Equal(res.child, want) {
				t.Errorf("%s: reach %v outcome %v, %d bytes rebuilt, the reference %d", path, reach, res.outcome, len(res.child), len(want))
			}
			counts["rebuilt"]++
		case bytes.Contains(log, []byte("Petite: Unpacking failed")):
			if reach != reachHooks || res.tried && res.outcome == unpack.UPXUnpacked {
				t.Errorf("%s: reach %v outcome %v where the reference's unpacking failed", path, reach, res.outcome)
			}
			counts["failed"]++
		default:
			// Only the reference's lines: its entry point (0 where it
			// prints none)
			if !bytes.Contains(log, []byte("Petite: Sections dump:")) {
				counts["no result"]++
				continue
			}
			want := uint64(0)
			if m := regexp.MustCompile(`Petite: Old EP: ([0-9a-f]+)|using bogus (1000)`).FindSubmatch(log); m != nil {
				want, _ = strconv.ParseUint(string(m[1])+string(m[2]), 16, 32)
			}
			if reach != reachNever || res.outcome != unpack.UPXUnpacked || len(res.child) < 0x100 {
				t.Errorf("%s: reach %v outcome %v where the reference unpacked", path, reach, res.outcome)
				continue
			}
			le := binary.LittleEndian
			if got := le.Uint32(res.child[0xd0+40:]); uint64(got) != want {
				t.Errorf("%s: entry point %#x, the reference %#x", path, got, want)
			}
			// and its sections dump: address, virtual size, raw size.
			dump := regexp.MustCompile(`\.SECT\d+ RVA:([0-9a-f]+) VSize:([0-9a-f]+) ROffset: [0-9a-f]+, RSize:([0-9a-f]+)`).FindAllSubmatch(log, -1)
			if int(le.Uint16(res.child[0xd0+6:])) != len(dump) {
				t.Errorf("%s: %d sections, the reference %d", path, le.Uint16(res.child[0xd0+6:]), len(dump))
				continue
			}
			for k, d := range dump {
				h := res.child[0xd0+24+0xe0+40*k:]
				for f, at := range []int{12, 8, 16} {
					if v, _ := strconv.ParseUint(string(d[f+1]), 16, 32); uint64(le.Uint32(h[at:])) != v {
						t.Errorf("%s: section %d field %d %#x, the reference %#x", path, k, at, le.Uint32(h[at:]), v)
					}
				}
			}
			counts["entry point and sections"]++
		}
	}
	t.Logf("%d programs: %v; not decided: %s", len(programs), counts, strings.Join(unknown, " "))
}

// TestYCOracle, TestASPackOracle and TestUpackOracle check those steps against
// the reference's runs (refscan.py output: FILE.log and, where it rebuilt a
// program, FILE.rebuilt): a program with a .rebuilt file must give that
// program, any other must go on to the hooks.
func TestYCOracle(t *testing.T)     { unpackerOracle(t, "EDR_YC_ORACLE") }
func TestASPackOracle(t *testing.T) { unpackerOracle(t, "EDR_ASPACK_ORACLE") }
func TestUpackOracle(t *testing.T)  { unpackerOracle(t, "EDR_UPACK_ORACLE") }

// TestPESpinOracle replays the reference's PESpin results on variants of
// the upstream sample (docs/2026-10-15-pespin/): byte inversions, the same
// with the stored key adjusted for the checksum, and re-encrypted builds.
func TestPESpinOracle(t *testing.T) { unpackerOracle(t, "EDR_PESPIN_ORACLE") }

func unpackerOracle(t *testing.T, env string) {
	dir := os.Getenv(env)
	if dir == "" {
		t.Skip(env + " not set")
	}
	programs, err := filepath.Glob(filepath.Join(dir, "*.exe"))
	if err != nil || len(programs) == 0 {
		t.Fatal(err, len(programs))
	}
	counts := map[string]int{}
	var unknown []string
	for _, path := range programs {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		log, err := os.ReadFile(path + ".log")
		if err != nil {
			continue
		}
		h, st := readHookHeader(data, false)
		if st != hookParsed {
			counts["header not read"]++
			if bytes.Contains(log, []byte("REBUILT")) {
				t.Errorf("%s: header %v where the reference unpacked", path, st)
			}
			continue
		}
		reach, res := h.unpackerHookReach(data, 0xfffff&^0x80)
		want, err := os.ReadFile(path + ".rebuilt")
		switch {
		case reach == reachUnknown:
			counts["unknown"]++
			unknown = append(unknown, filepath.Base(path))
		case err == nil:
			if reach != reachNever || res.outcome != unpack.UPXUnpacked || !bytes.Equal(res.child, want) {
				t.Errorf("%s: reach %v outcome %v, %d bytes rebuilt, the reference %d", path, reach, res.outcome, len(res.child), len(want))
			}
			counts["rebuilt"]++
		default:
			if reach != reachHooks || res.tried && res.outcome == unpack.UPXUnpacked {
				t.Errorf("%s: reach %v outcome %v where the reference went on", path, reach, res.outcome)
			}
			counts["goes on"]++
		}
	}
	if len(unknown) > 40 {
		unknown = append(unknown[:40], "...")
	}
	t.Logf("%d programs: %v; not decided: %s", len(programs), counts, strings.Join(unknown, " "))
}
