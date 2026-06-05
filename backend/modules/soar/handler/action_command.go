package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type ActionCommandHandler struct {
	uc connectors.ActionCommandUsecase
}

func NewActionCommandHandler(uc connectors.ActionCommandUsecase) *ActionCommandHandler {
	return &ActionCommandHandler{uc: uc}
}

// Create godoc
//
//	@Summary     Create incident action command
//	@Description Creates a new utm_incident_action_command record
//	@Tags        Incident Action Commands
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.CreateActionCommandRequest true "Action command to create"
//	@Success     200  {object} domain.UtmIncidentActionCommand
//	@Failure     400  {object} map[string]string
//	@Failure     500  {object} map[string]string
//	@Router      /incident-action-commands [post]
func (h *ActionCommandHandler) Create(c *gin.Context) {
	var req dto.CreateActionCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.uc.Create(req)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update godoc
//
//	@Summary     Update incident action command
//	@Description Updates an existing utm_incident_action_command record
//	@Tags        Incident Action Commands
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.UpdateActionCommandRequest true "Action command to update"
//	@Success     200  {object} domain.UtmIncidentActionCommand
//	@Failure     400  {object} map[string]string
//	@Failure     404  {object} map[string]string
//	@Failure     500  {object} map[string]string
//	@Router      /incident-action-commands [put]
func (h *ActionCommandHandler) Update(c *gin.Context) {
	var req dto.UpdateActionCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.uc.Update(req)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// List godoc
//
//	@Summary     List incident action commands
//	@Description Returns a paginated list of utm_incident_action_command
//	@Tags        Incident Action Commands
//	@Security    BearerAuth
//	@Produce     json
//	@Param       page       query int    false "Page number (1-based)"
//	@Param       size       query int    false "Page size"
//	@Param       actionId   query int64  false "Filter by action ID"
//	@Param       osPlatform query string false "Filter by OS platform"
//	@Param       command    query string false "Filter by command text"
//	@Success     200  {array}  domain.UtmIncidentActionCommand
//	@Header      200  {string} X-Total-Count "Total records"
//	@Failure     500  {object} map[string]string
//	@Router      /incident-action-commands [get]
func (h *ActionCommandHandler) List(c *gin.Context) {
	f := dto.ActionCommandFilter{
		Page:       queryInt(c, "page", 1),
		Size:       queryInt(c, "size", 20),
		ActionID:   queryInt64(c, "actionId"),
		OsPlatform: queryString(c, "osPlatform"),
		Command:    queryString(c, "command"),
	}
	items, total, err := h.uc.FindAll(f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	writePagedArray(c, items, total)
}

// GetByID godoc
//
//	@Summary     Get incident action command by ID
//	@Description Returns a single utm_incident_action_command record
//	@Tags        Incident Action Commands
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Action command ID"
//	@Success     200 {object} domain.UtmIncidentActionCommand
//	@Failure     404 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /incident-action-commands/{id} [get]
func (h *ActionCommandHandler) GetByID(c *gin.Context) {
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
//	@Summary     Delete incident action command
//	@Description Deletes a utm_incident_action_command record by ID
//	@Tags        Incident Action Commands
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Action command ID"
//	@Success     200 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /incident-action-commands/{id} [delete]
func (h *ActionCommandHandler) Delete(c *gin.Context) {
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
