package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
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
	// The address that was tried, whether or not it exists: a failed attempt on a
	// name nobody has is exactly what an audit trail is read for.
	audit.Record(c, audit_connectors.Event{Action: "auth.login", UserEmail: input.Login},
		audit_domain.AUTH_ATTEMPT, audit_domain.AUTH_SUCCESS, err)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordLoginUnavailable):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrTfaMailUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrChallengeCooldown):
			// A rate limit is the caller's answer, not a server fault.
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrLoginBlocked):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many failed attempts; try again later"})
		case errors.Is(err, domain.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		case errors.Is(err, domain.ErrUserInactive):
			c.JSON(http.StatusForbidden, gin.H{"error": "user is deactivated"})
		case errors.Is(err, domain.ErrTfaNoEmail):
			c.JSON(http.StatusForbidden, gin.H{"error": "tfa is required but no email is set for this user"})
		default:
			_ = catcher.Error("login failed", err, nil)
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
		_ = catcher.Error("refresh failed", err, nil)
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
		_ = catcher.Error("logout failed", err, nil)
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
// @Produce     plain
// @Success     200 {string} string "Login of the current user"
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
	c.String(http.StatusOK, resp.Email)
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
			_ = catcher.Error("update me failed", err, nil)
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
		_ = catcher.Error("list sessions failed", err, nil)
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
	sid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	if err := h.authUsecase.RevokeSession(c.Request.Context(), uid, sid); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		_ = catcher.Error("revoke session failed", err, nil)
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
		if errors.Is(err, domain.ErrNoActiveSession) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no active session to preserve"})
			return
		}
		_ = catcher.Error("revoke other sessions failed", err, nil)
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
	filename := fmt.Sprintf("u%s-%s%s", uid, hex.EncodeToString(rnd), ext)
	dir := filepath.Join(h.uploadDir, avatarSubdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		_ = catcher.Error("avatar dir mkdir failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}
	dst := filepath.Join(dir, filename)
	out, err := os.Create(dst)
	if err != nil {
		_ = catcher.Error("avatar create failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		_ = catcher.Error("avatar copy failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}

	imageURL := fmt.Sprintf("/uploads/%s/%s?v=%d", avatarSubdir, filename, time.Now().Unix())
	resp, err := h.authUsecase.UpdateAvatar(c.Request.Context(), uid, imageURL)
	if err != nil {
		_ = os.Remove(dst)
		_ = catcher.Error("update avatar failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar"})
		return
	}

	// Best-effort cleanup of any older avatar files for this user. The new file
	// has a fresh random suffix, so we look for previous u{id}- prefixed files
	// and delete them.
	deleteOldAvatarsExcept(dir, fmt.Sprintf("u%s-", uid), filename)

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
		_ = catcher.Error("remove avatar failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not remove avatar"})
		return
	}
	dir := filepath.Join(h.uploadDir, avatarSubdir)
	deleteOldAvatarsExcept(dir, fmt.Sprintf("u%s-", uid), "")
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
		_ = catcher.Error("change password failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not change password"})
	}
}

func userIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	id := actorID(c)
	return id, id != uuid.Nil
}

func sessionIDFromCtx(c *gin.Context) uuid.UUID {
	if a := middleware.ActorFromGin(c); a != nil {
		return a.SessionID
	}
	return uuid.Nil
}

func userEmailFromCtx(c *gin.Context) string {
	if v, ok := c.Get("user_email"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// actorID reads the caller's id from the request, which the auth middleware
// already parsed. A request that reached a handler behind userAuth always has
// one; anything else is a wiring mistake and lands as uuid.Nil, which matches no
// row rather than matching the first.
func actorID(c *gin.Context) uuid.UUID {
	if a := middleware.ActorFromGin(c); a != nil {
		return a.UserID
	}
	return uuid.Nil
}
