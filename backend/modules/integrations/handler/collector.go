package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CollectorHandler struct {
	usecase connectors.CollectorUsecase
}

func NewCollectorHandler(uc connectors.CollectorUsecase) *CollectorHandler {
	return &CollectorHandler{usecase: uc}
}

// @Summary     List online forwarder collectors
// @Description Lists ONLINE forwarder-module collectors eligible for remote
// @Description data-type configuration (the picker feed).
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array}  dto.CollectorResponse
// @Failure     503 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/collectors [get]
func (h *CollectorHandler) ListForwarders(c *gin.Context) {
	items, err := h.usecase.ListOnlineForwarders(c.Request.Context())
	if err != nil {
		writeCollectorError(c, "collector.listForwarders", err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// @Summary     Enable or disable a data-type integration on a forwarder
// @Description Pushes a remote config to the forwarder collector via
// @Description agent-manager's SetCollectorConfig RPC and blocks until the
// @Description collector acknowledges (or the request times out).
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id       path int                        true "Collector ID"
// @Param       dataType path string                     true "Data type / integration name"
// @Param       body     body dto.SetDataTypeConfigRequest true "Desired config"
// @Success     200 {object} dto.ConfigKnowledgeResponse
// @Failure     400 {object} map[string]string
// @Failure     503 {object} map[string]string
// @Failure     504 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/collectors/{id}/data-types/{dataType} [put]
func (h *CollectorHandler) SetDataType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collector id"})
		return
	}
	dataType := c.Param("dataType")
	if dataType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing data type"})
		return
	}

	var req dto.SetDataTypeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.SetDataTypeConfig(c.Request.Context(), uint32(id), dataType, req)
	if err != nil {
		writeCollectorError(c, "collector.setDataType", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Push cert/key/(optional CA) certificates to a forwarder
// @Description Forwards a base64-encoded PEM cert/key/(optional CA) triple
// @Description to agent-manager's reserved __tls_certs__ SetCollectorConfig
// @Description channel (action:"apply") and blocks until the forwarder
// @Description acknowledges. The backend never encrypts this payload —
// @Description agent-manager seals it before relaying to the forwarder.
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path int                                  true "Collector ID"
// @Param       body body dto.SetForwarderCertificatesRequest true "Base64-encoded PEM cert/key/(optional) CA"
// @Success     200 {object} dto.ConfigKnowledgeResponse
// @Failure     400 {object} map[string]string
// @Failure     503 {object} map[string]string
// @Failure     504 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/collectors/{id}/certificates [put]
func (h *CollectorHandler) SetCertificates(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collector id"})
		return
	}

	var req dto.SetForwarderCertificatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.SetForwarderCertificates(c.Request.Context(), uint32(id), req)
	if err != nil {
		writeCollectorError(c, "collector.setCertificates", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Get a forwarder collector's current TLS certificate status
// @Description Queries the forwarder's on-disk TLS certificate state via
// @Description the same reserved __tls_certs__ channel (action:"status").
// @Description No secret material travels in the response.
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Collector ID"
// @Success     200 {object} dto.TLSStatusResponse
// @Failure     400 {object} map[string]string
// @Failure     503 {object} map[string]string
// @Failure     504 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/collectors/{id}/tls-status [get]
func (h *CollectorHandler) GetTLSStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collector id"})
		return
	}

	resp, err := h.usecase.GetTLSStatus(c.Request.Context(), uint32(id))
	if err != nil {
		writeCollectorError(c, "collector.getTLSStatus", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// writeCollectorError maps domain sentinels and gRPC status codes returned
// by the agent-manager relay to HTTP statuses:
//   - domain.ErrInvalidCollectorConfig      -> 400 (missing proto/port)
//   - domain.ErrAgentManagerUnavailable     -> 503 (gRPC dial never succeeded)
//   - codes.InvalidArgument                 -> 400 (message sanitized, see below)
//   - codes.Unavailable ("collector offline") -> 503
//   - codes.DeadlineExceeded ("no ack in time") -> 504
//   - anything else                         -> 500 (logged via catcher)
func writeCollectorError(c *gin.Context, op string, err error) {
	if errors.Is(err, domain.ErrInvalidCollectorConfig) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrAgentManagerUnavailable) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration: " + status.Convert(err).Message()})
	case codes.Unavailable:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "collector offline"})
	case codes.DeadlineExceeded:
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "collector did not confirm in time"})
	default:
		_ = catcher.Error(op, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
