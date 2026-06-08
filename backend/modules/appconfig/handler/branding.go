package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	"github.com/utmstack/utmstack/backend/modules/appconfig/usecase"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
)

type BrandingHandler struct {
	usecase connectors.BrandingUsecase
}

func NewBrandingHandler(uc connectors.BrandingUsecase) *BrandingHandler {
	return &BrandingHandler{usecase: uc}
}

// Get godoc
//
//	@Summary		Get white-label branding
//	@Description	Returns the current branding overrides (admin). Falls back to defaults if unset.
//	@Tags			Branding
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.BrandingResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/branding [get]
func (h *BrandingHandler) Get(c *gin.Context) {
	resp, err := h.usecase.Get(c.Request.Context())
	if err != nil {
		_ = catcher.Error("branding get failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get branding"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Update godoc
//
//	@Summary		Update white-label branding
//	@Description	Updates logo, product name, colors, etc. Enabling white-labeling is a paid feature (returns 402 when not entitled).
//	@Tags			Branding
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.BrandingRequest	true	"Branding overrides"
//	@Success		200		{object}	dto.BrandingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		402		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/branding [put]
func (h *BrandingHandler) Update(c *gin.Context) {
	var req dto.BrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "branding.updated", ResourceType: "branding", ResourceID: "branding"},
		audit_domain.CONFIG_CHANGED, audit_domain.CONFIG_CHANGED, err)
	if errors.Is(err, usecase.ErrWhiteLabelNotEntitled) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		_ = catcher.Error("branding update failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save branding"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Public godoc
//
//	@Summary		Get effective branding (public)
//	@Description	Unauthenticated branding for the login page. Returns the default brand when white-labeling is disabled or not entitled.
//	@Tags			Branding
//	@Produce		json
//	@Success		200	{object}	dto.BrandingPublic
//	@Failure		500	{object}	map[string]string
//	@Router			/branding/public [get]
func (h *BrandingHandler) Public(c *gin.Context) {
	resp, err := h.usecase.GetPublic(c.Request.Context())
	if err != nil {
		_ = catcher.Error("branding public failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get branding"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
