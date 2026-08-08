package usecase

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

// The compliance report as a document someone hands to an auditor.
//
// Its order follows what that reader needs: what this covers and how it was
// measured, then the position, then the findings in the framework's own
// structure, then the two things they will be asked about — what a person
// signed off on, and what could not be measured at all.
//
// The strings are English. Backend localisation is separate work; this is the
// current state rather than a decision taken here.

// defaultCover ships with the product and is used when the branding has no
// cover of its own. Its top-left is transparent by design — the type goes
// there — which is why nothing is painted behind the text.
//
//go:embed assets/report-cover.png
var defaultCover []byte

type rgb [3]int

const (
	pageW      = 210.0
	margin     = 18.0
	contentW   = pageW - 2*margin
	pageBottom = 272.0 // where the footer rule sits
)

var (
	ink       = rgb{28, 30, 38}
	inkSoft   = rgb{112, 118, 130}
	hairline  = rgb{224, 227, 233}
	shade     = rgb{246, 247, 250}
	defAccent = rgb{79, 70, 229}
)

// Status colours are solid, and separated by lightness as well as hue, so the
// document survives the black-and-white print that is how a compliance report
// usually reaches a file.
func statusRGB(s domain.ComplianceStatus) rgb {
	switch s {
	case domain.StatusCompliant:
		return rgb{16, 122, 87}
	case domain.StatusNonCompliant:
		return rgb{186, 42, 42}
	case domain.StatusAtRisk:
		return rgb{173, 116, 12}
	case domain.StatusNotCovered:
		return rgb{96, 102, 115}
	case domain.StatusNotEvaluated:
		return rgb{128, 136, 152}
	case domain.StatusPending:
		return rgb{40, 108, 158}
	default: // OUT_OF_SCOPE
		return rgb{108, 88, 162}
	}
}

func statusLabel(s domain.ComplianceStatus) string {
	switch s {
	case domain.StatusCompliant:
		return "Compliant"
	case domain.StatusNonCompliant:
		return "Non-compliant"
	case domain.StatusAtRisk:
		return "At risk"
	case domain.StatusNotCovered:
		return "Not covered"
	case domain.StatusNotEvaluated:
		return "Not evaluated"
	case domain.StatusPending:
		return "Pending"
	default:
		return "Out of scope"
	}
}

var statusOrder = []domain.ComplianceStatus{
	domain.StatusCompliant, domain.StatusNonCompliant, domain.StatusAtRisk,
	domain.StatusNotCovered, domain.StatusNotEvaluated, domain.StatusPending,
	domain.StatusOutOfScope,
}

type tocEntry struct {
	label string
	page  int
	sub   bool
}

type doc struct {
	pdf    *fpdf.Fpdf
	tr     func(string) string
	accent rgb
	brand  connectors.ReportBrand
	rep    dto.ReportResponse
	toc    []tocEntry
	tocAt  int
	part   int
}

func renderReportPDF(rep dto.ReportResponse, brand connectors.ReportBrand) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(margin, 20, margin)
	// Pagination is explicit throughout, so a block is never split half-drawn.
	pdf.SetAutoPageBreak(false, 0)
	pdf.AliasNbPages("{nb}")

	d := &doc{
		pdf:    pdf,
		tr:     pdf.UnicodeTranslatorFromDescriptor(""),
		accent: parseHex(brand.AccentHex, defAccent),
		brand:  brand,
		rep:    rep,
	}
	pdf.SetFooterFunc(d.footer)

	d.cover()

	// Contents are drawn last, once every part knows its page, but they belong
	// at the front — so the page is claimed now and filled in at the end.
	pdf.AddPage()
	d.tocAt = pdf.PageNo()

	d.scope()
	d.summary()
	d.findings()
	d.verdicts()
	d.unevaluated()

	// The contents are written onto the page held near the front, then the
	// cursor is returned to the last page. fpdf takes the page count from
	// whichever page is current when the document closes, so leaving it on the
	// contents would truncate everything after it.
	last := pdf.PageNo()
	d.contents()
	pdf.SetPage(last)

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("compliance pdf: %w", err)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ── page furniture ───────────────────────────────────────────────────────────

func (d *doc) footer() {
	p := d.pdf
	if p.PageNo() == 1 {
		return // the cover carries its own marking
	}
	setDraw(p, hairline)
	p.SetLineWidth(0.2)
	p.Line(margin, pageBottom+4, pageW-margin, pageBottom+4)

	p.SetY(pageBottom + 6)
	p.SetFont("Helvetica", "", 7)
	setText(p, inkSoft)
	p.CellFormat(contentW/2, 5, d.tr(d.brandName()+"  ·  "+d.rep.FrameworkName), "", 0, "L", false, 0, "")
	p.CellFormat(contentW/2, 5, d.tr(fmt.Sprintf("Confidential  ·  %d / {nb}", p.PageNo())), "", 0, "R", false, 0, "")
}

func (d *doc) brandName() string {
	if d.brand.Name != "" {
		return d.brand.Name
	}
	return "UTMStack"
}

