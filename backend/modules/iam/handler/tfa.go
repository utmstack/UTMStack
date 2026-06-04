package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/logger"
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
func (h *TfaHandler) InitEnrollment(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input dto.TfaInitRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method is required (EMAIL or TOTP)"})
		return
	}
	resp, err := h.tfa.InitEnrollment(c.Request.Context(), uid, input.Method)
	audit.Record(c, audit_connectors.Event{Action: "tfa.init"}, audit_domain.TFA_CODE_SENT, audit_domain.TFA_CODE_SENT, err)
	if err != nil {
		writeTfaError(c, err, "tfa init failed")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Verify the TFA code during setup
// @Tags        TFA
// @Accept      json
// @Produce     json
// @Param       input body dto.TfaVerifyRequest true "Code"
// @Success     204 ""
// @Router      /tfa/verify [post]
func (h *TfaHandler) VerifyEnrollment(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input dto.TfaVerifyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method and 6-digit code are required"})
		return
	}
	err := h.tfa.VerifyEnrollment(c.Request.Context(), uid, input.Method, input.Code)
	audit.Record(c, audit_connectors.Event{Action: "tfa.verify"}, audit_domain.TFA_CODE_VERIFY_ATTEMPT, audit_domain.TFA_VERIFIED, err)
	if err != nil {
		writeTfaError(c, err, "tfa verify failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Complete TFA enrollment and persist the secret
// @Tags        TFA
// @Accept      json
// @Produce     json
// @Param       input body dto.TfaCompleteRequest true "Method"
// @Success     204 ""
// @Router      /tfa/complete [post]
func (h *TfaHandler) CompleteEnrollment(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input dto.TfaCompleteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method is required"})
		return
	}
	err := h.tfa.CompleteEnrollment(c.Request.Context(), uid, input.Method)
	audit.Record(c, audit_connectors.Event{Action: "tfa.complete"}, audit_domain.TFA_ENABLE_ATTEMPT, audit_domain.TFA_ENABLE_SUCCESS, err)
	if err != nil {
		writeTfaError(c, err, "tfa complete failed")
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Resend the TFA challenge (EMAIL only)
// @Tags        TFA
// @Produce     json
// @Param       method query string true "EMAIL"
// @Success     200 {object} dto.TfaRefreshResponse
// @Router      /tfa/refresh [get]
func (h *TfaHandler) RefreshChallenge(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	method := c.Query("method")
	if method == "" {
		method = "EMAIL"
	}
	resp, err := h.tfa.RefreshChallenge(c.Request.Context(), uid, method)
	if err != nil {
		if errors.Is(err, domain.ErrTfaCooldown) && resp != nil {
			c.JSON(http.StatusTooManyRequests, resp)
			return
		}
		writeTfaError(c, err, "tfa refresh failed")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Verify TFA at login and exchange the pre-auth token for a JWT
// @Tags        TFA
// @Accept      json
// @Produce     json
// @Param       input body dto.TfaVerifyCodeRequest true "Pre-auth token + code"
// @Success     200 {object} dto.LoginResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /auth/tfa/verify-code [post]
func (h *TfaHandler) VerifyLoginCode(c *gin.Context) {
	var input dto.TfaVerifyCodeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pre_auth_token and code are required"})
		return
	}
	resp, err := h.tfa.VerifyLoginCode(c.Request.Context(), input, loginContext(c))
	audit.Record(c, audit_connectors.Event{Action: "tfa.login.verify"}, audit_domain.TFA_CODE_VERIFY_ATTEMPT, audit_domain.TFA_VERIFIED, err)
	if err != nil {
		writeTfaError(c, err, "tfa login verification failed")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Unified 3-stage TFA enrollment endpoint
// @Tags        TFA
// @Accept      json
// @Produce     json
// @Param       input body dto.TfaEnrollmentRequest true "Stage + method (+ code for VERIFY)"
// @Success     200 {object} dto.TfaEnrollmentResponse
// @Router      /enrollment/tfa [post]
func (h *TfaHandler) UnifiedEnrollment(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input dto.TfaEnrollmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stage and method are required"})
		return
	}
	resp, err := h.tfa.UnifiedEnrollment(c.Request.Context(), uid, input)
	audit.Record(c, audit_connectors.Event{Action: "tfa.enrollment"}, audit_domain.TFA_ENABLE_ATTEMPT, audit_domain.TFA_ENABLE_SUCCESS, err)
	if err != nil {
		writeTfaError(c, err, "tfa enrollment failed")
		return
	}
	c.JSON(http.StatusOK, resp)
}

func writeTfaError(c *gin.Context, err error, fallbackMsg string) {
	switch {
	case errors.Is(err, domain.ErrTfaDisabled):
		c.JSON(http.StatusForbidden, gin.H{"error": "tfa is disabled"})
	case errors.Is(err, domain.ErrTfaInvalidPreAuth):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired pre-auth token"})
	case errors.Is(err, domain.ErrTfaMethodMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": "method does not match pre-auth token"})
	case errors.Is(err, domain.ErrTfaChallengeNotFound):
		c.JSON(http.StatusGone, gin.H{"error": "tfa challenge not found or expired"})
	case errors.Is(err, domain.ErrTfaInvalidCode):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired code"})
	case errors.Is(err, domain.ErrTfaCooldown):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests; try again shortly"})
	case errors.Is(err, domain.ErrTfaMethodUnsupported):
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported tfa method"})
	case errors.Is(err, domain.ErrTfaNotVerified):
		c.JSON(http.StatusBadRequest, gin.H{"error": "verify the code before completing enrollment"})
	case errors.Is(err, domain.ErrTfaAlreadyEnabled):
		c.JSON(http.StatusConflict, gin.H{"error": "tfa is already enabled for this user"})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	default:
		logger.Error(fallbackMsg + ": " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMsg})
	}
}
