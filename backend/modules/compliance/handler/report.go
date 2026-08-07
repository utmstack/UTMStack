package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

// ReportHandler serves on-demand compliance reports (live evaluation → JSON).
// The same domain.Report is what the frontend renders and what the PDF/email
// path turns into a document.
type ReportHandler struct{ uc connectors.EvaluatorUsecase }

func NewReportHandler(uc connectors.EvaluatorUsecase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

// GetFrameworkReport godoc
//
//	@Summary     Evaluate a framework (live)
//	@Description Evaluates the framework now and returns the report JSON (3 dimensions: analysis, coverage, activity). No snapshot is stored.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       key path     string true "Framework key"
//	@Success     200 {object} domain.Report
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/report [get]
func (h *ReportHandler) GetFrameworkReport(c *gin.Context) {
	rep, err := h.uc.EvaluateFramework(c.Request.Context(), c.Param("key"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

// GenerateReport godoc
//
//	@Summary     Generate + store a report
//	@Description Evaluates the framework and persists a tenant-scoped snapshot in postgres (history), returning the report JSON.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       key path     string true "Framework key"
//	@Success     201 {object} domain.Report
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/report [post]
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	rep, err := h.uc.GenerateReport(c.Request.Context(), c.Param("key"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rep)
}

// ListReports godoc
//
//	@Summary     List stored report snapshots
//	@Description Returns report snapshot metadata (newest first), optionally filtered by framework.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       framework query    string false "Filter by framework key"
//	@Param       limit     query    int    false "Max results (default 50)"
//	@Success     200       {array}  domain.ReportSnapshotMeta
//	@Failure     500       {object} map[string]string
//	@Router      /compliance/reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.uc.ListReports(c.Request.Context(), c.Query("framework"), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []domain.ReportSnapshotMeta{}
	}
	c.JSON(http.StatusOK, items)
}

// GetReport godoc
//
//	@Summary     Get a stored report snapshot
//	@Description Returns a full stored report snapshot by id.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id  path     string true "Snapshot id"
//	@Success     200 {object} domain.ReportSnapshot
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/reports/{id} [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	snap, err := h.uc.GetReport(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *ReportHandler) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	err := h.uc.DeleteReport(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "compliance.report.delete", ResourceType: "compliance_report", ResourceID: id},
		audit_domain.COMPLIANCE_REPORT_DELETE_ATTEMPT, audit_domain.COMPLIANCE_REPORT_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writePDF(c *gin.Context, pdf []byte, name string) {
	filename := "Compliance_Report"
	if name != "" {
		filename = sanitizeFilename(name)
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`.pdf"`)
	c.Data(http.StatusOK, "application/pdf", pdf)
}

func sanitizeFilename(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ', r == '-', r == '_':
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "Compliance_Report"
	}
	return string(out)
}

// @Summary     Download a framework compliance report as PDF (live evaluation)
// @Tags        Compliance Reports
// @Security    BearerAuth
// @Produce     application/pdf
// @Param       key path string true "Framework key"
// @Success     200 {file} binary
// @Router      /compliance/frameworks/{key}/report.pdf [get]
func (h *ReportHandler) GetFrameworkReportPDF(c *gin.Context) {
	pdf, name, err := h.uc.FrameworkReportPDF(c.Request.Context(), c.Param("key"))
	if err != nil {
		writeError(c, err)
		return
	}
	writePDF(c, pdf, name)
}

// SetStatusOverride godoc
//
//	@Summary     Set a manual status override for a control
//	@Description Overrides the evaluator's computed status for a (framework, control) pair. Applied on live evaluations only; historical snapshots are unchanged.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       key    path     string true "Framework key"
//	@Param       id     path     string true "Control id"
//	@Param       body   body     object true "New status (COMPLIANT | NON_COMPLIANT | AT_RISK | NOT_COVERED | OUT_OF_SCOPE | PENDING) and optional reason"
//	@Success     204
//	@Failure     400 {object} map[string]string
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/controls/{id}/status [put]
func (h *ReportHandler) SetStatusOverride(c *gin.Context) {
	var body struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.uc.SetStatusOverride(c.Request.Context(), c.Param("key"), c.Param("id"), body.Status, body.Reason)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.status.override", ResourceType: "compliance_control", ResourceID: c.Param("id")},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ClearStatusOverride godoc
//
//	@Summary     Clear a manual status override
//	@Description Removes the manual override so the control status falls back to the evaluator's computed value.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Param       key path string true "Framework key"
//	@Param       id  path string true "Control id"
//	@Success     204
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/controls/{id}/status [delete]
func (h *ReportHandler) ClearStatusOverride(c *gin.Context) {
	err := h.uc.ClearStatusOverride(c.Request.Context(), c.Param("key"), c.Param("id"))
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.status.override.clear", ResourceType: "compliance_control", ResourceID: c.Param("id")},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetControlNote godoc
//
//	@Summary     Set / update a user note on a control
//	@Description Upserts a freeform note attached to a (framework, control). Empty body deletes the note.
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       key  path     string true "Framework key"
//	@Param       id   path     string true "Control id"
//	@Param       body body     object true "Note body ({note: string})"
//	@Success     204
//	@Failure     400 {object} map[string]string
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/controls/{id}/note [put]
func (h *ReportHandler) SetControlNote(c *gin.Context) {
	var body struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.uc.SetControlNote(c.Request.Context(), c.Param("key"), c.Param("id"), body.Note)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.note.set", ResourceType: "compliance_control", ResourceID: c.Param("id")},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ClearControlNote godoc
//
//	@Summary     Delete a user note on a control
//	@Tags        Compliance Reports
//	@Security    BearerAuth
//	@Param       key path string true "Framework key"
//	@Param       id  path string true "Control id"
//	@Success     204
//	@Router      /compliance/frameworks/{key}/controls/{id}/note [delete]
func (h *ReportHandler) ClearControlNote(c *gin.Context) {
	err := h.uc.ClearControlNote(c.Request.Context(), c.Param("key"), c.Param("id"))
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.note.clear", ResourceType: "compliance_control", ResourceID: c.Param("id")},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Download a stored report snapshot as PDF
// @Tags        Compliance Reports
// @Security    BearerAuth
// @Produce     application/pdf
// @Param       id path string true "Snapshot id"
// @Success     200 {file} binary
// @Router      /compliance/reports/{id}/pdf [get]
func (h *ReportHandler) GetSnapshotPDF(c *gin.Context) {
	pdf, name, err := h.uc.SnapshotPDF(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	writePDF(c, pdf, name)
}
