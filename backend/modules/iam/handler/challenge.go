package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

// @Summary     Request a password reset email
// @Description Always returns 204 to prevent account enumeration.
// @Tags        Auth
// @Accept      json
// @Param       input body dto.ResetPasswordInitRequest true "Email of the account to reset"
// @Success     204 "Reset email dispatched if the account exists"
// @Failure     400 {object} map[string]string
// @Router      /auth/reset-password/init [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var input dto.ResetPasswordInitRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.authUsecase.RequestPasswordReset(c.Request.Context(), input)
	audit.Record(c, audit_connectors.Event{Action: "auth.reset_password.init"}, audit_domain.RESET_USER_PASSWORD_ATTEMPT, audit_domain.RESET_USER_PASSWORD_SUCCESS, err)
	if errors.Is(err, domain.ErrTfaMailUnavailable) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		_ = catcher.Error("password reset init failed", err, nil)
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Complete a password reset
// @Tags        Auth
// @Accept      json
// @Param       input body dto.ResetPasswordFinishRequest true "Reset key + new password"
// @Success     204 "Password updated"
// @Failure     400 {object} map[string]string
// @Failure     410 {object} map[string]string
// @Router      /auth/reset-password/finish [post]
func (h *AuthHandler) FinishPasswordReset(c *gin.Context) {
	var input dto.ResetPasswordFinishRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.authUsecase.SetPasswordFromChallenge(c.Request.Context(), input)
	audit.Record(c, audit_connectors.Event{Action: "auth.reset_password.finish"}, audit_domain.RESET_USER_PASSWORD_ATTEMPT, audit_domain.RESET_USER_PASSWORD_SUCCESS, err)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, domain.ErrChallengeNotFound):
		c.JSON(http.StatusGone, gin.H{"error": "invalid or expired reset key"})
	default:
		_ = catcher.Error("password reset finish failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not reset password"})
	}
}
