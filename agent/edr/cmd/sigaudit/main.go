// sigaudit measures how much of a signature feed the native compiler can
// represent, rule by rule, and optionally gates a release on approved gaps.
// It never writes a loadable database: an incomplete feed must still fail
// normal compilation rather than being filtered and presented as complete.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// defaultReferenceLevel is the functionality level of the pinned ClamAV
// 1.5.4 reference. An older deployed engine can still be audited explicitly
// with -reference-level; the default must not hide rules selected by 1.5.4.
const defaultReferenceLevel = 234

type options struct {
	Inputs         []string
	ReferenceLevel int
	ReferencePUA   bool
	JSON           string
	Gate           string
	Examples       int
	Top            int
	MaxBytes       int64
	StrictRuntime  bool
}

// Gate lists gaps a release owner has explicitly accepted. Reason matches the
// normalized diagnostic exactly, or as a prefix when it ends in "*". Max caps
// the accepted count; -1 accepts any count.
type Gate struct {
	Approved []Approval `json:"approved"`
}
type Approval struct {
	Format string `json:"format"`
	Reason string `json:"reason"`
	Max    int    `json:"max"`
	Note   string `json:"note,omitempty"`
}

func (a Approval) matches(r signatures.AuditReason) bool {
	if a.Format != "*" && !strings.EqualFold(a.Format, r.Format) {
		return false
	}
	if prefix, ok := strings.CutSuffix(a.Reason, "*"); ok {
		return strings.HasPrefix(r.Reason, prefix)
	}
	return a.Reason == r.Reason
}

// Violation is a rejected-rule reason no approval covers, or one whose count
// exceeds its approval.
type Violation struct {
	Format, Reason string
	Count, Allowed int
}

func evaluateGate(g Gate, r *signatures.AuditReport) []Violation {
	used := make([]int, len(g.Approved))
	var out []Violation
	for _, reason := range r.Reasons {
		covered := false
		for i, a := range g.Approved {
			if !a.matches(reason) {
				continue
			}
			used[i] += reason.Count
			if a.Max >= 0 && used[i] > a.Max {
				out = append(out, Violation{reason.Format, reason.Reason, used[i], a.Max})
			}
			covered = true
			break
		}
		if !covered {
			out = append(out, Violation{reason.Format, reason.Reason, reason.Count, 0})
		}
	}
	for _, e := range r.Errors {
		out = append(out, Violation{"input", e.File + ": " + e.Reason, 1, 0})
	}
	return out
}

