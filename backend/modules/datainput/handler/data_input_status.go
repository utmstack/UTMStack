package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/datainput/connectors"
	"github.com/utmstack/utmstack/backend/modules/datainput/dto"
)

type DataInputStatusHandler struct {
	usecase connectors.DataInputStatusUsecase
}

func NewDataInputStatusHandler(uc connectors.DataInputStatusUsecase) *DataInputStatusHandler {
	return &DataInputStatusHandler{usecase: uc}
}

// @Summary     Create data input status
// @Tags        Data Input Status
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateDataInputStatusRequest true "Data input status to create (id must be absent)"
// @Success     201 {object} dto.DataInputStatusResponse
// @Header      201 {string} Location "/api/v1/utm-data-input-statuses/{id}"
// @Failure     400 {object} map[string]string "idexists — id field must not be present on create"
// @Failure     500 {object} map[string]string
// @Router      /utm-data-input-statuses [post]
func (h *DataInputStatusHandler) Create(c *gin.Context) {
	var req dto.CreateDataInputStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeDataInputError(c, err)
		return
	}

	c.Header("Location", "/api/v1/utm-data-input-statuses/"+resp.ID)
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Update data input status
// @Tags        Data Input Status
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateDataInputStatusRequest true "Data input status to update (id required)"
// @Success     200 {object} dto.DataInputStatusResponse
// @Failure     400 {object} map[string]string "idnull — id field is required for update"
// @Failure     500 {object} map[string]string
// @Router      /utm-data-input-statuses [put]
func (h *DataInputStatusHandler) Update(c *gin.Context) {
	var req dto.UpdateDataInputStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idnull"})
		return
	}

	resp, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeDataInputError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List data input statuses (filtered view — errors return 200 with null body)
// @Tags        Data Input Status
// @Security    BearerAuth
// @Produce     json
// @Param       page query int false "Page (0-based)"
// @Param       size query int false "Page size"
// @Success     200 {array}  dto.DataInputStatusResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Router      /utm-data-input-statuses [get]
func (h *DataInputStatusHandler) List(c *gin.Context) {
	var f dto.DataInputStatusListFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeDataInputError(c, err)
		return
	}

	if items == nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	writePagedArray(c, items, total)
}

// @Summary     Count data input statuses
// @Tags        Data Input Status
// @Security    BearerAuth
// @Produce     json
// @Param       id.equals               query string false "Filter by exact id"
// @Param       id.contains             query string false "Filter by id containing value"
// @Param       source.equals           query string false "Filter by exact source"
// @Param       source.contains         query string false "Filter by source containing value"
// @Param       dataType.equals         query string false "Filter by exact dataType"
// @Param       dataType.contains       query string false "Filter by dataType containing value"
// @Param       timestamp.equals        query int    false "Filter by exact timestamp"
// @Param       timestamp.greaterThan   query int    false "Filter by timestamp greater than"
// @Param       timestamp.lessThan      query int    false "Filter by timestamp less than"
// @Param       median.equals           query int    false "Filter by exact median"
// @Param       median.greaterThan      query int    false "Filter by median greater than"
// @Param       median.lessThan         query int    false "Filter by median less than"
// @Success     200 {integer} int64
// @Router      /utm-data-input-statuses/count [get]
func (h *DataInputStatusHandler) Count(c *gin.Context) {
	var f dto.DataInputStatusCountFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.usecase.Count(c.Request.Context(), f)
	if err != nil {
		writeDataInputError(c, err)
		return
	}
	c.JSON(http.StatusOK, count)
}

// @Summary     Get data input status by ID
// @Tags        Data Input Status
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Composite key (dataType-source)"
// @Success     200 {object} dto.DataInputStatusResponse
// @Failure     404 {object} map[string]string
// @Router      /utm-data-input-statuses/{id} [get]
func (h *DataInputStatusHandler) GetByID(c *gin.Context) {
	id, ok := pathStringID(c, "id")
	if !ok {
		return
	}

	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeDataInputError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete data input status
// @Tags        Data Input Status
// @Security    BearerAuth
// @Param       id path string true "Composite key (dataType-source)"
// @Success     200 "Deleted"
// @Failure     500 {object} map[string]string
// @Router      /utm-data-input-statuses/{id} [delete]
func (h *DataInputStatusHandler) Delete(c *gin.Context) {
	id, ok := pathStringID(c, "id")
	if !ok {
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeDataInputError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
