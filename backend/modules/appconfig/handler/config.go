package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type Handler struct {
	usecase connectors.Usecase
}

func NewHandler(uc connectors.Usecase) *Handler {
	return &Handler{usecase: uc}
}

// @Summary List config entries
// @Tags Config
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.ConfigResponse
// @Router /config [get]
func (h *Handler) List(c *gin.Context) {
	resp, err := h.usecase.List(c.Request.Context())
	if err != nil {
		logger.Error("config list failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list config"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Get a config entry by key
// @Tags Config
// @Security BearerAuth
// @Produce json
// @Param key path string true "Config key"
// @Success 200 {object} dto.ConfigResponse
// @Failure 404 {object} map[string]string
// @Router /config/{key} [get]
func (h *Handler) Get(c *gin.Context) {
	key := c.Param("key")
	resp, err := h.usecase.Get(c.Request.Context(), key)
	if err != nil {
		logger.Error("config get failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get config"})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Update a config entry
// @Description Updates the value (and metadata) of an existing, seeded parameter.
// @Description Parameters are not created through the API, so an unknown key
// @Description returns 404.
// @Tags Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "Config key"
// @Param input body dto.UpsertRequest true "Value + metadata"
// @Success 200 {object} dto.ConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /config/{key} [put]
func (h *Handler) Update(c *gin.Context) {
	key := c.Param("key")
	var input dto.UpsertRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), c.GetString("user_login"), key, input)
	if resp == nil && err == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config parameter not found"})
		return
	}
	audit.Record(c, audit_connectors.Event{Action: "config.updated", ResourceType: "config", ResourceID: key},
		audit_domain.CONFIG_CHANGED, audit_domain.CONFIG_CHANGED, err)
	if err != nil {
		logger.Error("config update failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save config"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Check mail configurations
// @Description Sends a verification email through each supplied SMTP
// @Description configuration. Each entry's `from` address is used as both
// @Description sender and recipient (self-ping) — receiving the message
// @Description confirms the server accepted the send.
// @Tags Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body []domain.MailConfig true "Configurations to test"
// @Success 200 {object} dto.CheckMailResponse
// @Failure 400 {object} dto.CheckMailResponse
// @Router /config/check-mail [post]
func (h *Handler) CheckMail(c *gin.Context) {
	var configs []domain.MailConfig
	if err := c.ShouldBindJSON(&configs); err != nil {
		c.JSON(http.StatusBadRequest, dto.CheckMailResponse{Success: false, Message: err.Error()})
		return
	}
	err := h.usecase.CheckMail(c.Request.Context(), configs)
	audit.Record(c, audit_connectors.Event{Action: "config.mail.checked", ResourceType: "config", ResourceID: "mail"},
		audit_domain.CONFIG_CHANGED, audit_domain.CONFIG_CHANGED, err)
	if err != nil {
		logger.Error("config check mail failed: " + err.Error())
		c.JSON(http.StatusBadRequest, dto.CheckMailResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.CheckMailResponse{Success: true, Message: "test email sent"})
}
