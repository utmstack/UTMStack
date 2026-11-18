// edr/netblock/parse.go
package netblock

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ParseAccumulative reads a gzip'd newline-delimited value list. A value with a
// "/" is a CIDR, otherwise an IP. Blank lines and #-comments are ignored.
// Unparseable lines are skipped (a poisoned line must not drop the whole feed).
func ParseAccumulative(r io.Reader, level int) ([]Indicator, error) {
	return ParseAccumulativeTyped(r, level, "ip")
}

// ParseAccumulativeTyped reads a gzip'd newline-delimited value list and parses
// each value as the given artifact type. For "ip" the type is auto-detected per
// line (a "/" => cidr, else ip). For "domain"/"hostname" each line is parsed as
// that type. An unknown type falls back to "ip". Blank lines and #-comments are
// ignored; unparseable lines are skipped (a poisoned line must not drop the feed).
func ParseAccumulativeTyped(r io.Reader, level int, typ string) ([]Indicator, error) {
	inds, _, err := parseAccumulativePolicy(r, level, typ, MaxNetworkFeedExpandedBytes, MaxNetworkFeedIndicators, false)
	return inds, err
}

func parseAccumulativeBounded(r io.Reader, level int, typ string, maxBytes int64, maxRecords int) ([]Indicator, int64, error) {
	return parseAccumulativePolicy(r, level, typ, maxBytes, maxRecords, true)
}

// Live ingestion requires every record to be understood. The public parsing
// helpers retain their tolerant filtering behavior for compatibility with tools.
// Silently discarding a bad row in a complete replacement could remove enforced
// indicators while reporting a corrupted source as fresh.
func parseAccumulativePolicy(r io.Reader, level int, typ string, maxBytes int64, maxRecords int, strict bool) ([]Indicator, int64, error) {
	if maxBytes <= 0 || maxBytes > MaxNetworkFeedExpandedBytes || maxRecords <= 0 || maxRecords > MaxNetworkFeedIndicators {
		return nil, 0, ErrNetworkFeedBudget
	}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, 0, err
	}
	defer zr.Close()
	var out []Indicator
	limited := &io.LimitedReader{R: zr, N: maxBytes + 1}
	sc := bufio.NewScanner(limited)
	sc.Buffer(make([]byte, 0, 64*1024), maxNetworkFeedLineBytes)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var lineType string
		switch typ {
		case "domain", "hostname":
			lineType = typ
		default: // "ip" (and any unknown type) => auto-detect ip/cidr
			lineType = "ip"
			if strings.Contains(line, "/") {
				lineType = "cidr"
			}
		}
		ind, err := ParseIndicator(line, lineType, level)
		if err == nil && strict {
			err = validateFeedName(ind)
		}
		if err == nil {
			if len(out) >= maxRecords {
				return nil, maxBytes + 1 - limited.N, ErrNetworkFeedBudget
			}
			out = append(out, ind)
		} else if strict {
			return nil, maxBytes + 1 - limited.N, fmt.Errorf("network feed contains an invalid indicator: %w", err)
		}
	}
	consumed := maxBytes + 1 - limited.N
	if consumed > maxBytes {
		return nil, consumed, ErrNetworkFeedBudget
	}
	if err := sc.Err(); err != nil {
		return nil, consumed, err
	}
	return out, consumed, nil
}

// DailyOp is one incremental change from the daily ndjson delta.
type DailyOp struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Op    string `json:"op"` // add | del
}

// ParseDaily reads ndjson delta lines. Malformed lines are skipped.
func ParseDaily(r io.Reader) ([]DailyOp, error) {
	ops, _, err := parseDailyPolicy(r, MaxNetworkFeedExpandedBytes, MaxNetworkFeedDailyOps, false)
	return ops, err
}

func parseDailyBounded(r io.Reader, maxBytes int64, maxRecords int) ([]DailyOp, int64, error) {
	return parseDailyPolicy(r, maxBytes, maxRecords, true)
}

func parseDailyPolicy(r io.Reader, maxBytes int64, maxRecords int, strict bool) ([]DailyOp, int64, error) {
	if maxBytes <= 0 || maxBytes > MaxNetworkFeedExpandedBytes || maxRecords <= 0 || maxRecords > MaxNetworkFeedDailyOps {
		return nil, 0, ErrNetworkFeedBudget
	}
	var out []DailyOp
	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	sc := bufio.NewScanner(limited)
	sc.Buffer(make([]byte, 0, 64*1024), maxNetworkFeedLineBytes)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var op DailyOp
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			if strict {
				return nil, maxBytes + 1 - limited.N, fmt.Errorf("network delta contains malformed JSON: %w", err)
			}
			continue
		}
		if op.Value == "" || (op.Op != "add" && op.Op != "del") {
			if strict {
				return nil, maxBytes + 1 - limited.N, fmt.Errorf("network delta contains an invalid operation")
			}
			continue
		}
		if op.Type == "" {
			op.Type = "ip"
		}
		if strict {
			ind, err := ParseIndicator(op.Value, op.Type, 0)
			if err == nil {
				err = validateFeedName(ind)
			}
			if err != nil {
				return nil, maxBytes + 1 - limited.N, fmt.Errorf("network delta contains an invalid indicator: %w", err)
			}
		}
		if len(out) >= maxRecords {
			return nil, maxBytes + 1 - limited.N, ErrNetworkFeedBudget
		}
		out = append(out, op)
	}
	consumed := maxBytes + 1 - limited.N
	if consumed > maxBytes {
		return nil, consumed, ErrNetworkFeedBudget
	}
	if err := sc.Err(); err != nil {
		return nil, consumed, err
	}
	return out, consumed, nil
}

// Canonical feed names use a bounded ASCII DNS spelling, including punycode and
// service underscores. They are suffix indicators, not URL/glob expressions or
// arbitrary binary DNS labels. The RFC1035 wire-size bounds leave at most 253
// printable bytes without the terminal dot; a single terminal dot is already
// removed by ParseIndicator. Case-folded names must have this ASCII spelling;
// general internationalized names require A-labels. This is not IDNA processing.
func validateFeedName(ind Indicator) error {
	if ind.Type != "domain" && ind.Type != "hostname" {
		return nil
	}
	if len(ind.Value) == 0 || len(ind.Value) > 253 {
		return fmt.Errorf("network feed name exceeds supported length")
	}
	for _, label := range strings.Split(ind.Value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("network feed name contains an invalid label")
		}
		for _, c := range []byte(label) {
			if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
				continue
			}
			return fmt.Errorf("network feed name uses an unsupported spelling")
		}
	}
	return nil
}
