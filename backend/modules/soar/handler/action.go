package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type ActionHandler struct {
	uc connectors.ActionUsecase
}

func NewActionHandler(uc connectors.ActionUsecase) *ActionHandler {
	return &ActionHandler{uc: uc}
}

// Create godoc
//
//	@Summary     Create incident action
//	@Description Creates a new utm_incident_actions record
//	@Tags        Incident Actions
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.CreateActionRequest true "Action to create"
//	@Success     200  {object} domain.UtmIncidentAction
//	@Failure     500  {object} map[string]string
//	@Router      /utm-incident-actions [post]
func (h *ActionHandler) Create(c *gin.Context) {
	var req dto.CreateActionRequest
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

// Update godoc
//
//	@Summary     Update incident action
//	@Description Updates an existing utm_incident_actions record
//	@Tags        Incident Actions
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.UpdateActionRequest true "Action to update"
//	@Success     200  {object} domain.UtmIncidentAction
//	@Failure     404  {object} map[string]string
//	@Failure     500  {object} map[string]string
//	@Router      /utm-incident-actions [put]
func (h *ActionHandler) Update(c *gin.Context) {
	var req dto.UpdateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.uc.Update(req, loginFromCtx(c))
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// List godoc
//
//	@Summary     List incident actions
//	@Description Returns a paginated list of utm_incident_actions filtered to SHUTDOWN_SERVER and RUN_CMD
//	@Tags        Incident Actions
//	@Security    BearerAuth
//	@Produce     json
//	@Param       page           query int    false "Page number (1-based)"
//	@Param       size           query int    false "Page size"
//	@Param       actionCommand  query string false "Filter by action command"
//	@Param       actionType     query int    false "Filter by action type"
//	@Param       actionEditable query bool   false "Filter by editable flag"
//	@Success     200  {array}  domain.UtmIncidentAction
//	@Header      200  {string} X-Total-Count "Total records"
//	@Failure     500  {object} map[string]string
//	@Router      /utm-incident-actions [get]
func (h *ActionHandler) List(c *gin.Context) {
	f := dto.ActionFilter{
		Page:           queryInt(c, "page", 1),
		Size:           queryInt(c, "size", 20),
		ActionCommand:  queryString(c, "actionCommand"),
		ActionType:     queryIntPtr(c, "actionType"),
		ActionEditable: queryBoolPtr(c, "actionEditable"),
	}
	items, total, err := h.uc.FindAll(f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	// Java parity: filter to only SHUTDOWN_SERVER and RUN_CMD action commands.
	allowed := []string{"SHUTDOWN_SERVER", "RUN_CMD"}
	filtered := items[:0]
	for _, a := range items {
		if a.ActionCommand == nil {
			continue
		}
		for _, ac := range allowed {
			if strings.EqualFold(*a.ActionCommand, ac) {
				filtered = append(filtered, a)
				break
			}
		}
	}
	writePagedArray(c, filtered, total)
}

// Count godoc
//
//	@Summary     Count incident actions
//	@Description Returns the count of utm_incident_actions matching filters
//	@Tags        Incident Actions
//	@Security    BearerAuth
//	@Produce     json
//	@Param       actionCommand  query string false "Filter by action command"
//	@Param       actionType     query int    false "Filter by action type"
//	@Param       actionEditable query bool   false "Filter by editable flag"
//	@Success     200  {integer} int64
//	@Failure     500  {object}  map[string]string
//	@Router      /utm-incident-actions/count [get]
func (h *ActionHandler) Count(c *gin.Context) {
	f := dto.ActionFilter{
		ActionCommand:  queryString(c, "actionCommand"),
		ActionType:     queryIntPtr(c, "actionType"),
		ActionEditable: queryBoolPtr(c, "actionEditable"),
	}
	_, total, err := h.uc.FindAll(f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, total)
}

// GetByID godoc
//
//	@Summary     Get incident action by ID
//	@Description Returns a single utm_incident_actions record
//	@Tags        Incident Actions
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Action ID"
//	@Success     200 {object} domain.UtmIncidentAction
//	@Failure     404 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /utm-incident-actions/{id} [get]
func (h *ActionHandler) GetByID(c *gin.Context) {
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
//	@Summary     Delete incident action
//	@Description Deletes a utm_incident_actions record by ID
//	@Tags        Incident Actions
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Action ID"
//	@Success     200 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /utm-incident-actions/{id} [delete]
func (h *ActionHandler) Delete(c *gin.Context) {
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
