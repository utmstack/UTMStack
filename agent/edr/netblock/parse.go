// edr/netblock/parse.go
package netblock

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
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
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var out []Indicator
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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
		if ind, err := ParseIndicator(line, lineType, level); err == nil {
			out = append(out, ind)
		}
	}
	return out, sc.Err()
}

// DailyOp is one incremental change from the daily ndjson delta.
type DailyOp struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Op    string `json:"op"` // add | del
}

// ParseDaily reads ndjson delta lines. Malformed lines are skipped.
func ParseDaily(r io.Reader) ([]DailyOp, error) {
	var out []DailyOp
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var op DailyOp
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			continue
		}
		if op.Value == "" || (op.Op != "add" && op.Op != "del") {
			continue
		}
		if op.Type == "" {
			op.Type = "ip"
		}
		out = append(out, op)
	}
	return out, sc.Err()
}
