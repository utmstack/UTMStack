package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/connectors"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/domain"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/dto"
	"github.com/utmstack/utmstack/backend/modules/asset_metrics/repository"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type AssetMetricHandler struct {
	usecase connectors.AssetMetricUsecase
}

func NewAssetMetricHandler(uc connectors.AssetMetricUsecase) *AssetMetricHandler {
	return &AssetMetricHandler{usecase: uc}
}

// @Summary     Create asset metric
// @Tags        Asset Metrics
// @Accept      json
// @Produce     json
// @Param       input body dto.AssetMetricRequest true "Asset metric to create (id must be absent or empty)"
// @Success     201 {object} dto.AssetMetricResponse
// @Header      201 {string} Location "/api/v1/utm-asset-metrics/{id}"
// @Failure     400 {object} map[string]string "idexists — id field must not be present on create"
// @Failure     500 {object} map[string]string
// @Router      /utm-asset-metrics [post]
func (h *AssetMetricHandler) Create(c *gin.Context) {
	var req dto.AssetMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idexists"})
		return
	}

	m := domain.UtmAssetMetric{
		ID:        req.ID,
		AssetName: req.AssetName,
		Metric:    req.Metric,
		Amount:    req.Amount,
	}
	if err := h.usecase.Save(c.Request.Context(), m); err != nil {
		writeError(c, err)
		return
	}

	resp := toResponse(m)
	c.Header("Location", "/api/v1/utm-asset-metrics/"+resp.ID)
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Update asset metric
// @Tags        Asset Metrics
// @Accept      json
// @Produce     json
// @Param       input body dto.AssetMetricRequest true "Asset metric to update (id required)"
// @Success     200 {object} dto.AssetMetricResponse
// @Failure     400 {object} map[string]string "idnull — id field is required for update"
// @Failure     500 {object} map[string]string
// @Router      /utm-asset-metrics [put]
func (h *AssetMetricHandler) Update(c *gin.Context) {
	var req dto.AssetMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idnull"})
		return
	}

	m := domain.UtmAssetMetric{
		ID:        req.ID,
		AssetName: req.AssetName,
		Metric:    req.Metric,
		Amount:    req.Amount,
	}
	if err := h.usecase.Update(c.Request.Context(), m); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(m))
}

// @Summary     List all asset metrics (unpaginated)
// @Tags        Asset Metrics
// @Produce     json
// @Success     200 {array}  dto.AssetMetricResponse
// @Failure     500 {object} map[string]string
// @Router      /utm-asset-metrics [get]
func (h *AssetMetricHandler) ListAll(c *gin.Context) {
	items, err := h.usecase.FindAll(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []domain.UtmAssetMetric{}
	}
	resp := make([]dto.AssetMetricResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toResponse(item))
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Get asset metric by ID
// @Tags        Asset Metrics
// @Produce     json
// @Param       id path string true "Asset metric ID"
// @Success     200 {object} dto.AssetMetricResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-asset-metrics/{id} [get]
func (h *AssetMetricHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	m, err := h.usecase.FindByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset metric not found"})
			return
		}
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(*m))
}

// @Summary     Delete asset metric
// @Tags        Asset Metrics
// @Param       id path string true "Asset metric ID"
// @Success     200 "Deleted"
// @Failure     500 {object} map[string]string
// @Router      /utm-asset-metrics/{id} [delete]
func (h *AssetMetricHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func writeError(c *gin.Context, err error) {
	logger.Error("asset_metrics: operation failed: " + err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
}

func toResponse(m domain.UtmAssetMetric) dto.AssetMetricResponse {
	return dto.AssetMetricResponse{
		ID:        m.ID,
		AssetName: m.AssetName,
		Metric:    m.Metric,
		Amount:    m.Amount,
	}
}
