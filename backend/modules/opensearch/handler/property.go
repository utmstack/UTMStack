package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

type propertyUsecase interface {
	PropertyValues(ctx context.Context, req dto.PropertyValuesRequest) ([]string, error)
	PropertyValuesWithCount(ctx context.Context, req dto.PropertyValuesWithCountRequest) (map[string]int64, error)
	IndexProperties(ctx context.Context, pattern string) ([]dto.IndexPropertyType, error)
}

type PropertyHandler struct {
	uc propertyUsecase
}

func NewPropertyHandler(uc propertyUsecase) *PropertyHandler {
	return &PropertyHandler{uc: uc}
}

// @Summary     Get field values
// @Description Returns all distinct values for a keyword field across the given index pattern.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Produce     json
// @Param       keyword      query string true "Keyword field name"
// @Param       indexPattern query string true "Index pattern"
// @Success     200 {array}  string
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /opensearch/property/values [get]
func (h *PropertyHandler) Values(c *gin.Context) {
	var req dto.PropertyValuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	values, err := h.uc.PropertyValues(c.Request.Context(), req)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.JSON(http.StatusOK, values)
}

// @Summary     Get field values with document count
// @Description Returns a map of field value → document count for the given index and field.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.PropertyValuesWithCountRequest true "Request body"
// @Success     200 {object} map[string]int64
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /opensearch/property/values-with-count [post]
func (h *PropertyHandler) ValuesWithCount(c *gin.Context) {
	var req dto.PropertyValuesWithCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.uc.PropertyValuesWithCount(c.Request.Context(), req)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
