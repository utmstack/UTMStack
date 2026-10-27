package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/threatintel/domain"
	"github.com/utmstack/utmstack/backend/modules/threatintel/repository"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type AnalystHandler struct {
	store *repository.AnalystStore
}

var errConversationNotFound = errors.New("conversation not found")

func NewAnalystHandler(store *repository.AnalystStore) *AnalystHandler {
	return &AnalystHandler{store: store}
}

// scope returns (tenantID, userID) for the request, or "" "" when the actor
// is missing. Both values come from the authenticated actor — the tenant from
// the request context, the user from the middleware's "user_id" entry — so a
// tenant's conversations are never listed against another tenant.
func (h *AnalystHandler) scope(c *gin.Context) (string, string) {
	tenantID := authz.TenantIDFromContext(c.Request.Context())
	uidRaw, _ := c.Get("user_id")
	switch uid := uidRaw.(type) {
	case string:
		return tenantID, uid
	case uuid.UUID:
		return tenantID, uid.String()
	}
	return tenantID, ""
}

type conversationDTO struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Turns     interface{} `json:"turns"`
	History   interface{} `json:"history"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

func (h *AnalystHandler) List(c *gin.Context) {
	tenantID, userID := h.scope(c)
	if tenantID == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.store.List(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}
	out := make([]conversationDTO, 0, len(rows))
	for _, row := range rows {
		var turns interface{}
		_ = json.Unmarshal([]byte(row.Turns), &turns)
		var history interface{}
		_ = json.Unmarshal([]byte(row.History), &history)
		out = append(out, conversationDTO{ID: row.ID, Title: row.Title, Turns: turns, History: history, UpdatedAt: row.UpdatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"conversations": out})
}

func (h *AnalystHandler) Create(c *gin.Context) {
	tenantID, userID := h.scope(c)
	if tenantID == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req conversationDTO
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and valid JSON body required"})
		return
	}
	turns, _ := json.Marshal(req.Turns)
	history, _ := json.Marshal(req.History)
	conv := &domain.AnalystConversation{
		ID:       req.ID,
		TenantID: tenantID,
		UserID:   userID,
		Title:    req.Title,
		Turns:    string(turns),
		History:  string(history),
	}
	if err := h.store.Create(c.Request.Context(), conv); err != nil {
		audit.Record(c, audit_connectors.Event{Action: "threatintel.analyst.create", ResourceType: "threatintel_analyst_conversation", ResourceID: conv.ID, Metadata: map[string]any{"title": req.Title}},
			audit_domain.THREATINTEL_ANALYST_CREATE_ATTEMPT, audit_domain.THREATINTEL_ANALYST_CREATE_SUCCESS, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": conv.ID})
}

func (h *AnalystHandler) Update(c *gin.Context) {
	tenantID, userID := h.scope(c)
	if tenantID == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	var req conversationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid JSON body required"})
		return
	}
	conv, err := h.store.Get(c.Request.Context(), tenantID, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load conversation"})
		return
	}
	if conv == nil {
		audit.Record(c, audit_connectors.Event{Action: "threatintel.analyst.update", ResourceType: "threatintel_analyst_conversation", ResourceID: id},
			audit_domain.THREATINTEL_ANALYST_UPDATE_ATTEMPT, audit_domain.THREATINTEL_ANALYST_UPDATE_SUCCESS,
			errConversationNotFound)
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	conv.Title = req.Title
	turns, _ := json.Marshal(req.Turns)
	history, _ := json.Marshal(req.History)
	conv.Turns = string(turns)
	conv.History = string(history)
	if err := h.store.Save(c.Request.Context(), conv); err != nil {
		audit.Record(c, audit_connectors.Event{Action: "threatintel.analyst.update", ResourceType: "threatintel_analyst_conversation", ResourceID: id, Metadata: map[string]any{"title": req.Title}},
			audit_domain.THREATINTEL_ANALYST_UPDATE_ATTEMPT, audit_domain.THREATINTEL_ANALYST_UPDATE_SUCCESS, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save conversation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AnalystHandler) Delete(c *gin.Context) {
	tenantID, userID := h.scope(c)
	if tenantID == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	if err := h.store.Delete(c.Request.Context(), tenantID, userID, id); err != nil {
		audit.Record(c, audit_connectors.Event{Action: "threatintel.analyst.delete", ResourceType: "threatintel_analyst_conversation", ResourceID: id},
			audit_domain.THREATINTEL_ANALYST_DELETE_ATTEMPT, audit_domain.THREATINTEL_ANALYST_DELETE_SUCCESS, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete conversation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
