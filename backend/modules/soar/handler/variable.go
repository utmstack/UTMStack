package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type VariableHandler struct {
	uc connectors.VariableUsecase
}

func NewVariableHandler(uc connectors.VariableUsecase) *VariableHandler {
	return &VariableHandler{uc: uc}
}

// Create godoc
//
//	@Summary     Create incident variable
//	@Description Creates a new utm_incident_variables record. Secret values are encrypted at rest.
//	@Tags        Incident Variables
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.CreateVariableRequest true "Variable to create"
//	@Success     200  {object} dto.VariableResponse
//	@Failure     400  {object} map[string]string
//	@Failure     500  {object} map[string]string
//	@Router      /utm-incident-variables [post]
func (h *VariableHandler) Create(c *gin.Context) {
	var req dto.CreateVariableRequest
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
//	@Summary     Update incident variable
//	@Description Updates an existing utm_incident_variables record
//	@Tags        Incident Variables
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.UpdateVariableRequest true "Variable to update"
//	@Success     200  {object} dto.VariableResponse
//	@Failure     400  {object} map[string]string
//	@Failure     404  {object} map[string]string
//	@Failure     500  {object} map[string]string
//	@Router      /utm-incident-variables [put]
func (h *VariableHandler) Update(c *gin.Context) {
	var req dto.UpdateVariableRequest
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
//	@Summary     List incident variables
//	@Description Returns a paginated list of utm_incident_variables. Secret values are masked as "****".
//	@Tags        Incident Variables
//	@Security    BearerAuth
//	@Produce     json
//	@Param       page         query int    false "Page number (1-based)"
//	@Param       size         query int    false "Page size"
//	@Param       variableName query string false "Filter by variable name"
//	@Success     200  {array}  dto.VariableResponse
//	@Header      200  {string} X-Total-Count "Total records"
//	@Failure     500  {object} map[string]string
//	@Router      /utm-incident-variables [get]
func (h *VariableHandler) List(c *gin.Context) {
	f := dto.VariableFilter{
		Page:         queryInt(c, "page", 1),
		Size:         queryInt(c, "size", 20),
		VariableName: queryString(c, "variableName"),
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
//	@Summary     Get incident variable by ID
//	@Description Returns a single utm_incident_variables record. Secret value masked as "****".
//	@Tags        Incident Variables
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Variable ID"
//	@Success     200 {object} dto.VariableResponse
//	@Failure     404 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /utm-incident-variables/{id} [get]
func (h *VariableHandler) GetByID(c *gin.Context) {
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
//	@Summary     Delete incident variable
//	@Description Deletes a utm_incident_variables record by ID
//	@Tags        Incident Variables
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id path int64 true "Variable ID"
//	@Success     200 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /utm-incident-variables/{id} [delete]
func (h *VariableHandler) Delete(c *gin.Context) {
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
