package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
//	@Description Evaluates the framework and persists a snapshot in OpenSearch (history), returning the report JSON.
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
