package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/collectors/connectors"
	"github.com/utmstack/utmstack/backend/modules/collectors/dto"
	"github.com/utmstack/utmstack/backend/modules/collectors/usecase"
)

type CollectorHandler struct {
	uc connectors.CollectorUsecase
}

func NewCollectorHandler(uc connectors.CollectorUsecase) *CollectorHandler {
	return &CollectorHandler{uc: uc}
}

// ListCollectors godoc
// @Summary     List collectors
// @Description Syncs collectors from agent-manager and returns the local dataset.
// @Tags        Collectors
// @Security    BearerAuth
// @Produce     json
// @Param       pageNumber query int    false "Page number (1-based)"
// @Param       pageSize   query int    false "Page size (default 10)"
// @Param       hostname   query string false "Filter by hostname"
// @Param       module     query string false "Filter by module (AS_400 or UTMSTACK)"
// @Param       sortBy     query string false "Sort field"
// @Success     200 {array}  dto.CollectorResponse
// @Header      200 {string} X-Total-Count "Total number of collectors"
// @Failure     500 {object} map[string]string
// @Router      /collectors [get]
func (h *CollectorHandler) ListCollectors(c *gin.Context) {
	var q dto.CollectorListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.uc.ListAndSync(c.Request.Context(), q)
	if err != nil {
		writeCollectorError(c, err)
		return
	}

	// W1: return JSON envelope {"collectors":[...],"total":N} matching Java ListCollectorsResponseDTO.
	// Also set X-Total-Count header for backward compatibility.
	c.Header("X-Total-Count", fmt.Sprintf("%d", resp.Total))
	if resp.Collectors == nil {
		resp.Collectors = []dto.CollectorResponse{}
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteCollector godoc
// @Summary     Delete a collector
// @Description Deletes a collector from the local database only (no gRPC call).
// @Tags        Collectors
// @Security    BearerAuth
// @Param       id path int true "Collector ID"
// @Success     200
// @Failure     500 {object} map[string]string
// @Router      /collectors/{id} [delete]
func (h *CollectorHandler) DeleteCollector(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		// W3: return 404 when the collector does not exist — mirrors Java ApiException(NOT_FOUND).
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		writeCollectorError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
