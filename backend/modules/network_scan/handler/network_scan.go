package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type NetworkScanHandler struct {
	uc     connectors.NetworkScanUsecase
	report connectors.ReportUsecase
}

func NewNetworkScanHandler(uc connectors.NetworkScanUsecase, report connectors.ReportUsecase) *NetworkScanHandler {
	return &NetworkScanHandler{uc: uc, report: report}
}

func (h *NetworkScanHandler) SaveCustom(c *gin.Context) {
	var input dto.NetworkScanDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.uc.SaveOrUpdateCustomAsset(c.Request.Context(), input); err != nil {
		writeError(c, "save custom asset", err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *NetworkScanHandler) Update(c *gin.Context) {
	var input domain.UtmNetworkScan
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	out, err := h.uc.Update(c.Request.Context(), &input)
	if err != nil {
		writeError(c, "update asset", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *NetworkScanHandler) UpdateType(c *gin.Context) {
	var input dto.UpdateTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.uc.UpdateType(c.Request.Context(), input); err != nil {
		writeError(c, "update type", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NetworkScanHandler) UpdateGroup(c *gin.Context) {
	var input dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.uc.UpdateGroup(c.Request.Context(), input); err != nil {
		writeError(c, "update group", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NetworkScanHandler) ListByCriteria(c *gin.Context) {
	var crit domain.NetworkScanCriteria
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
		writeError(c, "list", err)
		return
	}
	c.Header("X-Total-Count", uintToStr(uint64(total)))
	c.JSON(http.StatusOK, items)
}

func (h *NetworkScanHandler) CountByCriteria(c *gin.Context) {
	var crit domain.NetworkScanCriteria
	if err := c.ShouldBindQuery(&crit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	total, err := h.uc.CountByCriteria(c.Request.Context(), crit)
	if err != nil {
		writeError(c, "count", err)
		return
	}
	c.JSON(http.StatusOK, total)
}

func (h *NetworkScanHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "get", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *NetworkScanHandler) SearchByFilters(c *gin.Context) {
	var f domain.NetworkScanFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)
	resp, err := h.uc.SearchByFilters(c.Request.Context(), f, domain.Pagination{
		Page: page.Page, PageSize: page.PageSize, Sort: page.Sort,
	})
	if err != nil {
		writeError(c, "search", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *NetworkScanHandler) SearchPropertyValues(c *gin.Context) {
	prop := c.Query("property")
	value := c.Query("value")
	forGroups := c.Query("forGroups") == "true"
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)
	values, err := h.uc.SearchPropertyValues(c.Request.Context(), prop, value, forGroups,
		domain.Pagination{Page: page.Page, PageSize: page.PageSize})
	if err != nil {
		writeError(c, "property values", err)
		return
	}
	c.JSON(http.StatusOK, values)
}

func (h *NetworkScanHandler) CountNewAssets(c *gin.Context) {
	total, err := h.uc.CountNewAssets(c.Request.Context())
	if err != nil {
		writeError(c, "count new", err)
		return
	}
	c.JSON(http.StatusOK, total)
}

func (h *NetworkScanHandler) DeleteCustom(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := h.uc.DeleteCustomAsset(c.Request.Context(), id); err != nil {
		writeError(c, "delete", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NetworkScanHandler) CanRunCommand(c *gin.Context) {
	name := c.Query("assetName")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assetName is required"})
		return
	}
	ok, err := h.uc.CanRunCommand(c.Request.Context(), name)
	if err != nil {
		writeError(c, "can-run-command", err)
		return
	}
	c.JSON(http.StatusOK, ok)
}

func (h *NetworkScanHandler) AgentOSPlatform(c *gin.Context) {
	out, err := h.uc.GetAgentsOSPlatform(c.Request.Context())
	if err != nil {
		writeError(c, "agent-os-platform", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *NetworkScanHandler) Report(c *gin.Context) {
	var f domain.NetworkScanFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	body, err := h.report.GenerateNetworkScanReport(c.Request.Context(), f)
	if err != nil {
		writeError(c, "report", err)
		return
	}
	// PDF rendering is deferred — return HTML; the frontend can print-to-PDF or pipe
	// to the existing web-pdf microservice. See usecase/report.go for context.
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", `inline; filename="network-scan-report.html"`)
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(body)
}

func uintToStr(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
