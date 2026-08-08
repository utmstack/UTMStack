package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type ScheduleHandler struct{ uc connectors.ScheduleUsecase }

func NewScheduleHandler(uc connectors.ScheduleUsecase) *ScheduleHandler {
	return &ScheduleHandler{uc: uc}
}

// Create godoc
//
//	@Summary     Create a report schedule
//	@Description Schedules a recurring compliance report (framework + cron + recipients) for the current user.
//	@Tags        Compliance Report Schedules
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body     dto.CreateScheduleRequest true "Schedule to create"
//	@Success     201  {object} dto.ScheduleResponse
//	@Failure     400  {object} map[string]string
//	@Router      /compliance-report-schedules [post]
func (h *ScheduleHandler) Create(c *gin.Context) {
	actor, _ := userIDFromCtx(c)
	var req dto.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), actor, req)
	ev := audit_connectors.Event{Action: "compliance.schedule.create", ResourceType: "compliance_report_schedule"}
	if resp != nil {
		ev.ResourceID = resp.ID.String()
	}
	audit.Record(c, ev, audit_domain.COMPLIANCE_SCHEDULE_CREATE_ATTEMPT, audit_domain.COMPLIANCE_SCHEDULE_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// Update godoc
//
//	@Summary     Update a report schedule
//	@Description Updates an existing schedule (framework, cron, recipients).
//	@Tags        Compliance Report Schedules
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body     dto.UpdateScheduleRequest true "Schedule body"
//	@Success     200  {object} dto.ScheduleResponse
//	@Failure     400  {object} map[string]string
//	@Failure     404  {object} map[string]string
//	@Router      /compliance-report-schedules [put]
func (h *ScheduleHandler) Update(c *gin.Context) {
	actor, _ := userIDFromCtx(c)
	var req dto.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), actor, req)
	audit.Record(c, audit_connectors.Event{Action: "compliance.schedule.update", ResourceType: "compliance_report_schedule", ResourceID: req.ID.String()},
		audit_domain.COMPLIANCE_SCHEDULE_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_SCHEDULE_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListByUser godoc
//
//	@Summary     List my report schedules
//	@Description Lists the current user's report schedules, optionally filtered by framework.
//	@Tags        Compliance Report Schedules
//	@Security    BearerAuth
//	@Produce     json
//	@Param       framework query    string false "Filter by framework key"
//	@Success     200       {array}  dto.ScheduleResponse
//	@Failure     400       {object} map[string]string
//	@Router      /compliance-report-schedules/by-user [get]
func (h *ScheduleHandler) ListByUser(c *gin.Context) {
	actor, _ := userIDFromCtx(c)
	var f dto.ScheduleFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, total, err := h.uc.ListByUser(c.Request.Context(), actor, f)
	if err != nil {
		writeError(c, err)
		return
	}
	writePagedArray(c, items, total)
}

// GetByID godoc
//
//	@Summary     Get a report schedule
//	@Description Returns a single report schedule by id.
//	@Tags        Compliance Report Schedules
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id  path     int true "Schedule id"
//	@Success     200 {object} dto.ScheduleResponse
//	@Failure     404 {object} map[string]string
//	@Router      /compliance-report-schedules/by-id/{id} [get]
func (h *ScheduleHandler) GetByID(c *gin.Context) {
	id, ok := pathUUID(c, "id")
	if !ok {
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Delete godoc
//
//	@Summary     Delete a report schedule
//	@Description Deletes a report schedule by id.
//	@Tags        Compliance Report Schedules
//	@Security    BearerAuth
//	@Param       id path int true "Schedule id"
//	@Success     204 "No Content"
//	@Failure     404 {object} map[string]string
//	@Router      /compliance-report-schedules/{id} [delete]
func (h *ScheduleHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id")
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "compliance.schedule.delete", ResourceType: "compliance_report_schedule", ResourceID: id.String()},
		audit_domain.COMPLIANCE_SCHEDULE_DELETE_ATTEMPT, audit_domain.COMPLIANCE_SCHEDULE_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
