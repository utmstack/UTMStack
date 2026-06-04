package usecase

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

// ExportCSV writes hits as CSV using the column definitions, mirroring the legacy
// UtilCsv.prepareToDownload behaviour:
//   - header is the column label (falls back to the .keyword-stripped field),
//   - the field is looked up with its ".keyword" suffix stripped and resolved
//     through nested objects via its dotted path,
//   - values are formatted per column type (dates), and lists/objects are
//     flattened to comma-joined / JSON strings; strings have newlines/tabs
//     collapsed to spaces.
func ExportCSV(hits []map[string]any, columns []dto.DataColumn, w io.Writer) error {
	cw := csv.NewWriter(w)

	header := make([]string, len(columns))
	for i, col := range columns {
		header[i] = csvHeader(col)
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("csv write header: %w", err)
	}

	row := make([]string, len(columns))
	for _, hit := range hits {
		for i, col := range columns {
			row[i] = csvCell(hit, col)
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("csv write row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}

func csvHeader(col dto.DataColumn) string {
	if strings.TrimSpace(col.Label) != "" {
		return col.Label
	}
	return stripKeyword(col.Field)
}

func csvCell(hit map[string]any, col dto.DataColumn) string {
	v, ok := lookupField(hit, stripKeyword(col.Field))
	if !ok || v == nil {
		return ""
	}
	return formatValue(v, col.Type)
}

func stripKeyword(field string) string {
	return strings.TrimSuffix(field, ".keyword")
}

// lookupField resolves a field path against the document. It first tries an exact
// key (covers sources that store dotted field names flat), then walks the dotted
// path through nested objects (e.g. "source.ip" -> hit["source"].(map)["ip"]).
func lookupField(hit map[string]any, path string) (any, bool) {
	if v, ok := hit[path]; ok {
		return v, true
	}
	parts := strings.Split(path, ".")
	var cur any = hit
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

const csvDateLayout = "2006-01-02 15:04:05 MST"

func formatValue(v any, colType string) string {
	if strings.EqualFold(colType, "date") {
		if s, ok := formatDate(v); ok {
			return s
		}
	}

	switch val := v.(type) {
	case string:
		return sanitize(val)
	case float64:
		// JSON decodes all numbers to float64; render integers without the
		// scientific notation %v would produce for large values.
		if val == math.Trunc(val) && !math.IsInf(val, 0) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case json.Number:
		return val.String()
	case []any:
		parts := make([]string, 0, len(val))
		for _, e := range val {
			parts = append(parts, fmt.Sprintf("%v", e))
		}
		return strings.Join(parts, ",")
	case map[string]any:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatDate renders a date column in UTC. ES/OS stores dates either as ISO-8601
// strings or epoch milliseconds; both are handled. (Legacy used the configured
// app timezone; the Go stack has no such config yet, so UTC is used.)
func formatDate(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
		} {
			if parsed, err := time.Parse(layout, t); err == nil {
				return parsed.UTC().Format(csvDateLayout), true
			}
		}
		return "", false
	case float64:
		return time.UnixMilli(int64(t)).UTC().Format(csvDateLayout), true
	case json.Number:
		if ms, err := t.Int64(); err == nil {
			return time.UnixMilli(ms).UTC().Format(csvDateLayout), true
		}
	}
	return "", false
}

func sanitize(s string) string {
	r := strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	return r.Replace(s)
}
