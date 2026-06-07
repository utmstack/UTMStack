package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type JobHandler struct {
	uc connectors.JobUsecase
}

func NewJobHandler(uc connectors.JobUsecase) *JobHandler {
	return &JobHandler{uc: uc}
}

// Create godoc
//
//	@Summary     Create incident job
//	@Description Creates a new utm_incident_jobs record (a response run against an agent)
//	@Tags        Incident Jobs
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.CreateJobRequest true "Job to create"
//	@Success     200  {object} domain.UtmIncidentJob
//	@Failure     400  {object} map[string]string
//	@Failure     500  {object} map[string]string
//	@Router      /soar/incident-jobs [post]
func (h *JobHandler) Create(c *gin.Context) {
	var req dto.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.uc.Create(req, loginFromCtx(c))
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// List godoc
//
//	@Summary     List incident jobs
//	@Description Returns a paginated list of utm_incident_jobs
//	@Tags        Incident Jobs
//	@Security    BearerAuth
//	@Produce     json
//	@Param       page       query int    false "Page number (1-based)"
//	@Param       size       query int    false "Page size"
//	@Param       actionId   query int64  false "Filter by action ID"
//	@Param       agent      query string false "Filter by agent hostname"
//	@Param       status     query int    false "Filter by status"
//	@Param       originId   query int    false "Filter by origin ID"
//	@Param       originType query string false "Filter by origin type"
//	@Success     200  {array}  domain.UtmIncidentJob
//	@Header      200  {string} X-Total-Count "Total records"
//	@Failure     500  {object} map[string]string
//	@Router      /soar/incident-jobs [get]
func (h *JobHandler) List(c *gin.Context) {
	f := dto.JobFilter{
		Params:     database.Params{Page: queryInt(c, "page", 0), Size: queryInt(c, "size", 20)},
		ActionID:   queryInt64(c, "actionId"),
		Agent:      queryString(c, "agent"),
		Status:     queryIntPtr(c, "status"),
		OriginID:   queryIntPtr(c, "originId"),
		OriginType: queryString(c, "originType"),
	}
	items, total, err := h.uc.FindAll(f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	writePagedArray(c, items, total)
}

// Count godoc
//
//	@Summary     Count incident jobs
//	@Description Returns the count of utm_incident_jobs matching filters
//	@Tags        Incident Jobs
//	@Security    BearerAuth
//	@Produce     json
//	@Param       actionId   query int64  false "Filter by action ID"
//	@Param       agent      query string false "Filter by agent hostname"
//	@Param       status     query int    false "Filter by status"
//	@Param       originId   query int    false "Filter by origin ID"
//	@Param       originType query string false "Filter by origin type"
//	@Success     200  {integer} int64
//	@Failure     500  {object}  map[string]string
//	@Router      /soar/incident-jobs/count [get]
func (h *JobHandler) Count(c *gin.Context) {
	f := dto.JobFilter{
		ActionID:   queryInt64(c, "actionId"),
		Agent:      queryString(c, "agent"),
		Status:     queryIntPtr(c, "status"),
		OriginID:   queryIntPtr(c, "originId"),
		OriginType: queryString(c, "originType"),
	}
	total, err := h.uc.Count(f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, total)
}

// GetByID godoc
//
//	@Summary     Get incident job by ID
//	@Description Returns a single utm_incident_jobs record
//	@Tags        Incident Jobs
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Job ID"
//	@Success     200 {object} domain.UtmIncidentJob
//	@Failure     404 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /soar/incident-jobs/{id} [get]
func (h *JobHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	result, err := h.uc.FindByID(id)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete godoc
//
//	@Summary     Delete incident job
//	@Description Deletes a utm_incident_jobs record by ID
//	@Tags        Incident Jobs
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Job ID"
//	@Success     200 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /soar/incident-jobs/{id} [delete]
func (h *JobHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	if err := h.uc.Delete(id); err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
