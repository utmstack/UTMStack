package usecase

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

// renderReportPDF renders a compliance report to a tabular PDF natively in this
// module (replaces the removed web-pdf service). It is internal to the usecase
// layer — no interface/connector needed.
func renderReportPDF(rep domain.Report) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("") // latin-1
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, tr(rep.FrameworkName+" — Compliance Report"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(110, 110, 110)
	pdf.CellFormat(0, 6, tr("Generated: "+rep.GeneratedAt), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(2)

	// Summary
	s := rep.Summary
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 8, fmt.Sprintf("Compliance score: %d%%", s.CompliantPct), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 6, tr(fmt.Sprintf(
		"Compliant %d  ·  Non-compliant %d  ·  At risk %d  ·  Not covered %d  ·  Pending %d  ·  Out of scope %d",
		s.Compliant, s.NonCompliant, s.AtRisk, s.NotCovered, s.Pending, s.OutOfScope)), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	const (
		wControl = 22.0
		wName    = 78.0
		wStatus  = 28.0
		wEvid    = 52.0
	)

	for _, sec := range rep.Sections {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, tr(sec.Name), "", 1, "L", false, 0, "")

		// header row
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetFillColor(235, 235, 235)
		pdf.CellFormat(wControl, 6, "Control", "1", 0, "L", true, 0, "")
		pdf.CellFormat(wName, 6, "Name", "1", 0, "L", true, 0, "")
		pdf.CellFormat(wStatus, 6, "Status", "1", 0, "L", true, 0, "")
		pdf.CellFormat(wEvid, 6, "Evidence", "1", 1, "L", true, 0, "")

		pdf.SetFont("Helvetica", "", 8)
		for _, row := range sec.Controls {
			pdf.CellFormat(wControl, 6, tr(row.ControlID), "1", 0, "L", false, 0, "")
			pdf.CellFormat(wName, 6, tr(truncate(row.Name, 58)), "1", 0, "L", false, 0, "")
			rc, gc, bc := statusColor(row.Status)
			pdf.SetTextColor(rc, gc, bc)
			pdf.CellFormat(wStatus, 6, tr(row.Status), "1", 0, "L", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.CellFormat(wEvid, 6, tr(truncate(row.Evidence, 38)), "1", 1, "L", false, 0, "")
		}
		pdf.Ln(2)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func statusColor(status string) (int, int, int) {
	switch status {
	case "COMPLIANT":
		return 0, 130, 0
	case "NON_COMPLIANT":
		return 200, 0, 0
	case "AT_RISK":
		return 200, 130, 0
	case "NOT_COVERED":
		return 120, 120, 120
	default: // PENDING, OUT_OF_SCOPE
		return 90, 90, 90
	}
}
