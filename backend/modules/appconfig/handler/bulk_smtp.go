package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/constants"
)

// BulkSMTPHandler handles platform-admin bulk SMTP operations.
type BulkSMTPHandler struct {
	uc           connectors.Usecase
	store        connectors.Store
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkSMTPHandler(uc connectors.Usecase, store connectors.Store, lister func(context.Context) ([]string, error)) *BulkSMTPHandler {
	return &BulkSMTPHandler{uc: uc, store: store, tenantLister: lister}
}

// Update godoc
//
//	@Summary		Bulk update SMTP config across tenants
//	@Tags			Platform Config
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.BulkSMTPUpdateRequest	true	"Fields + selector"
//	@Success		200		{object}	common_models.BulkResult
//	@Router			/platform/config/smtp/bulk/update [post]
func (h *BulkSMTPHandler) Update(c *gin.Context) {
	var req dto.BulkSMTPUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	actor := c.GetString("user_email")
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		var tenErr error
		for _, kv := range req.Fields {
			if _, err := h.uc.Update(ctx, actor, kv.Key, dto.UpsertRequest{Value: kv.Value}); err != nil {
				tenErr = err
				break
			}
		}
		result.Append(tid, tenErr)
	}
	c.JSON(http.StatusOK, result)
}

// Test godoc
//
//	@Summary		Bulk test SMTP config across tenants
//	@Tags			Platform Config
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.BulkSMTPTestRequest	true	"Selector"
//	@Success		200		{object}	common_models.BulkResult
//	@Router			/platform/config/smtp/bulk/test [post]
func (h *BulkSMTPHandler) Test(c *gin.Context) {
	var req dto.BulkSMTPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		cfg := h.loadMailConfig(ctx)
		result.Append(tid, h.uc.CheckMail(ctx, []domain.MailConfig{cfg}))
	}
	c.JSON(http.StatusOK, result)
}

func (h *BulkSMTPHandler) loadMailConfig(ctx context.Context) domain.MailConfig {
	get := func(key string) string {
		v, _, _ := h.store.GetString(ctx, key)
		return v
	}
	cfg := domain.MailConfig{
		Host:     get(constants.PROP_MAIL_HOST),
		Username: get(constants.PROP_MAIL_USERNAME),
		Password: get(constants.PROP_MAIL_PASSWORD),
		From:     get(constants.PROP_MAIL_FROM),
		AuthType: get(constants.PROP_MAIL_SMTP_AUTH),
	}
	if p, err := strconv.Atoi(get(constants.PROP_MAIL_PORT)); err == nil {
		cfg.Port = p
	}
	return cfg
}
