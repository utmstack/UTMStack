package usecase

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type rgb [3]int

var statusOrder = []string{
	domain.StatusCompliant, domain.StatusNonCompliant, domain.StatusAtRisk,
	domain.StatusNotCovered, domain.StatusPending, domain.StatusOutOfScope,
}

func statusRGB(status string) rgb {
	switch status {
	case domain.StatusCompliant:
		return rgb{16, 185, 129}
	case domain.StatusNonCompliant:
		return rgb{239, 68, 68}
	case domain.StatusAtRisk:
		return rgb{245, 158, 11}
	case domain.StatusNotCovered:
		return rgb{161, 161, 170}
	case domain.StatusOutOfScope:
		return rgb{167, 139, 250}
	default: // PENDING
		return rgb{56, 189, 248}
	}
}

func statusLabel(status string) string {
	switch status {
	case domain.StatusCompliant:
		return "Compliant"
	case domain.StatusNonCompliant:
		return "Non-compliant"
	case domain.StatusAtRisk:
		return "At risk"
	case domain.StatusNotCovered:
		return "Not covered"
	case domain.StatusOutOfScope:
		return "Out of scope"
	default:
		return "Pending"
	}
}

func scoreRGB(score int) rgb {
	switch {
	case score >= 80:
		return rgb{16, 185, 129}
	case score >= 50:
		return rgb{245, 158, 11}
	default:
		return rgb{239, 68, 68}
	}
}

func parseHex(h string, def rgb) rgb {
	h = strings.TrimPrefix(strings.TrimSpace(h), "#")
	if len(h) != 6 {
		return def
	}
	var r, g, b int
	if _, err := fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b); err != nil {
		return def
	}
	return rgb{r, g, b}
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

func sectionPct(controls []domain.ReportControlRow) int {
	rel, ok := 0, 0
	for _, c := range controls {
		switch c.Status {
		case domain.StatusCompliant:
			rel++
			ok++
		case domain.StatusNonCompliant, domain.StatusAtRisk:
			rel++
		}
	}
	if rel == 0 {
		return 0
	}
	return int(math.Round(float64(ok) / float64(rel) * 100))
}