func (d *doc) section(title string) {
	p := d.pdf
	d.part++
	p.AddPage()
	d.toc = append(d.toc, tocEntry{label: fmt.Sprintf("%d.  %s", d.part, title), page: p.PageNo()})

	p.SetY(20)
	p.SetFont("Helvetica", "", 8)
	setText(p, d.accent)
	p.CellFormat(contentW, 4, fmt.Sprintf("PART %d", d.part), "", 1, "L", false, 0, "")
	p.SetX(margin)
	p.SetFont("Helvetica", "B", 16)
	setText(p, ink)
	p.CellFormat(contentW, 9, d.tr(title), "", 1, "L", false, 0, "")
	setDraw(p, d.accent)
	p.SetLineWidth(0.9)
	p.Line(margin, p.GetY()+1, margin+20, p.GetY()+1)
	p.SetY(p.GetY() + 8)
}

// ensure breaks the page when the next block would not fit, so a heading is
// never orphaned at the foot of a page.
func (d *doc) ensure(h float64) {
	if d.pdf.GetY()+h > pageBottom {
		d.pdf.AddPage()
		d.pdf.SetY(20)
	}
}

func (d *doc) gap(h float64) { d.pdf.SetY(d.pdf.GetY() + h) }

// ── cover ────────────────────────────────────────────────────────────────────

// cover has two shapes, because a photograph and a wall of type want opposite
// things from a page.
//
// With a branding image it is the v11 shape: the picture owns the page, and a
// short block of type sits over a scrim in the upper third — who it is for,
// what it covers, when. The score is not there; it opens the executive summary
// instead, where it has room to be explained.
//
// Without one the page is typographic and carries the score, because otherwise
// the cover would be four lines floating on white.
func (d *doc) cover() {
	p := d.pdf
	p.AddPage()

	hasImage, custom := d.drawCoverImage()
	if hasImage && custom {
		// The shipped cover keeps its upper-left clear on purpose. An uploaded
		// one carries no such promise, and nobody checks their image against a
		// text overlay before saving it — so that one gets a scrim.
		p.SetAlpha(0.88, "Normal")
		setFill(p, rgb{255, 255, 255})
		p.Rect(0, 0, pageW, 118, "F")
		p.SetAlpha(1, "Normal")
	}

	setFill(p, d.accent)
	p.Rect(0, 0, pageW, 2.5, "F")

	y := 24.0
	if drawLogo(p, d.brand.LogoPath, margin, y, 44, 13) {
		y += 20
	} else {
		p.SetXY(margin, y)
		p.SetFont("Helvetica", "B", 13)
		setText(p, d.accent)
		p.CellFormat(contentW, 7, d.tr(strings.ToUpper(d.brandName())), "", 1, "L", false, 0, "")
		y += 14
	}

	p.SetXY(margin, y+12)
	p.SetFont("Helvetica", "", 8.5)
	setText(p, inkSoft)
	p.CellFormat(contentW, 5, "COMPLIANCE REPORT", "", 1, "L", false, 0, "")

	p.SetX(margin)
	p.SetFont("Helvetica", "B", 27)
	setText(p, ink)
	p.MultiCell(contentW, 12, d.tr(d.rep.FrameworkName), "", "L", false)

	setDraw(p, d.accent)
	p.SetLineWidth(1.4)
	p.Line(margin, p.GetY()+2, margin+26, p.GetY()+2)
	p.SetY(p.GetY() + 6)

	if d.rep.FrameworkSource != "" {
		p.SetX(margin)
		p.SetFont("Helvetica", "", 8.5)
		setText(p, inkSoft)
		p.MultiCell(contentW*0.66, 4.6, d.tr(d.rep.FrameworkSource), "", "L", false)
	}

	if !hasImage {
		// The position, stated rather than dialled. A ring reading 27% in red
		// is the first thing a reader sees and tells them nothing they can act
		// on; the numbers beside it do.
		s := d.rep.Summary
		p.SetY(p.GetY() + 18)
		top := p.GetY()
		p.SetX(margin)
		p.SetFont("Helvetica", "B", 46)
		setText(p, statusRGB(scoreStatus(s.CompliantPct)))
		p.CellFormat(42, 18, fmt.Sprintf("%d%%", s.CompliantPct), "", 0, "L", false, 0, "")

		p.SetXY(margin+44, top+2)
		p.SetFont("Helvetica", "B", 10)
		setText(p, ink)
		p.CellFormat(0, 6, "Compliance score", "", 1, "L", false, 0, "")
		p.SetX(margin + 44)
		p.SetFont("Helvetica", "", 9)
		setText(p, inkSoft)
		p.CellFormat(0, 5, fmt.Sprintf("%d of %d requirements met", s.Compliant, s.Evaluated), "", 1, "L", false, 0, "")
		if s.Evaluated != s.Total {
			p.SetX(margin + 44)
			p.CellFormat(0, 5, fmt.Sprintf("%d of the framework's %d could be evaluated", s.Evaluated, s.Total), "", 1, "L", false, 0, "")
		}

		p.SetY(top + 24)
		d.bar(margin, p.GetY(), contentW, 4, s)
		p.SetY(p.GetY() + 14)
	} else {
		p.SetY(p.GetY() + 8)
	}

	if d.brand.PreparedBy != "" {
		d.coverField(p.GetY(), "PREPARED BY", d.brand.PreparedBy)
	}
	d.coverField(p.GetY()+2, "REPORTING PERIOD",
		fmt.Sprintf("%s  —  %s", fmtDate(d.rep.WindowFrom), fmtDate(d.rep.WindowTo)))
	d.coverField(p.GetY()+2, "ISSUED", fmtDateTime(d.rep.GeneratedAt))

	// With a picture the foot of the page belongs to it, so the notice joins
	// the type at the top instead of being stamped over the artwork.
	if hasImage {
		p.SetY(p.GetY() + 6)
		p.SetX(margin)
		p.SetFont("Helvetica", "", 7.5)
		setText(p, inkSoft)
		p.MultiCell(contentW*0.55, 4.2, d.tr("Confidential — prepared for internal and audit use only."), "", "L", false)
		return
	}
	setDraw(p, hairline)
	p.SetLineWidth(0.2)
	p.Line(margin, pageBottom+2, pageW-margin, pageBottom+2)
	p.SetY(pageBottom + 4)
	p.SetFont("Helvetica", "", 7.5)
	setText(p, inkSoft)
	p.CellFormat(contentW, 5, d.tr("Confidential — prepared for internal and audit use only."), "", 1, "L", false, 0, "")
}

