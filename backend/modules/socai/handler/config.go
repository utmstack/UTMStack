package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/socai/dto"
	"github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/modules/socai/usecase"
)

type configService interface {
	Get(ctx context.Context) (*dto.ConfigResponse, error)
	Update(ctx context.Context, req dto.ConfigRequest) (*dto.ConfigResponse, error)
	ResetToDefault(ctx context.Context) (*dto.ConfigResponse, error)
}

type ConfigHandler struct {
	svc configService
}

func NewConfigHandler(svc configService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// Get godoc
//
// @Summary     Get SOC AI configuration
// @Description Returns the current soc-ai plugin configuration with secrets masked.
// @Tags        SOC AI
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.ConfigResponse
// @Failure     500 {object} map[string]string
// @Router      /soc-ai/config [get]
func (h *ConfigHandler) Get(c *gin.Context) {
	resp, err := h.svc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Update godoc
//
// @Summary     Update SOC AI configuration
// @Description Persists the soc-ai plugin configuration (writing the encrypted YAML
// @Description the plugin consumes). Secrets sent as "*****" keep their stored value.
// @Tags        SOC AI
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       config body dto.ConfigRequest true "SOC AI configuration"
// @Success     200 {object} dto.ConfigResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soc-ai/config [put]
func (h *ConfigHandler) Update(c *gin.Context) {
	var req dto.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Update(c.Request.Context(), req)
	if err != nil {
		// A failed connection check is a bad-config (400), not a server error.
		if errors.Is(err, usecase.ErrVerificationFailed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ResetToDefault godoc
//
// @Summary     Stop overriding the SOC AI configuration
// @Description Drops this tenant's own configuration so it goes back to inheriting the instance default. Refused for a caller acting for no tenant, since the default is not an override.
// @Tags        SOC AI
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.ConfigResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soc-ai/config [delete]
func (h *ConfigHandler) ResetToDefault(c *gin.Context) {
	resp, err := h.svc.ResetToDefault(c.Request.Context())
	if err != nil {
		if errors.Is(err, repository.ErrNoInstanceDefault) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
