package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

const (
	avatarMaxBytes = 2 << 20 // 2 MiB
	avatarSubdir   = "avatars"
)

var avatarAllowedTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type AuthHandler struct {
	authUsecase connectors.AuthUsecase
	uploadDir   string
}

func NewAuthHandler(authUsecase connectors.AuthUsecase, uploadDir string) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, uploadDir: uploadDir}
}

func loginContext(c *gin.Context) connectors.LoginContext {
	return connectors.LoginContext{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

// @Summary     Authenticate user
// @Description The `login` field is matched against `users.login` OR `users.email`.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       input body dto.LoginRequest true "Credentials"
// @Success     200 {object} dto.LoginResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     429 {object} map[string]string
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login and password are required"})
		return
	}
	resp, err := h.authUsecase.Login(c.Request.Context(), input, loginContext(c))
	audit.Record(c, audit_connectors.Event{Action: "auth.login"}, audit_domain.AUTH_ATTEMPT, audit_domain.AUTH_SUCCESS, err)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLoginBlocked):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many failed attempts; try again later"})
		case errors.Is(err, domain.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		case errors.Is(err, domain.ErrUserDeactivated):
			c.JSON(http.StatusForbidden, gin.H{"error": "user is deactivated"})
		case errors.Is(err, domain.ErrTfaNoEmail):
			c.JSON(http.StatusForbidden, gin.H{"error": "tfa is required but no email is set for this user"})
		default:
			logger.Error("login failed: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Refresh access token
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       input body dto.RefreshRequest true "Current refresh token"
// @Success     200 {object} dto.TokenPair
// @Failure     401 {object} map[string]string
// @Router      /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var input dto.RefreshRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}
	pair, err := h.authUsecase.Refresh(c.Request.Context(), input, loginContext(c))
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefresh) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
			return
		}
		logger.Error("refresh failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not refresh"})
		return
	}
	c.JSON(http.StatusOK, pair)
}

// @Summary     Revoke refresh token
// @Tags        Auth
// @Accept      json
// @Param       input body dto.LogoutRequest true "Refresh token to revoke"
// @Success     204 "Logged out"
// @Router      /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var input dto.LogoutRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}
	err := h.authUsecase.Logout(c.Request.Context(), input)
	audit.Record(c, audit_connectors.Event{Action: "auth.logout"}, audit_domain.AUTH_LOGOUT, audit_domain.AUTH_LOGOUT, err)
	if err != nil {
		logger.Error("logout failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not logout"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Current user
// @Tags        Auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.UserResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	id, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	resp, err := h.authUsecase.Me(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Current user
// @Tags        Auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 string
// @Failure     401 {object} map[string]string
// @Router      /auth/authenticate [get]
func (h *AuthHandler) Authenticate(c *gin.Context) {
	id, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	resp, err := h.authUsecase.Me(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.String(http.StatusOK, resp.Login)
}

// @Summary     Update own profile
// @Tags        Auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateMeRequest true "Fields to update"
// @Success     200 {object} dto.UserResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /auth/me [put]
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	id, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input dto.UpdateMeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.authUsecase.UpdateMe(c.Request.Context(), id, input)
	audit.Record(c, audit_connectors.Event{Action: "auth.me.update"}, audit_domain.USER_UPDATE_ATTEMPT, audit_domain.USER_UPDATE_SUCCESS, err)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "email is already in use"})
		default:
			logger.Error("update me failed: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update profile"})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List active sessions for the current user
