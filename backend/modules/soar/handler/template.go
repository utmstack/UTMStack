package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type TemplateHandler struct {
	usecase connectors.TemplateUsecase
}

func NewTemplateHandler(uc connectors.TemplateUsecase) *TemplateHandler {
	return &TemplateHandler{usecase: uc}
}

// @Summary     List alert response action templates
// @Tags        SOAR Templates
// @Security    BearerAuth
// @Produce     json
// @Param       page query int false "Page number (0-based)"
// @Param       size query int false "Page size"
// @Success     200 {array}  dto.TemplateResponse
// @Header      200 {integer} X-Total-Count "Total number of records"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/action-templates [get]
func (h *TemplateHandler) List(c *gin.Context) {
	var f dto.TemplateFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	writePagedArray(c, result.Items, result.Total)
}
