package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type AssetTypesHandler struct {
	uc connectors.AssetTypesUsecase
}

func NewAssetTypesHandler(uc connectors.AssetTypesUsecase) *AssetTypesHandler {
	return &AssetTypesHandler{uc: uc}
}

func (h *AssetTypesHandler) List(c *gin.Context) {
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)
	items, total, err := h.uc.List(c.Request.Context(), domain.Pagination{
		Page: page.Page, PageSize: page.PageSize, Sort: page.Sort,
	})
	if err != nil {
		writeError(c, "list types", err)
		return
	}
	c.Header("X-Total-Count", uintToStr(uint64(total)))
	c.JSON(http.StatusOK, items)
}
