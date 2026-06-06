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

type FilterHandler struct {
	usecase connectors.FilterUsecase
}

func NewFilterHandler(uc connectors.FilterUsecase) *FilterHandler {
	return &FilterHandler{usecase: uc}
}

// @Summary     Create filter
// @Description Creates a new user filter YAML in the user overlay (pipeline: format).
// @Tags        Filters
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateFilterRequest true "relPath + pipeline YAML content"
// @Success     200 {object} dto.FilterResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [post]
func (h *FilterHandler) Create(c *gin.Context) {
	var req dto.CreateFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "eventprocessing.filter.create", ResourceType: "filter", ResourceID: req.RelPath},
		audit_domain.LOGSTASH_FILTER_CREATE_ATTEMPT, audit_domain.LOGSTASH_FILTER_CREATE_SUCCESS, err)
	if err != nil {
		writeFilterError(c, err)
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
// @Param       input body dto.UpdateFilterRequest true "relPath + new pipeline YAML content"
// @Success     200 {object} dto.FilterResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [put]
func (h *FilterHandler) Update(c *gin.Context) {
	var req dto.UpdateFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "eventprocessing.filter.update", ResourceType: "filter", ResourceID: req.RelPath},
		audit_domain.LOGSTASH_FILTER_UPDATE_ATTEMPT, audit_domain.LOGSTASH_FILTER_UPDATE_SUCCESS, err)
	if err != nil {
		writeFilterError(c, err)
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
// @Success     200 {array}  dto.FilterResponse
// @Header      200 {string} X-Total-Count "Total records"
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters [get]
func (h *FilterHandler) List(c *gin.Context) {
	var f dto.FilterFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, total, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	writePagedArray(c, items, total)
}

// @Summary     Get filter by relPath
// @Description Returns a single filter entry.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Param       relPath query string true "Relative path (e.g. syslog/syslog-generic.yaml)"
// @Success     200 {object} dto.FilterResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters/find [get]
func (h *FilterHandler) GetByRelPath(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}
	resp, err := h.usecase.GetByRelPath(c.Request.Context(), relPath)
	if err != nil {
		writeFilterError(c, err)
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
func (h *FilterHandler) Delete(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}
	err := h.usecase.Delete(c.Request.Context(), relPath)
	audit.Record(c, audit_connectors.Event{Action: "eventprocessing.filter.delete", ResourceType: "filter", ResourceID: relPath},
		audit_domain.LOGSTASH_FILTER_DELETE_ATTEMPT, audit_domain.LOGSTASH_FILTER_DELETE_SUCCESS, err)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// @Summary     Activate or deactivate a filter
// @Description Renames .yaml ↔ .yaml.disabled in the user overlay.
// @Tags        Filters
// @Security    BearerAuth
// @Produce     json
// @Param       relPath query string true  "Relative path"
// @Param       active  query bool   true  "true to activate, false to deactivate"
// @Success     200 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/filters/activate [put]
func (h *FilterHandler) ActivateDeactivate(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}
	active := c.Query("active") == "true"
	if err := h.usecase.SetActive(c.Request.Context(), relPath, active); err != nil {
		writeFilterError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