func (d *doc) coverField(y float64, label, value string) {
	p := d.pdf
	p.SetXY(margin, y)
	_ = y
	p.SetFont("Helvetica", "", 7.5)
	setText(p, inkSoft)
	p.CellFormat(contentW, 4, label, "", 1, "L", false, 0, "")
	p.SetX(margin)
	p.SetFont("Helvetica", "B", 11)
	setText(p, ink)
	p.CellFormat(contentW, 6, d.tr(value), "", 1, "L", false, 0, "")
}

// ── contents ─────────────────────────────────────────────────────────────────

func (d *doc) contents() {
	p := d.pdf
	p.SetPage(d.tocAt)
	p.SetY(20)
	p.SetFont("Helvetica", "B", 16)
	setText(p, ink)
	p.CellFormat(contentW, 9, "Contents", "", 1, "L", false, 0, "")
	setDraw(p, d.accent)
	p.SetLineWidth(0.9)
	p.Line(margin, p.GetY()+1, margin+20, p.GetY()+1)
	p.SetY(p.GetY() + 10)

	for _, e := range d.toc {
		if p.GetY() > pageBottom-8 {
			break // a contents that spills is worse than one that is trimmed
		}
		indent, size, style := 0.0, 9.5, "B"
		if e.sub {
			indent, size, style = 6, 8.5, ""
		}
		p.SetX(margin + indent)
		p.SetFont("Helvetica", style, size)
		if e.sub {
			setText(p, inkSoft)
		} else {
			setText(p, ink)
		}
		label := d.tr(e.label)
		if p.GetStringWidth(label) > contentW-indent-22 {
			label = clip(p, label, contentW-indent-22)
		}
		w := p.GetStringWidth(label)
		p.CellFormat(w+1, 6, label, "", 0, "L", false, 0, "")

		numW := 8.0
		dotsW := contentW - indent - w - numW - 2
		if dotsW > 4 {
			setText(p, hairline)
			p.SetFont("Helvetica", "", 8)
			p.CellFormat(dotsW, 6, strings.Repeat(".", int(dotsW/1.15)), "", 0, "L", false, 0, "")
		} else if dotsW > 0 {
			p.CellFormat(dotsW, 6, "", "", 0, "L", false, 0, "")
		}
		setText(p, inkSoft)
		p.SetFont("Helvetica", "", 9)
		p.CellFormat(numW, 6, fmt.Sprintf("%d", e.page), "", 1, "R", false, 0, "")
	}
}

// ── part 1: scope and method ─────────────────────────────────────────────────

func (d *doc) scope() {
	d.section("Scope and method")

	d.para("This report states the compliance position of the monitored environment against the framework named on the cover, measured from the data held in the platform over the reporting period. Every figure is derived from that data; nothing here is entered by hand except the verdicts recorded in Part 4.")
	d.gap(4)

	d.subhead("How a control is measured")
	d.bullet("Analysis", "the control's own checks, run against the event store over the reporting period.")
	d.bullet("Coverage", "how many enabled correlation rules are tagged with the control.")
	d.bullet("Activity", "how many alerts those rules produced within the same period.")
	d.gap(4)

	d.subhead("How the score is calculated")
	d.para("The score is the share of requirements met, counted over the requirements that could be judged — not over every requirement in the framework. Both figures appear on the cover so the denominator is never in doubt. Three outcomes stay out of it, all for the same reason: nobody measured them, and counting an unmeasured requirement as a failure would report a gap that was never observed.")
	d.gap(2)
	for _, s := range []domain.ComplianceStatus{domain.StatusOutOfScope, domain.StatusPending, domain.StatusNotEvaluated} {
		d.bullet(statusLabel(s), excludedWhy(s))
	}
	d.gap(4)

	d.subhead("Reporting period")
	d.para(fmt.Sprintf("All figures cover %s to %s. Check results and activity counts are bounded by this window: a control that passed outside it is not treated as passing here.",
		fmtDate(d.rep.WindowFrom), fmtDate(d.rep.WindowTo)))
}

