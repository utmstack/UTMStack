package usecase

import (
	"encoding/csv"
	"fmt"
	"io"
)

func ExportCSV(hits []map[string]any, columns []string, w io.Writer) error {
	cw := csv.NewWriter(w)

	// Header row.
	if err := cw.Write(columns); err != nil {
		return fmt.Errorf("csv write header: %w", err)
	}

	// Data rows.
	row := make([]string, len(columns))
	for _, hit := range hits {
		for i, col := range columns {
			if v, ok := hit[col]; ok && v != nil {
				row[i] = fmt.Sprintf("%v", v)
			} else {
				row[i] = ""
			}
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("csv write row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
