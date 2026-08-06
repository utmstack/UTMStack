package handler

import (
	"errors"
	"github.com/google/uuid"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type APIKeyHandler struct {
	usecase connectors.APIKeyUsecase
}

func NewAPIKeyHandler(uc connectors.APIKeyUsecase) *APIKeyHandler {
	return &APIKeyHandler{usecase: uc}
}

// Authenticate is the internal endpoint the inputs gateway calls to validate an
// API key presented by a direct log pusher. 200 when the key is valid (exists,
// not expired, IP allowed, user active); 401 otherwise. Internal-only.
func (h *APIKeyHandler) Authenticate(c *gin.Context) {
	var req dto.APIKeyAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	res, err := h.usecase.Authenticate(c.Request.Context(), req.APIKey, req.ClientIP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"userId": res.UserID, "login": res.Email, "tenantId": res.TenantID})
}

func apiKeyID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	return id, err == nil
}

func (h *APIKeyHandler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAPIKeyNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "api key name already in use"})
	case errors.Is(err, domain.ErrAPIKeyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "api key not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

// @Summary     Create an API key
// @Tags        API Keys
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.APIKeyUpsertRequest true "API key name, allowed IPs and optional expiry"
// @Success     201 {object} dto.APIKeyResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /api-keys [post]
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req dto.APIKeyUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), actorID(c), req)
	audit.Record(c, audit_connectors.Event{Action: "apikey.create"}, audit_domain.API_KEY_CREATE_ATTEMPT, audit_domain.API_KEY_CREATE_SUCCESS, err)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// @Summary     List API keys
// @Tags        API Keys
// @Security    BearerAuth
// @Produce     json
// @Param       page      query int false "default 1"
// @Param       page_size query int false "default 20"
// @Success     200 {object} dto.APIKeyListResponse
// @Failure     401 {object} map[string]string
// @Router      /api-keys [get]
func (h *APIKeyHandler) List(c *gin.Context) {
	var q dto.ListAPIKeysQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.List(c.Request.Context(), actorID(c), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list api keys"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Get an API key by id
// @Tags        API Keys
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "API key id"
// @Success     200 {object} dto.APIKeyResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api-keys/{id} [get]
func (h *APIKeyHandler) Get(c *gin.Context) {
	id, ok := apiKeyID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.usecase.Get(c.Request.Context(), actorID(c), id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update an API key
// @Tags        API Keys
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path int true "API key id"
// @Param       input body dto.APIKeyUpsertRequest true "Updated name, allowed IPs and optional expiry"
// @Success     200 {object} dto.APIKeyResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /api-keys/{id} [put]
func (h *APIKeyHandler) Update(c *gin.Context) {
	id, ok := apiKeyID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.APIKeyUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), actorID(c), id, req)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete an API key
// @Tags        API Keys
// @Security    BearerAuth
// @Param       id path int true "API key id"
// @Success     204 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api-keys/{id} [delete]
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, ok := apiKeyID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err := h.usecase.Delete(c.Request.Context(), actorID(c), id)
	audit.Record(c, audit_connectors.Event{Action: "apikey.delete"}, audit_domain.API_KEY_DELETE_ATTEMPT, audit_domain.API_KEY_DELETE_SUCCESS, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Rotate an API key (generate a new secret)
// @Description Issues a new key value for the given id; the previous value stops working immediately.
// @Tags        API Keys
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "API key id"
// @Success     200 {object} map[string]string "object with the new api_key value"
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api-keys/{id}/generate [post]
func (h *APIKeyHandler) Generate(c *gin.Context) {
	id, ok := apiKeyID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	key, err := h.usecase.Generate(c.Request.Context(), actorID(c), id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_key": key})
}
