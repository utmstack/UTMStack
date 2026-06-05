package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type AssetGroupHandler struct {
	uc connectors.AssetGroupUsecase
}

func NewAssetGroupHandler(uc connectors.AssetGroupUsecase) *AssetGroupHandler {
	return &AssetGroupHandler{uc: uc}
}

func (h *AssetGroupHandler) Create(c *gin.Context) {
	var in dto.AssetGroupDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	g := dto.FromAssetGroupDTO(in)
	out, err := h.uc.Create(c.Request.Context(), g)
	if err != nil {
		writeError(c, "create group", err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *AssetGroupHandler) Update(c *gin.Context) {
	var in dto.AssetGroupDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	g := dto.FromAssetGroupDTO(in)
	out, err := h.uc.Update(c.Request.Context(), g)
	if err != nil {
		writeError(c, "update group", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AssetGroupHandler) Search(c *gin.Context) {
	var f domain.AssetGroupFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)
	resp, err := h.uc.SearchByFilter(c.Request.Context(), f, domain.Pagination{
		Page: page.Page, PageSize: page.PageSize, Sort: page.Sort,
	})
	if err != nil {
		writeError(c, "search groups", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AssetGroupHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "get group", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AssetGroupHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		writeError(c, "delete group", err)
		return
	}
	c.Status(http.StatusNoContent)
}
