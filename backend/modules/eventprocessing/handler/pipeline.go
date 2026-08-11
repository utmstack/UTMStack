package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type PipelineHandler struct {
	usecase connectors.PipelineUsecase
}

func NewPipelineHandler(uc connectors.PipelineUsecase) *PipelineHandler {
	return &PipelineHandler{usecase: uc}
}

// @Summary     Create filter
// @Description Creates a new user filter YAML in the user overlay (pipeline: format).
// @Tags        Filters
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreatePipelineRequest true "relPath + pipeline YAML content"
// @Success     200 {object} dto.PipelineResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [post]
func (h *PipelineHandler) Create(c *gin.Context) {
	var req dto.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "eventprocessing.filter.create", ResourceType: "filter", ResourceID: req.RelPath},
		audit_domain.LOGSTASH_FILTER_CREATE_ATTEMPT, audit_domain.LOGSTASH_FILTER_CREATE_SUCCESS, err)
	if err != nil {
		writePipelineError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update filter
// @Description Replaces the content of a user filter. System filters are read-only.
// @Tags        Filters
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdatePipelineRequest true "relPath + new pipeline YAML content"
// @Success     200 {object} dto.PipelineResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [put]
func (h *PipelineHandler) Update(c *gin.Context) {
	var req dto.UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "eventprocessing.filter.update", ResourceType: "filter", ResourceID: req.RelPath},
		audit_domain.LOGSTASH_FILTER_UPDATE_ATTEMPT, audit_domain.LOGSTASH_FILTER_UPDATE_SUCCESS, err)
	if err != nil {
		writePipelineError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List filters
// @Description Returns all filters (system + user) with optional filtering.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Param       relPath.contains query string false "Filter by relPath containing value"
// @Param       isActive.equals  query bool   false "Filter by active state"
// @Param       system.equals    query bool   false "true = system only, false = user only"
// @Param       page             query int    false "Page (1-based)"
// @Param       size             query int    false "Page size (default 50)"
// @Success     200 {array}  dto.PipelineResponse
// @Header      200 {string} X-Total-Count "Total records"
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [get]
func (h *PipelineHandler) List(c *gin.Context) {
	var f dto.PipelineFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writePipelineError(c, err)
		return
	}
	writePagedArray(c, res.Items, res.Total)
}

// @Summary     List filter data types
// @Description Returns the sorted, distinct dataTypes declared across all filters.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} string
// @Router      /eventprocessing/filters/data-types [get]
func (h *PipelineHandler) DataTypes(c *gin.Context) {
	dts := h.usecase.DataTypes(c.Request.Context())
	if dts == nil {
		dts = []string{}
	}
	c.JSON(http.StatusOK, dts)
}

// @Summary     Get filter by relPath
// @Description Returns a single filter entry.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Param       relPath query string true "Relative path (e.g. syslog/syslog-generic.yaml)"
// @Success     200 {object} dto.PipelineResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters/find [get]
func (h *PipelineHandler) GetByRelPath(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}
	resp, err := h.usecase.GetByRelPath(c.Request.Context(), relPath)
	if err != nil {
		writePipelineError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete filter
// @Description Deletes a user filter. System filters cannot be deleted.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Param       relPath query string true "Relative path"
// @Success     200 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [delete]
func (h *PipelineHandler) Delete(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}
	err := h.usecase.Delete(c.Request.Context(), relPath)
	audit.Record(c, audit_connectors.Event{Action: "eventprocessing.filter.delete", ResourceType: "filter", ResourceID: relPath},
		audit_domain.LOGSTASH_FILTER_DELETE_ATTEMPT, audit_domain.LOGSTASH_FILTER_DELETE_SUCCESS, err)
	if err != nil {
		writePipelineError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// @Summary     Activate or deactivate a filter
// @Description Renames .yaml <-> .yaml.disabled in whichever overlay (system or user) owns the filter.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Param       relPath query string true  "Relative path"
// @Param       active  query bool   true  "true to activate, false to deactivate"
// @Success     200 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters/activate [put]
func (h *PipelineHandler) ActivateDeactivate(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}
	active := c.Query("active") == "true"
	if err := h.usecase.SetActive(c.Request.Context(), relPath, active); err != nil {
		writePipelineError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// @Summary     Reorder a filter
// @Description Sets any filter's (system or custom) position in the global pipeline order.
// @Tags        Filters
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdatePipelineOrderRequest true "relPath + new order"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /eventprocessing/pipelines/order [put]
func (h *PipelineHandler) SetOrder(c *gin.Context) {
	var req dto.UpdatePipelineOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.SetOrder(c.Request.Context(), req.Order); err != nil {
		writePipelineError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