// @Tags        Auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} dto.SessionResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/sessions [get]
func (h *AuthHandler) ListSessions(c *gin.Context) {
	id, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sessions, err := h.authUsecase.ListSessions(c.Request.Context(), id, sessionIDFromCtx(c))
	if err != nil {
		logger.Error("list sessions failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list sessions"})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// @Summary     Revoke a specific session
// @Tags        Auth
// @Security    BearerAuth
// @Param       id path int true "Session ID"
// @Success     204 "Revoked"
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /auth/sessions/{id} [delete]
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	if err := h.authUsecase.RevokeSession(c.Request.Context(), uid, sid); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		logger.Error("revoke session failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke session"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Revoke all sessions except the current one
// @Tags        Auth
// @Security    BearerAuth
// @Success     204 "Revoked"
// @Failure     401 {object} map[string]string
// @Router      /auth/sessions [delete]
func (h *AuthHandler) RevokeOtherSessions(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.authUsecase.RevokeOtherSessions(c.Request.Context(), uid, sessionIDFromCtx(c)); err != nil {
		logger.Error("revoke other sessions failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke sessions"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Upload avatar for the current user
// @Tags        Auth
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "Image file (png/jpg/webp/gif, ≤2MB)"
// @Success     200 {object} dto.UserResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     413 {object} map[string]string
// @Failure     415 {object} map[string]string
// @Router      /auth/me/avatar [post]
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, avatarMaxBytes+512)
	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 2MB)"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}
	if file.Size > avatarMaxBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 2MB)"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not open uploaded file"})
		return
	}
	defer src.Close()

	// Sniff content type from the first 512 bytes (MIME magic), don't trust the
	// header — clients can lie. http.DetectContentType is what Gin's Static uses
	// internally.
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	contentType := http.DetectContentType(head[:n])
	ext, allowed := avatarAllowedTypes[contentType]
	if !allowed {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "only png/jpeg/webp/gif images are allowed"})
		return
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read file"})
		return
	}

	// Random filename component so old paths can't be guessed and the browser
	// doesn't serve a stale cached image.
	rnd := make([]byte, 8)
	if _, err := rand.Read(rnd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate filename"})
		return
	}
	filename := fmt.Sprintf("u%d-%s%s", uid, hex.EncodeToString(rnd), ext)
	dir := filepath.Join(h.uploadDir, avatarSubdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.Error("avatar dir mkdir failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}
	dst := filepath.Join(dir, filename)
	out, err := os.Create(dst)
	if err != nil {
		logger.Error("avatar create failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		logger.Error("avatar copy failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}

	imageURL := fmt.Sprintf("/uploads/%s/%s?v=%d", avatarSubdir, filename, time.Now().Unix())
	resp, err := h.authUsecase.UpdateAvatar(c.Request.Context(), uid, imageURL)
	if err != nil {
		_ = os.Remove(dst)
		logger.Error("update avatar failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}

	// Best-effort cleanup of any older avatar files for this user. The new file
	// has a fresh random suffix, so we look for previous u{id}- prefixed files
	// and delete them.
	deleteOldAvatarsExcept(dir, fmt.Sprintf("u%d-", uid), filename)

	c.JSON(http.StatusOK, resp)
}

// @Summary     Remove avatar for the current user
// @Tags        Auth
// @Security    BearerAuth
// @Success     200 {object} dto.UserResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/me/avatar [delete]
func (h *AuthHandler) RemoveAvatar(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	resp, err := h.authUsecase.UpdateAvatar(c.Request.Context(), uid, "")
	if err != nil {
		logger.Error("remove avatar failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not remove avatar"})
		return
	}
	dir := filepath.Join(h.uploadDir, avatarSubdir)
	deleteOldAvatarsExcept(dir, fmt.Sprintf("u%d-", uid), "")
	c.JSON(http.StatusOK, resp)
}

func deleteOldAvatarsExcept(dir, prefix, keep string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || name == keep {
			continue
		}
		_ = os.Remove(filepath.Join(dir, name))
	}
}

// @Summary     Change own password
// @Tags        Auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.ChangePasswordRequest true "Old + new password"
// @Success     204 "Password changed"
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	id, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.authUsecase.ChangePassword(c.Request.Context(), id, input)
	audit.Record(c, audit_connectors.Event{Action: "auth.change_password"}, audit_domain.PASSWORD_CHANGE_ATTEMPT, audit_domain.PASSWORD_CHANGE_SUCCESS, err)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, domain.ErrCurrentPassword):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
	case errors.Is(err, domain.ErrSamePassword):
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must differ from current"})
	default:
		logger.Error("change password failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not change password"})
	}
}

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
	if err != nil {
		logger.Error("password reset init failed: " + err.Error())
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
	err := h.authUsecase.FinishPasswordReset(c.Request.Context(), input)
	audit.Record(c, audit_connectors.Event{Action: "auth.reset_password.finish"}, audit_domain.RESET_USER_PASSWORD_ATTEMPT, audit_domain.RESET_USER_PASSWORD_SUCCESS, err)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, domain.ErrInvalidResetKey):
		c.JSON(http.StatusGone, gin.H{"error": "invalid or expired reset key"})
	default:
		logger.Error("password reset finish failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not reset password"})
	}
}

func userIDFromCtx(c *gin.Context) (uint64, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := raw.(uint64)
	return id, ok
}

func sessionIDFromCtx(c *gin.Context) uint64 {
	if v, ok := c.Get("session_id"); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func userLoginFromCtx(c *gin.Context) string {
	if v, ok := c.Get("user_login"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