func excludedWhy(s domain.ComplianceStatus) string {
	switch s {
	case domain.StatusOutOfScope:
		return "a governance control — a policy or process that log data cannot prove either way."
	case domain.StatusPending:
		return "a check is declared for the control but has not yet been written."
	default:
		return "the checks exist, but the environment produced no data of the kind they read. Part 5 lists these by source."
	}
}

// ── part 2: executive summary ────────────────────────────────────────────────

func (d *doc) summary() {
	d.section("Executive summary")
	p := d.pdf
	s := d.rep.Summary

	yTop := p.GetY()
	p.SetX(margin)
	p.SetFont("Helvetica", "B", 38)
	setText(p, statusRGB(scoreStatus(s.CompliantPct)))
	p.CellFormat(36, 15, fmt.Sprintf("%d%%", s.CompliantPct), "", 0, "L", false, 0, "")
	p.SetXY(margin+38, yTop+1)
	p.SetFont("Helvetica", "", 9)
	setText(p, inkSoft)
	headline := fmt.Sprintf("%d of the framework's %d requirements are met.", s.Compliant, s.Total)
	if s.Evaluated != s.Total {
		headline = fmt.Sprintf(
			"%d of %d evaluated requirements are met. The framework has %d in total; the %d not counted could not be judged, and are broken out below.",
			s.Compliant, s.Evaluated, s.Total, s.Total-s.Evaluated)
	}
	p.MultiCell(contentW-38, 5, d.tr(headline), "", "L", false)

	p.SetY(yTop + 19)
	d.bar(margin, p.GetY(), contentW, 5.5, s)
	p.SetY(p.GetY() + 12)

	// The breakdown as a table. A legend beside a ring makes a reader match
	// colours in order to read a number that was already a number.
	d.countTable(s)

	if s.NotEvaluated > 0 {
		d.gap(7)
		d.callout("Coverage gap",
			fmt.Sprintf("%d requirements could not be evaluated because the environment produced no data of the kind their checks read. They are excluded from the score rather than counted as failures. Part 5 lists them by the source that would settle them — connecting those sources is what moves this number.", s.NotEvaluated))
	}
}

func (d *doc) countTable(s dto.ReportSummary) {
	p := d.pdf
	colStatus, colN, colPct := contentW-60, 28.0, 32.0

	p.SetX(margin)
	setFill(p, shade)
	p.SetFont("Helvetica", "B", 7)
	setText(p, inkSoft)
	p.CellFormat(colStatus, 7, "  OUTCOME", "", 0, "L", true, 0, "")
	p.CellFormat(colN, 7, "REQUIREMENTS  ", "", 0, "R", true, 0, "")
	p.CellFormat(colPct, 7, "SHARE  ", "", 1, "R", true, 0, "")

	for _, st := range statusOrder {
		n := countOf(s, st)
		pct := 0
		if s.Total > 0 {
			pct = n * 100 / s.Total
		}
		y := p.GetY()
		p.SetX(margin)
		setFill(p, statusRGB(st))
		p.Rect(margin+3, y+2.7, 2.2, 2.2, "F")
		p.SetFont("Helvetica", "", 9)
		setText(p, ink)
		p.CellFormat(colStatus, 7.5, "      "+statusLabel(st), "", 0, "L", false, 0, "")
		p.SetFont("Helvetica", "B", 9)
		p.CellFormat(colN, 7.5, fmt.Sprintf("%d  ", n), "", 0, "R", false, 0, "")
		p.SetFont("Helvetica", "", 9)
		setText(p, inkSoft)
		p.CellFormat(colPct, 7.5, fmt.Sprintf("%d%%  ", pct), "", 1, "R", false, 0, "")
		setDraw(p, hairline)
		p.SetLineWidth(0.1)
		p.Line(margin, p.GetY(), pageW-margin, p.GetY())
	}
}

// ── part 3: findings ─────────────────────────────────────────────────────────

