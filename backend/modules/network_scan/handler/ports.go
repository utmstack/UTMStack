package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type PortsHandler struct {
	uc connectors.PortsUsecase
}

func NewPortsHandler(uc connectors.PortsUsecase) *PortsHandler {
	return &PortsHandler{uc: uc}
}

func (h *PortsHandler) Create(c *gin.Context) {
	var p domain.UtmPorts
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	out, err := h.uc.Create(c.Request.Context(), &p)
	if err != nil {
		writeError(c, "create port", err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *PortsHandler) Update(c *gin.Context) {
	var p domain.UtmPorts
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	out, err := h.uc.Update(c.Request.Context(), &p)
	if err != nil {
		writeError(c, "update port", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *PortsHandler) List(c *gin.Context) {
	var crit domain.PortsCriteria
	if err := c.ShouldBindQuery(&crit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)
	items, total, err := h.uc.ListByCriteria(c.Request.Context(), crit, domain.Pagination{
		Page: page.Page, PageSize: page.PageSize, Sort: page.Sort,
	})
	if err != nil {
		writeError(c, "list ports", err)
		return
	}
	c.Header("X-Total-Count", uintToStr(uint64(total)))
	c.JSON(http.StatusOK, items)
}

func (h *PortsHandler) Count(c *gin.Context) {
	var crit domain.PortsCriteria
	if err := c.ShouldBindQuery(&crit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	total, err := h.uc.CountByCriteria(c.Request.Context(), crit)
	if err != nil {
		writeError(c, "count ports", err)
		return
	}
	c.JSON(http.StatusOK, total)
}

func (h *PortsHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "get port", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *PortsHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		writeError(c, "delete port", err)
		return
	}
	c.Status(http.StatusNoContent)
}
