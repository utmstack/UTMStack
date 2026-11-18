package signatures

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// hookLogRecord is one file of the reference's debug log, as the workspace
// script hooklog.py writes it: the hook points reached (259 all PE, 257 PE
// unpacker), the hook programs run as [point, program file, return value],
// the built-in unpackers attempted and the verdict.
type hookLogRecord struct {
	Path      string   `json:"path"`
	Points    []int    `json:"points"`
	Runs      [][3]any `json:"runs"`
	Unpackers []string `json:"unpackers"`
	Verdict   string   `json:"verdict"`
	// Trusted: the reference verified an Authenticode signature against the
	// feed's certificate rules and stopped before its parser.
	Trusted bool `json:"trusted"`
}

// TestHookOracle compares the hook phase with the reference's debug output on
// the files the log names. It runs where those files and the databases are:
// EDR_HOOK_ORACLE names the log, EDR_OFFICIAL_SIGNATURE_DIR the databases.
// Only files scanned as one layer are compared (a nested object's hook
// points cannot be told apart in the log). Every difference is reported;
// answers this engine leaves unknown where the reference tried an unpacker
// are counted, not differences.
func TestHookOracle(t *testing.T) {
	logPath, dir := os.Getenv("EDR_HOOK_ORACLE"), os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if logPath == "" || dir == "" {
		t.Skip("EDR_HOOK_ORACLE or EDR_OFFICIAL_SIGNATURE_DIR not set")
	}
	d, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	programs := officialProgramNames(t, dir) // program file -> trigger name
	f, err := os.Open(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	counts := map[string]int{}
	examples := map[string][]string{}
	note := func(kind, path string) {
		counts[kind]++
		if len(examples[kind]) < 5 {
			examples[kind] = append(examples[kind], path)
		}
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var rec hookLogRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			t.Fatal(err)
		}
		points := fmt.Sprint(rec.Points)
		if points != "[]" && points != "[259]" && points != "[259 257]" {
			note("skipped: nested layers", rec.Path)
			continue
		}
		data, err := os.ReadFile(rec.Path)
		if err != nil {
			t.Fatal(err)
		}
		note("compared", rec.Path)
		refDetects := strings.HasSuffix(rec.Verdict, " FOUND") && !strings.Contains(rec.Verdict, "Heuristics.")
		typ := d.rawScanType(data)
		var runs []string
		w := &matchWork{ctx: context.Background(), limit: d.limits.matchLimit(len(data)),
			hookTrace: func(name string, kind uint16, value uint64, cause error) {
				run := fmt.Sprintf("%d %s %d false", kind, name, value)
				if cause != nil {
					run = fmt.Sprintf("%d %s failed: %v", kind, name, cause)
				}
				runs = append(runs, run)
			}}
		// Verdicts of whole files are compared by the dry-run comparison:
		// this matching has no extraction, so it cannot detect what the
		// reference finds inside containers. Only a detection the reference
		// does not make is a difference here.
		res, _ := d.matchObject(w, data, MatchContext{Type: typ})
		if res.Name != "" && !refDetects {
			note("difference: detection the reference does not make", rec.Path)
		}
		// The hook points.
		h, st := readHookHeader(data, true)
		ours259 := targetOf(typ) == 1 && st == hookParsed && !h.epReadFail && d.dconf.PE != 0
		ref259 := slices.Contains(rec.Points, 259)
		ref257 := slices.Contains(rec.Points, 257)
		if refDetects || res.Name != "" {
			continue // a detection in the rule pass ends the object before the parser
		}
		if rec.Trusted {
			// This engine does not verify signatures (certificate trust rules
			// are counted, not applied), so it runs the hooks the reference
			// skips on a trusted file; only a detection would differ, and that
			// is checked above.
			note("known: trusted by the reference's certificate rules", rec.Path)
			continue
		}
		switch {
		case ref259 && !ours259 && st == hookUnknown:
			note("unknown: header", rec.Path)
			continue
		case ref259 != ours259:
			note(fmt.Sprintf("difference: all-PE point reference %v ours %v (status %d)", ref259, ours259, st), rec.Path)
			continue
		case !ours259:
			continue
		}
		reach, _ := h.unpackerHookReach(data, d.dconf.PE)
		switch {
		case reach == reachUnknown && len(rec.Unpackers) > 0:
			note(fmt.Sprintf("unknown: unpacker tried (%s), reference reached the hooks %v", strings.Join(rec.Unpackers, ","), ref257), rec.Path)
		case reach == reachUnknown:
			note("superset: unknown, reference tried no unpacker", rec.Path)
		case reach == reachHooks && !ref257:
			note("difference: unpacker point reached by us only", rec.Path)
		case reach == reachNever && ref257:
			note("difference: unpacker point reached by the reference only", rec.Path)
		}
		// The programs run.
		// Where this engine does not run the unpacker hooks (never reached,
		// or unknown), only the reference's all-PE runs are expected.
		var want []string
		for _, run := range rec.Runs {
			point, _ := run[0].(float64)
			file, _ := run[1].(string)
			value, _ := run[2].(float64)
			if name, ok := programs[file]; ok && d.hookKind(name) != 0 && (point == 259 || reach == reachHooks) {
				want = append(want, fmt.Sprintf("%d %s %d false", int(point), name, int(value)))
			}
		}
		sort.Strings(want)
		sort.Strings(runs)
		if ours := strings.Join(runs, "; "); ours != strings.Join(want, "; ") {
			note(fmt.Sprintf("difference: runs reference [%s] ours [%s]", strings.Join(want, "; "), ours), rec.Path)
		} else if len(want) > 0 {
			note("same hook runs", rec.Path)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	kinds := make([]string, 0, len(counts))
	for k := range counts {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	failed := false
	for _, k := range kinds {
		t.Logf("%6d  %s  e.g. %s", counts[k], k, strings.Join(examples[k], ", "))
		failed = failed || strings.HasPrefix(k, "difference")
	}
	if failed {
		t.Fail()
	}
}

// hookKind is the hook kind of the loaded hook rule with this trigger name.
func (d *Database) hookKind(name string) uint16 {
	for i := range d.rules {
		if d.rules[i].hook != 0 && d.rules[i].name == name {
			return d.rules[i].hook
		}
	}
	return 0
}

// officialProgramNames maps each program file of bytecode.cvd to its trigger
// name, as the rules carry it.
func officialProgramNames(t *testing.T, dir string) map[string]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "bytecode.cvd"))
	if err != nil {
		t.Fatal(err)
	}
	budget := archiveBudget{limits: Limits{}.defaults(), total: int64(len(b)), files: 1}
	members, _, err := unpackContainer("bytecode.cvd", b, &budget)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]string{}
	for _, m := range members {
		p, err := bytecode.Parse(m.data)
		if err != nil || p.LogicalSignature == "" {
			continue
		}
		names[filepath.Base(m.name)] = strings.SplitN(p.LogicalSignature, ";", 2)[0]
	}
	return names
}