func (d *doc) findings() {
	d.section("Findings by section")
	p := d.pdf

	byID := map[string]dto.ControlRow{}
	for _, c := range d.rep.Controls {
		byID[c.ControlID] = c
	}

	for _, sec := range d.rep.Sections {
		d.ensure(36)
		d.toc = append(d.toc, tocEntry{label: sec.Name, page: p.PageNo(), sub: true})

		yTop := p.GetY()
		p.SetX(margin)
		p.SetFont("Helvetica", "B", 11)
		setText(p, ink)
		p.MultiCell(contentW-50, 5.6, d.tr(sec.Name), "", "L", false)
		nameBottom := p.GetY()

		d.bar(pageW-margin-46, yTop+1.6, 32, 3.2, sec.Summary)
		p.SetXY(pageW-margin-12, yTop)
		p.SetFont("Helvetica", "B", 10)
		setText(p, statusRGB(scoreStatus(sec.Summary.CompliantPct)))
		p.CellFormat(12, 5.6, fmt.Sprintf("%d%%", sec.Summary.CompliantPct), "", 1, "R", false, 0, "")

		p.SetY(nameBottom + 1)
		setDraw(p, hairline)
		p.SetLineWidth(0.3)
		p.Line(margin, p.GetY(), pageW-margin, p.GetY())
		p.SetY(p.GetY() + 3)

		for _, req := range sec.Requirements {
			d.requirement(req, byID)
		}
		d.gap(4)
	}
}

// requirement is the level the framework asks at, and the level the score is
// over. The previous document went from section straight to controls, which
// lost the sentence an audit is about: which requirement this answers.
func (d *doc) requirement(req dto.ReportRequirement, byID map[string]dto.ControlRow) {
	p := d.pdf
	d.ensure(24)

	nameW := contentW - 32
	yTop := p.GetY()
	p.SetFont("Helvetica", "B", 8.5)
	lines := p.SplitLines([]byte(d.tr(req.Name)), nameW-5)
	h := float64(len(lines))*4.4 + 3.2

	setFill(p, shade)
	p.Rect(margin, yTop, contentW, h, "F")
	p.SetXY(margin+2.5, yTop+1.5)
	setText(p, ink)
	p.MultiCell(nameW-5, 4.4, d.tr(req.Name), "", "L", false)
	d.chip(pageW-margin-2.5, yTop+1.2, req.Status)

	p.SetY(yTop + h + 1.6)

	// One control of the same name is the same thing said twice; its own id
	// follows on the next line, so the requirement's would only be noise.
	collapsed := len(req.ControlIDs) == 1
	if c, ok := byID[firstOf(req.ControlIDs)]; collapsed && ok {
		collapsed = strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(req.Name))
	}
	if req.ID != "" && !collapsed {
		p.SetX(margin + 2.5)
		p.SetFont("Helvetica", "", 6.8)
		setText(p, inkSoft)
		p.CellFormat(contentW, 3.4, d.tr(req.ID), "", 1, "L", false, 0, "")
	}

	for _, cid := range req.ControlIDs {
		c, ok := byID[cid]
		if !ok {
			continue
		}
		// Most requirements here name a single control that carries the same
		// title. Printing both makes the page look padded and gives a reader
		// two identical lines to reconcile.
		d.control(c, strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(req.Name)))
	}
	d.gap(3)
}

func (d *doc) control(c dto.ControlRow, hideName bool) {
	p := d.pdf
	d.ensure(16)

	const idW = 30.0
	nameW := contentW - 6 - idW - 28
	yTop := p.GetY()

	p.SetXY(margin+6, yTop)
	p.SetFont("Helvetica", "", 7.2)
	setText(p, inkSoft)
	p.CellFormat(idW, 4.4, d.tr(c.ControlID), "", 0, "L", false, 0, "")

	bottom := yTop
	if !hideName {
		p.SetXY(margin+6+idW, yTop)
		p.SetFont("Helvetica", "", 8.5)
		setText(p, ink)
		// Wrapped, not cut. A name ending in "…" mid-word is what makes a
		// document read as a screen dump.
		p.MultiCell(nameW, 4.4, d.tr(c.Name), "", "L", false)
		bottom = p.GetY()
		// The requirement above already carries the verdict when the two are
		// the same thing, so the chip would only be a second copy of it.
		d.chip(pageW-margin-2.5, yTop-0.3, c.Status)
	}

	if ev := d.evidence(c); ev != "" {
		p.SetXY(margin+6+idW, bottom)
		p.SetFont("Helvetica", "", 7.2)
		setText(p, inkSoft)
		p.MultiCell(nameW, 3.9, d.tr(ev), "", "L", false)
		bottom = p.GetY()
	}

	if c.Note != "" {
		noteTop := bottom + 0.6
		p.SetXY(margin+8.5+idW, noteTop)
		p.SetFont("Helvetica", "I", 7.2)
		setText(p, ink)
		p.MultiCell(nameW-2.5, 3.9, d.tr("“"+c.Note+"”"+by(c)), "", "L", false)
		setDraw(p, d.accent)
		p.SetLineWidth(0.5)
		p.Line(margin+7+idW, noteTop, margin+7+idW, p.GetY()-0.4)
		bottom = p.GetY()
	}

	p.SetY(bottom + 1.4)
	setDraw(p, hairline)
	p.SetLineWidth(0.1)
	p.Line(margin+6, p.GetY(), pageW-margin, p.GetY())
	p.SetY(p.GetY() + 1.4)
}

