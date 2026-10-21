package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/threatintel/dto"
	"github.com/utmstack/utmstack/backend/modules/threatintel/usecase"
)

type FeedsHandler struct{ uc *usecase.FeedsService }

func NewFeedsHandler(uc *usecase.FeedsService) *FeedsHandler { return &FeedsHandler{uc: uc} }

// Status godoc
//
//	@Summary		Whether this instance contributes to the ThreatWinds feed
//	@Description	Answers whether sending is on and whether the credentials are in place. The credentials themselves are never returned.
//	@Tags			Threat Intel
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.FeedsStatus
//	@Router			/threat-intel/feeds/contribution [get]
func (h *FeedsHandler) Status(c *gin.Context) {
	status, err := h.uc.Status()
	if err != nil {
		_ = catcher.Error("threatintel: reading the feeds configuration failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read the configuration"})
		return
	}
	c.JSON(http.StatusOK, status)
}

// SetEnabled godoc
//
//	@Summary		Turn the ThreatWinds contribution on or off
//	@Description	Decides whether this instance's incidents are sent to ThreatWinds. Credentials are kept either way.
//	@Tags			Threat Intel
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.FeedsToggleRequest	true	"On or off"
//	@Success		200		{object}	dto.FeedsStatus
//	@Router			/threat-intel/feeds/contribution [put]
func (h *FeedsHandler) SetEnabled(c *gin.Context) {
	var req dto.FeedsToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.SetEnabled(req); err != nil {
		_ = catcher.Error("threatintel: saving the feeds configuration failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save the configuration"})
		return
	}
	h.Status(c)
}

