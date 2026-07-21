package csvio

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/plugins/enrichment/internal/registry"
)

func LoadCSV(r io.Reader, separator rune, sizeBytes int64) (*registry.Dataset, error) {
	reader := csv.NewReader(r)
	reader.Comma = separator
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv parse: %w", err)
	}

	var nonEmpty [][]string
	for _, rec := range records {
		if len(rec) > 0 {
			nonEmpty = append(nonEmpty, rec)
		}
	}

	if len(nonEmpty) < 1 {
		return nil, errors.New("empty CSV: no header line found")
	}

	headers := make([]string, len(nonEmpty[0]))
	for i, h := range nonEmpty[0] {
		headers[i] = strings.TrimSpace(h)
	}

	colIndex := make(map[string]int, len(headers))
	for i, h := range headers {
		colIndex[h] = i
	}

	rawRows := make([][]string, 0, len(nonEmpty)-1)
	for _, rec := range nonEmpty[1:] {
		row := make([]string, len(rec))
		for i, v := range rec {
			row[i] = strings.TrimSpace(v)
		}
		rawRows = append(rawRows, row)
	}

	return &registry.Dataset{
		Separator:  separator,
		Headers:    headers,
		ColIndex:   colIndex,
		RawRows:    rawRows,
		Indices:    make(map[string]map[string][]int),
		UploadedAt: time.Now().UTC(),
		SizeBytes:  sizeBytes,
		RowCount:   len(rawRows),
	}, nil
}