// evidence reduces a control to the numbers behind it: what the checks
// returned, and — when nothing ran — which data never arrived. The stored
// evidence line often repeats the control's own name, which earns no space.
func (d *doc) evidence(c dto.ControlRow) string {
	if len(c.Checks) == 0 {
		return c.Evidence
	}
	var na, failed []string
	passed, errored := 0, 0
	for _, ch := range c.Checks {
		switch ch.Outcome {
		case domain.CheckPassed:
			passed++
		case domain.CheckNotApplicable:
			if ch.DataType != "" {
				na = append(na, ch.DataType)
			}
		case domain.CheckError:
			errored++
		case domain.CheckFailed:
			need := int64(1)
			if ch.Required != nil {
				need = int64(*ch.Required)
			}
			failed = append(failed, fmt.Sprintf("%s — %d of %d required", ch.Name, ch.Hits, need))
		}
	}
	switch {
	case passed == 0 && len(failed) == 0 && len(na) > 0:
		return "No " + strings.Join(dedupe(na), " or ") + " data reached the platform during the period."
	case len(failed) > 0:
		return clipText(strings.Join(failed, "  ·  "), 190)
	case errored > 0:
		return fmt.Sprintf("%d of %d checks could not run.", errored, len(c.Checks))
	case passed > 0:
		return fmt.Sprintf("%d of %d checks passed.", passed, len(c.Checks))
	}
	return c.Evidence
}