func main() {
	var o options
	flag.IntVar(&o.ReferenceLevel, "reference-level", defaultReferenceLevel, "functionality level of the engine being replaced; rules outside its window are not gaps")
	flag.BoolVar(&o.ReferencePUA, "reference-pua", false, "the replaced engine loaded potentially-unwanted-application databases")
	flag.StringVar(&o.JSON, "json", "", "write the full report as JSON to this path")
	flag.StringVar(&o.Gate, "gate", "", "JSON file of approved gaps; exit 2 when any other gap remains")
	flag.IntVar(&o.Examples, "examples", 3, "example rules kept per reason")
	flag.IntVar(&o.Top, "top", 40, "reasons printed in the summary")
	flag.Int64Var(&o.MaxBytes, "max-bytes", 4<<30, "upper bound on audited input and expanded container bytes")
	flag.BoolVar(&o.StrictRuntime, "strict-runtime", false, "fail the release gate if any accepted rule can be unknown when reached")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: sigaudit [flags] <signature file or directory>...\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	o.Inputs = flag.Args()
	code, err := run(o, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sigaudit:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func run(o options, out io.Writer) (int, error) {
	if len(o.Inputs) == 0 {
		return 1, errors.New("no signature inputs given")
	}
	var gate *Gate
	if o.Gate != "" {
		b, err := os.ReadFile(o.Gate)
		if err != nil {
			return 1, err
		}
		gate = &Gate{}
		dec := json.NewDecoder(strings.NewReader(string(b)))
		dec.DisallowUnknownFields()
		if err = dec.Decode(gate); err != nil {
			return 1, fmt.Errorf("invalid gate file: %w", err)
		}
	}
	files, err := readInputs(o.Inputs, o.MaxBytes)
	if err != nil {
		return 1, err
	}
	start := time.Now()
	report, err := signatures.Audit(files, signatures.AuditOptions{
		ReferenceLevel: o.ReferenceLevel,
		ReferencePUA:   o.ReferencePUA,
		Examples:       o.Examples,
		Limits: signatures.Limits{
			MaxFileBytes:  o.MaxBytes,
			MaxTotalBytes: o.MaxBytes,
			MaxFiles:      1 << 20,
			MaxSignatures: 1 << 30,
			MaxLineBytes:  1 << 20,
		},
	})
	elapsed := time.Since(start)
	if err != nil {
		return 1, err
	}
	files = nil
	runtime.GC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	summarize(out, report, o.Top, elapsed, mem.HeapInuse)
	if o.JSON != "" {
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return 1, err
		}
		if err = os.WriteFile(o.JSON, append(b, '\n'), 0o644); err != nil {
			return 1, err
		}
	}
	runtime.KeepAlive(report)
	if gate == nil {
		if report.Complete() && (!o.StrictRuntime || report.Totals.Unknown == 0) {
			return 0, nil
		}
		if o.StrictRuntime && report.Totals.Unknown > 0 {
			fmt.Fprintf(out, "\nstrict runtime gate: FAIL (%d accepted rules can be unknown when reached)\n", report.Totals.Unknown)
		}
		return 2, nil
	}
	violations := evaluateGate(*gate, report)
	if len(violations) == 0 && (!o.StrictRuntime || report.Totals.Unknown == 0) {
		fmt.Fprintln(out, "\nrelease gate: PASS (every remaining gap is approved)")
		return 0, nil
	}
	unknown := 0
	if o.StrictRuntime {
		unknown = report.Totals.Unknown
	}
	fmt.Fprintf(out, "\nrelease gate: FAIL (%d unapproved gap reasons, %d unresolved runtime rules)\n", len(violations), unknown)
	if o.StrictRuntime && report.Totals.Unknown > 0 {
		fmt.Fprintf(out, "  runtime unknown when reached: %d accepted rules\n", report.Totals.Unknown)
	}
	for _, v := range violations {
		fmt.Fprintf(out, "  %s  %q  count %d, approved %d\n", v.Format, v.Reason, v.Count, v.Allowed)
	}
	return 2, nil
}

func readInputs(inputs []string, limit int64) (map[string][]byte, error) {
	files := map[string][]byte{}
	var total int64
	add := func(path string) error {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", path)
		}
		name := filepath.Base(path)
		if _, dup := files[name]; dup {
			return fmt.Errorf("duplicate input name %s", name)
		}
		if total += info.Size(); total > limit {
			return fmt.Errorf("inputs exceed %d bytes", limit)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[name] = b
		return nil
	}
	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if err = add(in); err != nil {
				return nil, err
			}
			continue
		}
		entries, err := os.ReadDir(in)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if err = add(filepath.Join(in, e.Name())); err != nil {
				return nil, err
			}
		}
	}
	return files, nil
}

