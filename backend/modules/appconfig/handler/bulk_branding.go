package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// BulkBrandingHandler applies branding changes across multiple tenants in one call.
type BulkBrandingHandler struct {
	brand        connectors.BrandingUsecase
	uploadDir    string
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkBrandingHandler(brand connectors.BrandingUsecase, uploadDir string, tenantLister func(context.Context) ([]string, error)) *BulkBrandingHandler {
	return &BulkBrandingHandler{brand: brand, uploadDir: uploadDir, tenantLister: tenantLister}
}

// resolveTenants returns the target tenant IDs, enumerating active ones when AllTenants is set.
// ponytail: DefaultTenantID filtered out of AllTenants — bulk ops must not
// overwrite the platform-plane config; callers can still target it explicitly
// via TenantIDs.
func resolveTenants(ctx context.Context, sel common_models.BulkTenantSelector, lister func(context.Context) ([]string, error)) ([]string, error) {
	if !sel.AllTenants {
		return sel.TenantIDs, nil
	}
	all, err := lister(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(all))
	for _, id := range all {
		if id != authz.DefaultTenantID {
			out = append(out, id)
		}
	}
	return out, nil
}

// Update godoc
//
//	@Summary		Bulk update branding across tenants
//	@Description	Applies the same branding configuration to multiple tenants. Partial failures are recorded; succeeded tenants are listed separately.
//	@Tags			Branding
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.BulkBrandingUpdateRequest	true	"Selector + branding overrides"
//	@Success		200		{object}	common_models.BulkResult
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/platform/branding/bulk/update [post]
func (h *BulkBrandingHandler) Update(c *gin.Context) {
	var req dto.BulkBrandingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	actorEmail := c.GetString("user_email")
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		// ponytail: skip default (platform-plane) tenant — bulk calls must not silently overwrite operator branding
		if tid == authz.DefaultTenantID {
			continue
		}
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.brand.Update(ctx, actorEmail, req.Branding)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// UploadAsset godoc
//
//	@Summary		Bulk upload a branding asset across tenants
//	@Description	Saves the file once then points the given slot URL at it for every selected tenant. Partial failures are recorded.
//	@Tags			Branding
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			slot		path		string	true	"Asset slot: logo|logoDark|favicon|reportLogo|reportCover"
//	@Param			file		formData	file	true	"Image file (png/jpg/webp/gif/svg/ico, ≤5MB)"
//	@Param			selector	formData	string	true	"JSON-encoded BulkTenantSelector"
//	@Success		200			{object}	common_models.BulkResult
//	@Failure		400			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/platform/branding/bulk/upload-asset/{slot} [post]
func (h *BulkBrandingHandler) UploadAsset(c *gin.Context) {
	slot := c.Param("slot")
	if !validBrandingSlots[slot] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown asset slot"})
		return
	}

	// Parse selector from multipart form field "selector".
	var sel common_models.BulkTenantSelector
	if raw := strings.TrimSpace(c.PostForm("selector")); raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing selector field"})
		return
	} else if err := json.Unmarshal([]byte(raw), &sel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid selector JSON: " + err.Error()})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, brandingMaxBytes+512)
	fh, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 5MB)"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}

	// Reuse existing BrandingHandler to store the file once.
	bh := &BrandingHandler{uploadDir: h.uploadDir}
	url, err := bh.storeBrandingFile(slot, fh)
	if err != nil {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
		return
	}

	tenantIDs, err := resolveTenants(c.Request.Context(), sel, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	actorEmail := c.GetString("user_email")
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		// ponytail: skip default (platform-plane) tenant — bulk calls must not silently overwrite operator branding
		if tid == authz.DefaultTenantID {
			continue
		}
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.brand.SetAsset(ctx, actorEmail, slot, url)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}
