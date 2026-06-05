package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
)

type DataTypeHandler struct {
	usecase connectors.DataTypeUsecase
}

func NewDataTypeHandler(uc connectors.DataTypeUsecase) *DataTypeHandler {
	return &DataTypeHandler{usecase: uc}
}

// @Summary     Create data type
// @Tags        Data Types
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateDataTypeRequest true "Data type to create"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /data-types [post]
func (h *DataTypeHandler) Create(c *gin.Context) {
	var req dto.CreateDataTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.usecase.Create(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Update data type
// @Tags        Data Types
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateDataTypeRequest true "Data type to update"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /data-types [put]
func (h *DataTypeHandler) Update(c *gin.Context) {
	var req dto.UpdateDataTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.usecase.Update(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Update include/exclude list for data types
// @Tags        Data Types
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body []dto.UpdateIncludeExcludeItem true "List of items to update"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /data-types/include-exclude-list [put]
func (h *DataTypeHandler) UpdateIncludeExcludeList(c *gin.Context) {
	var items []dto.UpdateIncludeExcludeItem
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.UpdateIncludeExcludeList(c.Request.Context(), items); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     List data types
// @Tags        Data Types
// @Security    BearerAuth
// @Produce     json
// @Param       page   query int    false "Page (0-based)"
// @Param       size   query int    false "Page size"
// @Param       search query string false "Partial match on data type name or description"
// @Success     200 {array}  dto.DataTypeResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Failure     500 {object} map[string]string
// @Router      /data-types [get]
func (h *DataTypeHandler) List(c *gin.Context) {
	var f dto.DataTypeFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	writePagedArray(c, result.Items, result.Total)
}

// @Summary     Get data type by ID
// @Tags        Data Types
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Data type ID"
// @Success     200 {object} dto.DataTypeResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /data-types/{id} [get]
func (h *DataTypeHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		writeCorrelationError(c, err)
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Delete data type
// @Tags        Data Types
// @Security    BearerAuth
// @Param       id path int true "Data type ID"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string "systemOwner data type cannot be deleted"
// @Failure     500 {object} map[string]string
// @Router      /data-types/{id} [delete]
func (h *DataTypeHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