func renderReportPDF(rep domain.Report, brand connectors.ReportBrand) ([]byte, error) {
	if brand.Name == "" {
		brand.Name = "UTMStack"
	}
	accent := parseHex(brand.AccentHex, rgb{99, 102, 241})

	pdf := fpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pw := 210.0
	pdf.SetMargins(18, 18, 18)
	pdf.SetAutoPageBreak(true, 22)

	pdf.SetFooterFunc(func() {
		if pdf.PageNo() <= 1 {
			return
		}
		pdf.SetY(-14)
		pdf.SetFont("Helvetica", "", 7)
		setText(pdf, rgb{150, 150, 155})
		pdf.CellFormat(160, 6, tr(fmt.Sprintf("%s  ·  %s  ·  %s", brand.Name, rep.FrameworkName, rep.GeneratedAt)), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, fmt.Sprintf("%d", pdf.PageNo()-1), "", 0, "R", false, 0, "")
	})

	// ── Cover ──────────────────────────────────────────────────────────────
	pdf.AddPage()
	setFill(pdf, accent)
	pdf.Rect(0, 0, pw, 6, "F")

	cx := pw / 2
	logoBottom := 30.0
	if drawLogo(pdf, brand.LogoPath, cx, 28, 60, 22) {
		logoBottom = 52
	} else {
		pdf.SetY(34)
		pdf.SetFont("Helvetica", "B", 16)
		setText(pdf, accent)
		pdf.CellFormat(0, 8, tr(strings.ToUpper(brand.Name)), "", 1, "C", false, 0, "")
		logoBottom = 46
	}

	pdf.SetY(logoBottom + 14)
	pdf.SetFont("Helvetica", "B", 9)
	setText(pdf, rgb{160, 160, 165})
	pdf.CellFormat(0, 6, "C O M P L I A N C E   R E P O R T", "", 1, "C", false, 0, "")
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 26)
	setText(pdf, rgb{24, 24, 27})
	pdf.MultiCell(0, 11, tr(rep.FrameworkName), "", "C", false)

	gaugeY := pdf.GetY() + 38
	sc := rep.Summary.CompliantPct
	drawRing(pdf, cx, gaugeY, 26, 7, rgb{238, 238, 240}, scoreRGB(sc), float64(sc)/100)
	pdf.SetXY(cx-26, gaugeY-7)
	pdf.SetFont("Helvetica", "B", 22)
	setText(pdf, rgb{24, 24, 27})
	pdf.CellFormat(52, 10, fmt.Sprintf("%d%%", sc), "", 0, "C", false, 0, "")
	pdf.SetXY(cx-26, gaugeY+4)
	pdf.SetFont("Helvetica", "", 7)
	setText(pdf, rgb{160, 160, 165})
	pdf.CellFormat(52, 5, "SCORE", "", 0, "C", false, 0, "")

	pillY := gaugeY + 34
	drawPill(pdf, cx, pillY, tr(statusLabel(scoreStatus(sc))), scoreRGB(sc))

	total := rep.Summary.Compliant + rep.Summary.NonCompliant + rep.Summary.AtRisk + rep.Summary.NotCovered + rep.Summary.Pending + rep.Summary.OutOfScope
	pdf.SetY(pillY + 14)
	pdf.SetFont("Helvetica", "", 9)
	setText(pdf, rgb{120, 120, 125})
	pdf.CellFormat(0, 5, tr("Generated "+rep.GeneratedAt), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("%d controls evaluated", total), "", 1, "C", false, 0, "")

	// ── Executive summary ──────────────────────────────────────────────────
	pdf.AddPage()
	sectionHeading(pdf, tr, "01", "Executive Summary", accent)
	pdf.Ln(6)

	donutY := pdf.GetY() + 28
	donutX := 50.0
	counts := map[string]int{
		domain.StatusCompliant: rep.Summary.Compliant, domain.StatusNonCompliant: rep.Summary.NonCompliant,
		domain.StatusAtRisk: rep.Summary.AtRisk, domain.StatusNotCovered: rep.Summary.NotCovered,
		domain.StatusPending: rep.Summary.Pending, domain.StatusOutOfScope: rep.Summary.OutOfScope,
	}
	drawDonut(pdf, donutX, donutY, 24, 9, counts, total)
	pdf.SetXY(donutX-24, donutY-5)
	pdf.SetFont("Helvetica", "B", 16)
	setText(pdf, rgb{24, 24, 27})
	pdf.CellFormat(48, 8, fmt.Sprintf("%d", total), "", 0, "C", false, 0, "")
	pdf.SetXY(donutX-24, donutY+3)
	pdf.SetFont("Helvetica", "", 6)
	setText(pdf, rgb{160, 160, 165})
	pdf.CellFormat(48, 4, "CONTROLS", "", 0, "C", false, 0, "")

	lx := 95.0
	ly := donutY - 24
	for _, st := range statusOrder {
		setFill(pdf, statusRGB(st))
		pdf.Rect(lx, ly+1.4, 3, 3, "F")
		setText(pdf, rgb{60, 60, 65})
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetXY(lx+5, ly)
		pdf.CellFormat(55, 6, tr(statusLabel(st)), "", 0, "L", false, 0, "")
		setText(pdf, rgb{24, 24, 27})
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(14, 6, fmt.Sprintf("%d", counts[st]), "", 0, "R", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		setText(pdf, rgb{160, 160, 165})
		p := 0
		if total > 0 {
			p = int(math.Round(float64(counts[st]) / float64(total) * 100))
		}
		pdf.CellFormat(12, 6, fmt.Sprintf("%d%%", p), "", 0, "R", false, 0, "")
		ly += 8
	}

	// ── Findings by section ────────────────────────────────────────────────
	pdf.SetY(donutY + 34)
	sectionHeading(pdf, tr, "02", "Findings by Section", accent)
	pdf.Ln(4)

	for _, sec := range rep.Sections {
		renderSection(pdf, tr, sec)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderSection(pdf *fpdf.Fpdf, tr func(string) string, sec domain.ReportSection) {
	pct := sectionPct(sec.Controls)
	if pdf.GetY() > 250 {
		pdf.AddPage()
	}
	pdf.SetFont("Helvetica", "B", 11)
	setText(pdf, rgb{30, 30, 35})
	pdf.CellFormat(120, 7, tr(sec.Name), "", 0, "L", false, 0, "")
	barX, barW := 150.0, 36.0
	by := pdf.GetY() + 2.5
	setFill(pdf, rgb{238, 238, 240})
	pdf.RoundedRect(barX, by, barW, 2.5, 1.25, "1234", "F")
	setFill(pdf, scoreRGB(pct))
	if w := barW * float64(pct) / 100; w > 0 {
		pdf.RoundedRect(barX, by, w, 2.5, 1.25, "1234", "F")
	}
	setText(pdf, rgb{90, 90, 95})
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetXY(barX+barW+2, pdf.GetY())
	pdf.CellFormat(10, 7, fmt.Sprintf("%d%%", pct), "", 1, "R", false, 0, "")

	const (
		wCtrl = 24.0
		wName = 62.0
		wStat = 26.0
		wEvid = 62.0
	)
	pdf.SetFont("Helvetica", "B", 7)
	setFill(pdf, rgb{245, 245, 247})
	setText(pdf, rgb{120, 120, 125})
	pdf.CellFormat(wCtrl, 6, "CONTROL", "", 0, "L", true, 0, "")
	pdf.CellFormat(wName, 6, "NAME", "", 0, "L", true, 0, "")
	pdf.CellFormat(wStat, 6, "STATUS", "", 0, "L", true, 0, "")
	pdf.CellFormat(wEvid, 6, "EVIDENCE", "", 1, "L", true, 0, "")

	for _, row := range sec.Controls {
		if pdf.GetY() > 272 {
			pdf.AddPage()
		}
		y := pdf.GetY()
		pdf.SetFont("Helvetica", "", 8)
		setText(pdf, rgb{120, 120, 125})
		pdf.CellFormat(wCtrl, 7, tr(truncate(row.ControlID, 16)), "", 0, "L", false, 0, "")
		setText(pdf, rgb{40, 40, 45})
		pdf.CellFormat(wName, 7, tr(truncate(row.Name, 46)), "", 0, "L", false, 0, "")
		c := statusRGB(row.Status)
		pdf.SetFont("Helvetica", "B", 7)
		setFillTint(pdf, c)
		setText(pdf, c)
		pdf.SetXY(18+wCtrl+wName, y+1)
		pdf.CellFormat(wStat-3, 5, tr(statusLabel(row.Status)), "", 0, "C", true, 0, "")
		pdf.SetXY(18+wCtrl+wName+wStat, y)
		pdf.SetFont("Helvetica", "", 7)
		setText(pdf, rgb{130, 130, 135})
		ev := row.Evidence
		if ev == "" {
			ev = "—"
		}
		pdf.CellFormat(wEvid, 7, tr(truncate(ev, 52)), "", 1, "L", false, 0, "")
		setDraw(pdf, rgb{238, 238, 240})
		pdf.SetLineWidth(0.1)
		pdf.Line(18, pdf.GetY(), 18+wCtrl+wName+wStat+wEvid, pdf.GetY())
	}
	pdf.Ln(5)
}

/* ── primitives ─────────────────────────────────────────────────────────── */

func setFill(pdf *fpdf.Fpdf, c rgb) { pdf.SetFillColor(c[0], c[1], c[2]) }
func setText(pdf *fpdf.Fpdf, c rgb) { pdf.SetTextColor(c[0], c[1], c[2]) }
func setDraw(pdf *fpdf.Fpdf, c rgb) { pdf.SetDrawColor(c[0], c[1], c[2]) }
func setFillTint(pdf *fpdf.Fpdf, c rgb) {
	pdf.SetFillColor((c[0]+255*4)/5, (c[1]+255*4)/5, (c[2]+255*4)/5)
}

func scoreStatus(score int) string {
	if score >= 80 {
		return domain.StatusCompliant
	}
	if score >= 50 {
		return domain.StatusAtRisk
	}
	return domain.StatusNonCompliant
}

func drawRing(pdf *fpdf.Fpdf, cx, cy, r, lw float64, bg, fg rgb, frac float64) {
	pdf.SetLineCapStyle("round")
	pdf.SetLineWidth(lw)
	setDraw(pdf, bg)
	pdf.Arc(cx, cy, r, r, 0, 0, 360, "D")
	if frac > 0 {
		setDraw(pdf, fg)
		pdf.Arc(cx, cy, r, r, 0, 90-frac*360, 90, "D")
	}
	pdf.SetLineCapStyle("butt")
}

func drawDonut(pdf *fpdf.Fpdf, cx, cy, r, lw float64, counts map[string]int, total int) {
	pdf.SetLineWidth(lw)
	pdf.SetLineCapStyle("butt")
	if total == 0 {
		setDraw(pdf, rgb{228, 228, 231})
		pdf.Arc(cx, cy, r, r, 0, 0, 360, "D")
		return
	}
	cum := 0.0
	for _, st := range statusOrder {
		n := counts[st]
		if n == 0 {
			continue
		}
		f := float64(n) / float64(total)
		setDraw(pdf, statusRGB(st))
		pdf.Arc(cx, cy, r, r, 0, 90-(cum+f)*360, 90-cum*360, "D")
		cum += f
	}
}

func drawPill(pdf *fpdf.Fpdf, cx, cy float64, label string, c rgb) {
	pdf.SetFont("Helvetica", "B", 10)
	w := pdf.GetStringWidth(label) + 12
	x := cx - w/2
	setFillTint(pdf, c)
	pdf.RoundedRect(x, cy-4, w, 8, 4, "1234", "F")
	setText(pdf, c)
	pdf.SetXY(x, cy-4)
	pdf.CellFormat(w, 8, label, "", 0, "C", false, 0, "")
}

func sectionHeading(pdf *fpdf.Fpdf, tr func(string) string, n, title string, accent rgb) {
	y := pdf.GetY()
	setText(pdf, rgb{205, 205, 210})
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(10, 8, n, "", 0, "L", false, 0, "")
	setText(pdf, rgb{24, 24, 27})
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, tr(strings.ToUpper(title)), "", 1, "L", false, 0, "")
	setDraw(pdf, accent)
	pdf.SetLineWidth(0.6)
	pdf.Line(18, y+9, 192, y+9)
}

func drawLogo(pdf *fpdf.Fpdf, path string, cx, top, maxW, maxH float64) bool {
	if path == "" {
		return false
	}
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false
	}
	var typ string
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		typ = "PNG"
	case ".jpg", ".jpeg":
		typ = "JPG"
	default:
		return false
	}
	info := pdf.RegisterImageOptions(path, fpdf.ImageOptions{ImageType: typ, ReadDpi: true})
	if info == nil {
		return false
	}
	w, h := info.Extent()
	if w <= 0 || h <= 0 {
		return false
	}
	scale := math.Min(maxW/w, maxH/h)
	dw, dh := w*scale, h*scale
	pdf.ImageOptions(path, cx-dw/2, top, dw, dh, false, fpdf.ImageOptions{ImageType: typ, ReadDpi: true}, 0, "")
	return true
}
