package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type ProbeHandler struct {
	uc connectors.ProbeUsecase
}

func NewProbeHandler(uc connectors.ProbeUsecase) *ProbeHandler {
	return &ProbeHandler{uc: uc}
}

func (h *ProbeHandler) Scan(c *gin.Context) {
	var in dto.ProbeScanRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	out, err := h.uc.Scan(c.Request.Context(), in)
	if err != nil {
		writeError(c, "probe scan", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ProbeHandler) Ping(c *gin.Context) {
	probeID, ok := queryUint(c, "probeId")
	if !ok {
		return
	}
	if err := h.uc.Ping(c.Request.Context(), probeID); err != nil {
		writeError(c, "probe ping", err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProbeHandler) CheckInterface(c *gin.Context) {
	probeID, ok := queryUint(c, "probeId")
	if !ok {
		return
	}
	iface := c.Query("iface")
	if iface == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "iface is required"})
		return
	}
	ok2, err := h.uc.CheckInterface(c.Request.Context(), probeID, iface)
	if err != nil {
		writeError(c, "probe check_interface", err)
		return
	}
	c.JSON(http.StatusOK, ok2)
}

func queryUint(c *gin.Context, key string) (uint64, bool) {
	s := c.Query(key)
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil || v == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + key})
		return 0, false
	}
	return v, true
}
