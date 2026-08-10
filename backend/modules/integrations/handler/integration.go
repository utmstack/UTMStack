package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// IntegrationHandler exposes the catalog: the shipped integrations every tenant
// sees plus the ones this tenant added. Their configuration is served elsewhere.
type IntegrationHandler struct {
	usecase connectors.IntegrationUsecase
}

func NewIntegrationHandler(uc connectors.IntegrationUsecase) *IntegrationHandler {
	return &IntegrationHandler{usecase: uc}
}

// @Summary     List integrations
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Param       ingestType     query string  false "Filter by ingest type (agent, collector, forwarder, plugin)"
// @Param       nameContains   query string  false "Partial match on name"
// @Param       page           query int     false "Page (0-based)"
// @Param       size           query int     false "Page size"
// @Success     200 {array}  dto.IntegrationResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations [get]
func (h *IntegrationHandler) List(c *gin.Context) {
	filter := connectors.IntegrationListFilter{}
	if v := c.Query("ingestType"); v != "" {
		it := domain.IngestType(v)
		if !it.Valid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ingestType"})
			return
		}
		filter.IngestType = &it
	}
	if v := c.Query("nameContains"); v != "" {
		filter.NameContains = &v
	}
	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			filter.Page = p
		}
	}
	if v := c.Query("size"); v != "" {
		if s, err := strconv.Atoi(v); err == nil {
			filter.Size = s
		}
	}

	items, total, err := h.usecase.List(c.Request.Context(), filter)
	if err != nil {
		writeIntegrationError(c, "integration.list", err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

// @Summary     Get an integration by ID
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Integration ID"
// @Success     200 {object} dto.IntegrationResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/{id} [get]
func (h *IntegrationHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeIntegrationError(c, "integration.get", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List the data-type catalog
// @Description Each integration owns one data_type; this is the catalog the rule
// @Description editor uses to pick which data types a correlation rule targets.
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array}  dto.DataTypeOption
// @Failure     500 {object} map[string]string
// @Router      /integrations/data-types [get]
func (h *IntegrationHandler) DataTypes(c *gin.Context) {
	resp, err := h.usecase.DataTypes(c.Request.Context())
	if err != nil {
		writeIntegrationError(c, "integration.dataTypes", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Create a custom integration
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateIntegrationRequest true "Custom integration"
// @Success     201 {object} dto.IntegrationResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations [post]
func (h *IntegrationHandler) Create(c *gin.Context) {
	var req dto.CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeIntegrationError(c, "integration.create", err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Update a custom integration's metadata
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string                        true "Integration ID"
// @Param       input body dto.UpdateIntegrationRequest   true "Metadata"
// @Success     200 {object} dto.IntegrationResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/{id} [put]
func (h *IntegrationHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), id, req)
	if err != nil {
		writeIntegrationError(c, "integration.update", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete a custom integration
// @Tags        Integrations
// @Security    BearerAuth
// @Param       id path string true "Integration ID"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/{id} [delete]
func (h *IntegrationHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeIntegrationError(c, "integration.delete", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return uuid.Nil, false
	}
	return id, true
}

// writeIntegrationError maps domain sentinels to HTTP statuses. A request with
// no tenant is the caller's problem, not a server fault, so it answers 400
// rather than turning into an opaque 500.
func writeIntegrationError(c *gin.Context, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrIntegrationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidIngestType):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrSystemIntegration):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, tenancy.ErrNoTenant):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		_ = catcher.Error(op, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
