package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// ReportHandler serves the standing report: one per framework per tenant, the
// same document the UI renders, the PDF is drawn from, and the schedule mails.
type ReportHandler struct{ uc connectors.EvaluatorUsecase }

func NewReportHandler(uc connectors.EvaluatorUsecase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

// GetReport godoc
//
//	@Summary     Get the standing report for a framework
//	@Description Returns the last evaluation, including any human edits made to it. Does not re-run the framework.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       key path     string true "Framework key"
//	@Success     200 {object} dto.ReportResponse
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/report [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	rep, err := h.uc.Get(c.Request.Context(), c.Param("key"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

// Evaluate godoc
//
//	@Summary     Run a framework evaluation
//	@Description Re-evaluates the framework and replaces the standing report, keeping the edits already on it. windowDays defaults to the framework's schedule, then to 30.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       key        path  string true  "Framework key"
//	@Param       windowDays query int    false "Days the report covers"
//	@Success     200 {object} dto.ReportResponse
//	@Failure     403 {object} map[string]string
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/report [post]
func (h *ReportHandler) Evaluate(c *gin.Context) {
	windowDays, _ := strconv.Atoi(c.Query("windowDays"))
	rep, err := h.uc.Evaluate(c.Request.Context(), c.Param("key"), windowDays)
	audit.Record(c, audit_connectors.Event{Action: "compliance.report.evaluate", ResourceType: "compliance_report", ResourceID: c.Param("key")},
		audit_domain.COMPLIANCE_REPORT_EVALUATE_ATTEMPT, audit_domain.COMPLIANCE_REPORT_EVALUATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

// ListReports godoc
//
//	@Summary     List the tenant's standing reports
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {array} dto.ReportMeta
//	@Router      /compliance/reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	items, err := h.uc.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []dto.ReportMeta{}
	}
	c.JSON(http.StatusOK, items)
}

// DeleteReport godoc
//
//	@Summary     Delete a framework's report
//	@Description Discards the report and every annotation on it.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Param       key path string true "Framework key"
//	@Success     204
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/report [delete]
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	key := c.Param("key")
	err := h.uc.Delete(c.Request.Context(), key)
	audit.Record(c, audit_connectors.Event{Action: "compliance.report.delete", ResourceType: "compliance_report", ResourceID: key},
		audit_domain.COMPLIANCE_REPORT_DELETE_ATTEMPT, audit_domain.COMPLIANCE_REPORT_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetReportPDF godoc
//
//	@Summary     Download the standing report as PDF
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     application/pdf
//	@Param       key path string true "Framework key"
//	@Success     200 {file} binary
//	@Router      /compliance/frameworks/{key}/report.pdf [get]
func (h *ReportHandler) GetReportPDF(c *gin.Context) {
	pdf, name, err := h.uc.PDF(c.Request.Context(), c.Param("key"), actorEmail(c))
	if err != nil {
		writeError(c, err)
		return
	}
	writePDF(c, pdf, name)
}

// EditControl godoc
//
//	@Summary     Record a human verdict on a control
//	@Description Overrides the engine on one control row. Requirements, sections and the score recompute from it — they are never edited directly. A note is required: an override no one can explain is worth nothing to whoever reads the report later.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       key  path string                 true "Framework key"
//	@Param       id   path string                 true "Control id"
//	@Param       body body dto.EditControlRequest true "Status and justification"
//	@Success     200 {object} dto.ReportResponse
//	@Failure     400 {object} map[string]string
//	@Failure     404 {object} map[string]string
//	@Failure     409 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/controls/{id}/status [put]
func (h *ReportHandler) EditControl(c *gin.Context) {
	var req dto.EditControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rep, err := h.uc.EditControl(c.Request.Context(), actorEmail(c), c.Param("key"), c.Param("id"), req)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.status.edit", ResourceType: "compliance_control", ResourceID: c.Param("id")},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

// History godoc
//
//	@Summary     Score over time for a framework
//	@Description One point per day. Total and evaluated travel with the score because the percentage alone cannot tell a fix from a log source going quiet.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       key  path  string true  "Framework key"
//	@Param       from query string false "RFC3339; defaults to 90 days ago"
//	@Param       to   query string false "RFC3339; defaults to now"
//	@Success     200 {array} dto.ScorePoint
//	@Router      /compliance/frameworks/{key}/history [get]
func (h *ReportHandler) History(c *gin.Context) {
	to := parseTimeOr(c.Query("to"), time.Now().UTC())
	from := parseTimeOr(c.Query("from"), to.AddDate(0, 0, -90))
	items, err := h.uc.History(c.Request.Context(), c.Param("key"), from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []dto.ScorePoint{}
	}
	c.JSON(http.StatusOK, items)
}

// HistoryPDF godoc
//
//	@Summary     Download the report behind a point on the chart
//	@Description Renders that day's stored document. A point whose body has aged past retention returns 404.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     application/pdf
//	@Param       key path  string true "Framework key"
//	@Param       day query string true "YYYY-MM-DD"
//	@Success     200 {file} binary
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/history.pdf [get]
func (h *ReportHandler) HistoryPDF(c *gin.Context) {
	day, err := time.Parse("2006-01-02", c.Query("day"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "day must be YYYY-MM-DD"})
		return
	}
	pdf, name, err := h.uc.HistoryPDF(c.Request.Context(), c.Param("key"), actorEmail(c), day)
	if err != nil {
		writeError(c, err)
		return
	}
	writePDF(c, pdf, name)
}

// actorEmail is who is acting. It reaches a report as "prepared by" and an
// edit as its author; a uuid would say nothing to whoever reads either later.
func actorEmail(c *gin.Context) string {
	if a := middleware.ActorFromGin(c); a != nil {
		return a.Email
	}
	return ""
}

func parseTimeOr(s string, def time.Time) time.Time {
	if s == "" {
		return def
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC()
	}
	return def
}

func writePDF(c *gin.Context, pdf []byte, name string) {
	filename := "Compliance_Report"
	if name != "" {
		filename = sanitizeFilename(name)
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`.pdf"`)
	c.Data(http.StatusOK, "application/pdf", pdf)
}

// sanitizeFilename keeps a framework name usable as a download filename.
func sanitizeFilename(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ', r == '-', r == '_', r == '.':
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "Compliance_Report"
	}
	return string(out)
}