func summarize(out io.Writer, r *signatures.AuditReport, top int, elapsed time.Duration, heap uint64) {
	pct := func(n, d int) string {
		if d == 0 {
			return "-"
		}
		return fmt.Sprintf("%.1f%%", 100*float64(n)/float64(d))
	}
	fmt.Fprintf(out, "Signature feed audit: engine level %d, reference level %d, reference PUA %v\n", r.EngineLevel, r.ReferenceLevel, r.ReferencePUA)
	fmt.Fprintf(out, "input %d bytes, expanded %d bytes, %d containers, audit time %s, heap after audit %d MiB\n\n",
		r.InputBytes, r.ExpandedBytes, len(r.Containers), elapsed.Round(time.Millisecond), heap>>20)
	tw := tabwriter.NewWriter(out, 0, 4, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "kind\trules\treference loads\taccepted\tunknown when reached\trejected\tcoverage\tref-skipped\taccepted, ref retired\t")
	kinds := make([]string, 0, len(r.ByKind))
	for k := range r.ByKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	row := func(name string, c signatures.AuditCounts) {
		loads := c.Accepted - c.AcceptedOutsideReference + c.Rejected
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\t%s\t%d\t%d\t\n", name, c.Rules, loads, c.Accepted-c.AcceptedOutsideReference, c.Unknown, c.Rejected,
			pct(c.Accepted-c.AcceptedOutsideReference, loads), c.ReferenceSkipped, c.AcceptedOutsideReference)
	}
	for _, k := range kinds {
		row(k, r.ByKind[k])
	}
	row("TOTAL", r.Totals)
	tw.Flush()

	if len(r.SkippedByReason) > 0 {
		reasons := make([]string, 0, len(r.SkippedByReason))
		for reason := range r.SkippedByReason {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)
		fmt.Fprintln(out, "\nreference-skipped rules by reason:")
		for _, reason := range reasons {
			fmt.Fprintf(out, "%10d  %s\n", r.SkippedByReason[reason], reason)
		}
	}
	if len(r.UnknownByReason) > 0 {
		reasons := make([]string, 0, len(r.UnknownByReason))
		for reason := range r.UnknownByReason {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)
		fmt.Fprintln(out, "\naccepted rules loaded as unknown when reached, by the analysis they need:")
		for _, reason := range reasons {
			fmt.Fprintf(out, "%10d  %s\n", r.UnknownByReason[reason], reason)
		}
	}
	if r.DynamicConfig != signatures.DefaultDynamicConfig() {
		fmt.Fprintf(out, "\ndynamic configuration differs from the defaults: %+v\n", r.DynamicConfig)
	}

	fmt.Fprintln(out, "\nper file:")
	tw = tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "file\tkind\trules\taccepted\trejected\tref-skipped\t")
	for _, f := range r.Files {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t\n", f.Name, f.Kind, f.Rules, f.Accepted, f.Rejected, f.ReferenceSkipped)
	}
	tw.Flush()

	if len(r.Reasons) > 0 {
		fmt.Fprintf(out, "\ntop rejection reasons (%d distinct):\n", len(r.Reasons))
		for i, reason := range r.Reasons {
			if i == top {
				fmt.Fprintf(out, "  ... %d more in the JSON report\n", len(r.Reasons)-top)
				break
			}
			fmt.Fprintf(out, "%8d  %-5s %s\n", reason.Count, reason.Format, reason.Reason)
			if f := topFeatures(reason.Features, 6); f != "" {
				fmt.Fprintf(out, "          uses: %s\n", f)
			}
		}
	}
	if len(r.LimitFindings) > 0 {
		fmt.Fprintln(out, "\nproduction limit findings:")
		for _, f := range r.LimitFindings {
			fmt.Fprintln(out, "  "+f)
		}
	}
	if len(r.Errors) > 0 {
		fmt.Fprintln(out, "\ninput errors:")
		for _, e := range r.Errors {
			fmt.Fprintf(out, "  %s: %s\n", e.File, e.Reason)
		}
	}
}

func topFeatures(m map[string]int, n int) string {
	type kv struct {
		k string
		v int
	}
	var all []kv
	for k, v := range m {
		all = append(all, kv{k, v})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].v != all[j].v {
			return all[i].v > all[j].v
		}
		return all[i].k < all[j].k
	})
	var parts []string
	for i, e := range all {
		if i == n {
			break
		}
		parts = append(parts, fmt.Sprintf("%s=%d", e.k, e.v))
	}
	return strings.Join(parts, " ")
}
