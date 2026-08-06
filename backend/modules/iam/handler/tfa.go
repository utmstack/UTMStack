package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type TfaHandler struct {
	tfa connectors.TfaUsecase
}

func NewTfaHandler(tfa connectors.TfaUsecase) *TfaHandler {
	return &TfaHandler{tfa: tfa}
}

// @Summary     Start TFA enrollment
// @Tags        TFA
// @Accept      json
// @Produce     json
// @Param       input body dto.TfaInitRequest true "Method"
// @Success     200 {object} dto.TfaInitResponse
// @Router      /tfa/init [post]
func (h *TfaHandler) Enroll(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	var req dto.TfaEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.tfa.Enroll(c.Request.Context(), uid, req)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TfaHandler) Disable(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	var req dto.TfaDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.tfa.Disable(c.Request.Context(), uid, req.Password); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TfaHandler) VerifyLoginCode(c *gin.Context) {
	var req dto.TfaVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.tfa.VerifyLoginCode(c.Request.Context(), req, loginContext(c))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TfaHandler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTfaInvalidCode), errors.Is(err, domain.ErrTfaInvalidPreAuth):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrTfaMailUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrChallengeCooldown):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrChallengeExpired), errors.Is(err, domain.ErrChallengeNotFound),
		errors.Is(err, domain.ErrTfaFactorNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrTfaFactorExists), errors.Is(err, domain.ErrTfaTypeUnsupported),
		errors.Is(err, domain.ErrCurrentPassword):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