func firstOf(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func by(c dto.ControlRow) string {
	if c.EditedBy == "" {
		return ""
	}
	return "  — " + c.EditedBy
}

// ── part 4: recorded verdicts ────────────────────────────────────────────────

// verdicts is what an audit is actually about: where a person recorded a view
// of their own, what they said, and what the engine had said at the time.
func (d *doc) verdicts() {
	var edited []dto.ControlRow
	for _, c := range d.rep.Controls {
		if c.EditedAt != nil {
			edited = append(edited, c)
		}
	}
	if len(edited) == 0 {
		return
	}
	sort.Slice(edited, func(i, j int) bool { return edited[i].ControlID < edited[j].ControlID })

	d.section("Recorded verdicts and notes")
	d.para("Each entry below was written by a named person against a specific control. Where a verdict differs from the engine's, both are shown: the measurement is never replaced, only overruled on the record.")
	d.gap(5)

	p := d.pdf
	for _, c := range edited {
		d.ensure(28)
		yTop := p.GetY()

		p.SetX(margin)
		p.SetFont("Helvetica", "B", 9)
		setText(p, ink)
		p.MultiCell(contentW-30, 4.8, d.tr(c.ControlID+"  ·  "+c.Name), "", "L", false)
		d.chip(pageW-margin-2.5, yTop-0.3, c.Status)

		p.SetX(margin)
		p.SetFont("Helvetica", "", 7.2)
		setText(p, inkSoft)
		meta := []string{"Engine: " + statusLabel(c.EngineStatus)}
		switch {
		case c.OriginalStatus == "":
			meta = append(meta, "note only — the verdict was left to the engine")
		case c.OriginalStatus != c.EngineStatus:
			meta = append(meta, "recorded when the engine said "+statusLabel(c.OriginalStatus))
		}
		if c.EditedBy != "" {
			meta = append(meta, c.EditedBy)
		}
		if c.EditedAt != nil {
			meta = append(meta, fmtDate(*c.EditedAt))
		}
		p.MultiCell(contentW-30, 4.2, d.tr(strings.Join(meta, "  ·  ")), "", "L", false)

		if c.Note != "" {
			noteTop := p.GetY() + 0.8
			p.SetXY(margin+3.5, noteTop)
			p.SetFont("Helvetica", "", 8.5)
			setText(p, ink)
			p.MultiCell(contentW-7, 4.6, d.tr(c.Note), "", "L", false)
			setDraw(p, d.accent)
			p.SetLineWidth(0.7)
			p.Line(margin+1, noteTop, margin+1, p.GetY()-0.5)
		}

		p.SetY(p.GetY() + 2.5)
		setDraw(p, hairline)
		p.SetLineWidth(0.1)
		p.Line(margin, p.GetY(), pageW-margin, p.GetY())
		p.SetY(p.GetY() + 4)
	}
}

// ── part 5: what could not be measured ───────────────────────────────────────

// unevaluated turns the excluded requirements into the one page that is
// actionable: every missing data type is a source somebody can connect.
func (d *doc) unevaluated() {
	byType := map[string][]dto.ControlRow{}
	for _, c := range d.rep.Controls {
		if c.Status != domain.StatusNotEvaluated {
			continue
		}
		for _, ch := range c.Checks {
			if ch.Outcome == domain.CheckNotApplicable && ch.DataType != "" {
				byType[ch.DataType] = append(byType[ch.DataType], c)
				break
			}
		}
	}
	if len(byType) == 0 {
		return
	}
	types := make([]string, 0, len(byType))
	for k := range byType {
		types = append(types, k)
	}
	sort.Strings(types)

	d.section("Not evaluated — missing sources")
	d.para("These controls have checks, but the platform received no data of the kind those checks read during the period. They are excluded from the score rather than counted as failures. Each heading below is a source that would settle the controls beneath it.")
	d.gap(5)

	p := d.pdf
	for _, dt := range types {
		rows := byType[dt]
		d.ensure(22)
		p.SetX(margin)
		p.SetFont("Helvetica", "B", 10)
		setText(p, ink)
		p.CellFormat(contentW-36, 6.5, d.tr(dt), "", 0, "L", false, 0, "")
		p.SetFont("Helvetica", "", 8.5)
		setText(p, inkSoft)
		p.CellFormat(36, 6.5, fmt.Sprintf("%d controls", len(rows)), "", 1, "R", false, 0, "")
		setDraw(p, hairline)
		p.SetLineWidth(0.2)
		p.Line(margin, p.GetY(), pageW-margin, p.GetY())
		p.SetY(p.GetY() + 2)

		for _, c := range rows {
			d.ensure(6)
			p.SetX(margin + 4)
			p.SetFont("Helvetica", "", 8)
			setText(p, inkSoft)
			p.MultiCell(contentW-8, 4.2, d.tr(c.ControlID+"  ·  "+c.Name), "", "L", false)
		}
		d.gap(4)
	}
}

// ── shared drawing ───────────────────────────────────────────────────────────

// bar is the whole population stacked. It says more in the same space than a
// percentage does, because it shows what the remainder is made of.
func (d *doc) bar(x, y, w, h float64, s dto.ReportSummary) {
	p := d.pdf
	if s.Total == 0 {
		setFill(p, hairline)
		p.Rect(x, y, w, h, "F")
		return
	}
	cx := x
	for _, st := range statusOrder {
		n := countOf(s, st)
		if n == 0 {
			continue
		}
		seg := w * float64(n) / float64(s.Total)
		setFill(p, statusRGB(st))
		p.Rect(cx, y, seg, h, "F")
		cx += seg
	}
}

// chip is outlined rather than filled: it holds up in black and white, and it
// does not shout from the middle of a table. x is its right edge.
func (d *doc) chip(x, y float64, s domain.ComplianceStatus) {
	p := d.pdf
	label := statusLabel(s)
	c := statusRGB(s)
	p.SetFont("Helvetica", "B", 6.3)
	w := p.GetStringWidth(label) + 4

	setDraw(p, c)
	p.SetLineWidth(0.3)
	p.RoundedRect(x-w, y+0.5, w, 4.3, 0.7, "1234", "D")
	setText(p, c)
	p.SetXY(x-w, y+0.5)
	p.CellFormat(w, 4.3, label, "", 0, "C", false, 0, "")
}

func (d *doc) callout(title, body string) {
	p := d.pdf
	d.ensure(26)
	yTop := p.GetY()
	p.SetX(margin + 4)
	p.SetFont("Helvetica", "B", 9)
	setText(p, ink)
	p.CellFormat(contentW-6, 5.5, title, "", 1, "L", false, 0, "")
	p.SetX(margin + 4)
	p.SetFont("Helvetica", "", 8.5)
	setText(p, inkSoft)
	p.MultiCell(contentW-6, 4.6, d.tr(body), "", "L", false)
	setDraw(p, statusRGB(domain.StatusNotEvaluated))
	p.SetLineWidth(0.9)
	p.Line(margin, yTop, margin, p.GetY())
	p.SetY(p.GetY() + 2)
}

func (d *doc) subhead(s string) {
	p := d.pdf
	d.ensure(12)
	p.SetX(margin)
	p.SetFont("Helvetica", "B", 9.5)
	setText(p, ink)
	p.CellFormat(contentW, 6, s, "", 1, "L", false, 0, "")
}

func (d *doc) para(s string) {
	p := d.pdf
	d.ensure(13)
	p.SetX(margin)
	p.SetFont("Helvetica", "", 8.5)
	setText(p, inkSoft)
	p.MultiCell(contentW, 4.8, d.tr(s), "", "L", false)
}

func (d *doc) bullet(name, body string) {
	p := d.pdf
	d.ensure(12)
	y := p.GetY()
	setFill(p, d.accent)
	p.Rect(margin+0.5, y+1.9, 1.8, 1.8, "F")
	p.SetXY(margin+5, y)
	p.SetFont("Helvetica", "B", 8.5)
	setText(p, ink)
	nw := p.GetStringWidth(name) + 2
	p.CellFormat(nw, 4.9, name, "", 0, "L", false, 0, "")
	p.SetFont("Helvetica", "", 8.5)
	setText(p, inkSoft)
	p.MultiCell(contentW-5-nw, 4.9, d.tr(body), "", "L", false)
	d.gap(1.2)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func countOf(s dto.ReportSummary, st domain.ComplianceStatus) int {
	switch st {
	case domain.StatusCompliant:
		return s.Compliant
	case domain.StatusNonCompliant:
		return s.NonCompliant
	case domain.StatusAtRisk:
		return s.AtRisk
	case domain.StatusNotCovered:
		return s.NotCovered
	case domain.StatusNotEvaluated:
		return s.NotEvaluated
	case domain.StatusPending:
		return s.Pending
	default:
		return s.OutOfScope
	}
}

func scoreStatus(score int) domain.ComplianceStatus {
	switch {
	case score >= 80:
		return domain.StatusCompliant
	case score >= 50:
		return domain.StatusAtRisk
	default:
		return domain.StatusNonCompliant
	}
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func clipText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func clip(p *fpdf.Fpdf, s string, w float64) string {
	for len(s) > 4 && p.GetStringWidth(s) > w {
		s = s[:len(s)-2]
	}
	return s + "..."
}

func fmtDate(t time.Time) string     { return t.UTC().Format("2 January 2006") }
func fmtDateTime(t time.Time) string { return t.UTC().Format("2 January 2006, 15:04 UTC") }

func setFill(p *fpdf.Fpdf, c rgb) { p.SetFillColor(c[0], c[1], c[2]) }
func setText(p *fpdf.Fpdf, c rgb) { p.SetTextColor(c[0], c[1], c[2]) }
func setDraw(p *fpdf.Fpdf, c rgb) { p.SetDrawColor(c[0], c[1], c[2]) }

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

// drawCoverImage lays the cover picture over the whole page: the tenant's if
// branding supplies one, the shipped one otherwise. custom reports which, since
// only the shipped image is known to leave room for the type.
func (d *doc) drawCoverImage() (drawn, custom bool) {
	if d.brand.CoverPath != "" {
		if drawCover(d.pdf, d.brand.CoverPath) {
			return true, true
		}
	}
	return drawEmbeddedCover(d.pdf), false
}

// drawEmbeddedCover draws the shipped image, which travels in the binary rather
// than on disk so a report never depends on a file the deployment may not have.
func drawEmbeddedCover(pdf *fpdf.Fpdf) bool {
	const name = "compliance-default-cover"
	info := pdf.GetImageInfo(name)
	if info == nil {
		info = pdf.RegisterImageOptionsReader(name,
			fpdf.ImageOptions{ImageType: "png", ReadDpi: true}, bytes.NewReader(defaultCover))
	}
	if info == nil || info.Width() == 0 || info.Height() == 0 {
		return false
	}
	// The shipped cover is a composition, not a texture: its wordmark and note
	// sit in the bottom-right, so it is fitted to the page width and anchored
	// to the foot. Cropping to fill would cut exactly the part that was drawn
	// to be seen.
	w := pageW
	h := w * info.Height() / info.Width()
	pdf.ImageOptions(name, 0, 297-h, w, h, false,
		fpdf.ImageOptions{ImageType: "png", ReadDpi: true}, 0, "")
	return true
}

// drawCover fills the page with the branding image, cropping rather than
// squashing: a stretched photograph is the first thing that reads as amateur.
func drawCover(pdf *fpdf.Fpdf, path string) bool {
	ext, info := imageInfo(pdf, path)
	if info == nil {
		return false
	}
	placeCover(pdf, path, ext, info)
	return true
}

// placeCover crops an uploaded image to fill the page, centred. A photograph
// with no known composition is better cropped than letterboxed.
func placeCover(pdf *fpdf.Fpdf, name, ext string, info *fpdf.ImageInfoType) {
	const pageH = 297.0
	w, h := pageW, pageW*info.Height()/info.Width()
	if h < pageH {
		h, w = pageH, pageH*info.Width()/info.Height()
	}
	pdf.ImageOptions(name, (pageW-w)/2, (pageH-h)/2, w, h, false,
		fpdf.ImageOptions{ImageType: ext, ReadDpi: true}, 0, "")
}

func imageInfo(pdf *fpdf.Fpdf, path string) (string, *fpdf.ImageInfoType) {
	if path == "" {
		return "", nil
	}
	if _, err := os.Stat(path); err != nil {
		return "", nil
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "jpeg" {
		ext = "jpg"
	}
	if ext != "png" && ext != "jpg" {
		return "", nil
	}
	info := pdf.RegisterImageOptions(path, fpdf.ImageOptions{ImageType: ext, ReadDpi: true})
	if info == nil || info.Width() == 0 || info.Height() == 0 {
		return "", nil
	}
	return ext, info
}

func drawLogo(pdf *fpdf.Fpdf, path string, x, top, maxW, maxH float64) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "jpeg" {
		ext = "jpg"
	}
	if ext != "png" && ext != "jpg" {
		return false
	}
	info := pdf.RegisterImageOptions(path, fpdf.ImageOptions{ImageType: ext, ReadDpi: true})
	if info == nil || info.Width() == 0 || info.Height() == 0 {
		return false
	}
	w, h := maxW, maxW*info.Height()/info.Width()
	if h > maxH {
		h, w = maxH, maxH*info.Width()/info.Height()
	}
	pdf.ImageOptions(path, x, top, w, h, false, fpdf.ImageOptions{ImageType: ext, ReadDpi: true}, 0, "")
	return true
}
