package handler

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

type TenantHandler struct{ uc connectors.TenantUsecase }

func NewTenantHandler(uc connectors.TenantUsecase) *TenantHandler {
	return &TenantHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Provision a tenant
//	@Description	Provisions a tenant. Limited to administrators of the default tenant, which is the platform plane.
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.CreateRequest	true	"Tenant to provision"
//	@Success		201		{object}	domain.Tenant
//	@Failure		400		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Router			/tenants [post]
func (h *TenantHandler) Create(c *gin.Context) {
	var req dto.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.uc.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{
		Action:       "tenant.create",
		ResourceType: "tenant",
		ResourceID:   tenantID(t),
		Metadata:     map[string]any{"name": req.Name, "domain": req.Domain},
	}, audit_domain.TENANT_CREATE_ATTEMPT, audit_domain.TENANT_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, t)
}

// Update godoc
//
//	@Summary		Update a tenant
//	@Description	Renames a tenant, corrects its domain, changes its status or sets what its plan allows.
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Tenant id"
//	@Param			input	body		dto.UpdateRequest	true	"Fields to change"
//	@Success		200		{object}	domain.Tenant
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/tenants/{id} [put]
func (h *TenantHandler) Update(c *gin.Context) {
	var req dto.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	t, err := h.uc.Update(c.Request.Context(), tid, req)
	audit.Record(c, audit_connectors.Event{
		Action:       "tenant.update",
		ResourceType: "tenant",
		ResourceID:   c.Param("id"),
		Metadata:     updateMetadata(req),
	}, audit_domain.TENANT_UPDATE_ATTEMPT, audit_domain.TENANT_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, t)
}

// GetByID godoc
//
//	@Summary		Get a tenant
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Tenant id"
//	@Success		200	{object}	domain.Tenant
//	@Failure		404	{object}	map[string]string
//	@Router			/tenants/{id} [get]
func (h *TenantHandler) GetByID(c *gin.Context) {
	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	t, err := h.uc.GetByID(c.Request.Context(), tid)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, t)
}

// List godoc
//
//	@Summary		List tenants
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Produce		json
//	@Param			name	query		string	false	"Filter by name"
//	@Param			domain	query		string	false	"Filter by domain"
//	@Param			status	query		string	false	"Filter by status"
//	@Success		200		{array}		domain.Tenant
//	@Router			/tenants [get]
func (h *TenantHandler) List(c *gin.Context) {
	var f dto.Filter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.uc.List(c.Request.Context(), f)
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []domain.Tenant{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

// Terminate godoc
//
//	@Summary		Terminate a tenant
//	@Description	Marks the tenant for deletion. The row and its data are kept: an operator has to be able to see what was terminated.
//	@Tags			Tenants
//	@Produce		json
//	@Param			id	path	string	true	"Tenant id"
//	@Success		204
//	@Failure		404	{object}	map[string]string
//	@Router			/tenants/{id} [delete]
func (h *TenantHandler) Terminate(c *gin.Context) {
	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	err := h.uc.Terminate(c.Request.Context(), tid)
	audit.Record(c, audit_connectors.Event{
		Action:       "tenant.terminate",
		ResourceType: "tenant",
		ResourceID:   c.Param("id"),
	}, audit_domain.TENANT_TERMINATE_ATTEMPT, audit_domain.TENANT_TERMINATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Reactivate godoc
//
//	@Summary		Reactivate a terminated tenant
//	@Description	Flips a TERMINATED tenant back to ACTIVE. Fails if the tenant is not terminated.
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Tenant id"
//	@Success		200	{object}	domain.Tenant
//	@Failure		400	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/tenants/{id}/reactivate [post]
func (h *TenantHandler) Reactivate(c *gin.Context) {
	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	t, err := h.uc.Reactivate(c.Request.Context(), tid)
	audit.Record(c, audit_connectors.Event{
		Action:       "tenant.reactivate",
		ResourceType: "tenant",
		ResourceID:   c.Param("id"),
	}, audit_domain.TENANT_REACTIVATE_ATTEMPT, audit_domain.TENANT_REACTIVATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// PermanentlyDelete godoc
//
//	@Summary		Permanently delete a terminated tenant
//	@Description	Hard-deletes the tenant row and all rows scoped by tenant_id across the schema. Only allowed when the tenant is TERMINATED.
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Tenant id"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/tenants/{id}/permanent [delete]
func (h *TenantHandler) PermanentlyDelete(c *gin.Context) {
	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	err := h.uc.PermanentlyDelete(c.Request.Context(), tid)
	audit.Record(c, audit_connectors.Event{
		Action:       "tenant.purge",
		ResourceType: "tenant",
		ResourceID:   c.Param("id"),
	}, audit_domain.TENANT_PURGE_ATTEMPT, audit_domain.TENANT_PURGE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrDomainTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrNotOwnTenant), errors.Is(err, domain.ErrDefaultTenant):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrDomainRequired),
		errors.Is(err, domain.ErrDomainInvalid), errors.Is(err, domain.ErrStatusInvalid),
		errors.Is(err, domain.ErrAlreadyTerminated), errors.Is(err, domain.ErrSupportInvalid),
		errors.Is(err, domain.ErrLimitNegative), errors.Is(err, domain.ErrLimitInvalid),
		errors.Is(err, domain.ErrNotTerminated):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetSupportAccess godoc
//
//	@Summary		Read the support access this tenant has granted
//	@Description	Answers for the caller's own tenant only. The platform reads this from the tenant list instead.
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Tenant id"
//	@Success		200	{object}	dto.SupportAccessResponse
//	@Failure		403	{object}	map[string]string
//	@Router			/tenants/{id}/support-access [get]
func (h *TenantHandler) GetSupportAccess(c *gin.Context) {
	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	t, err := h.uc.GetByID(c.Request.Context(), tid)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SupportAccessResponse{SupportAccess: t.SupportAccess})
}

// SetSupportAccess godoc
//
//	@Summary		Grant or revoke platform support access to this tenant
//	@Tags			Tenants
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Tenant id"
//	@Param			request	body		dto.SupportAccessRequest		true	"Access level"
//	@Success		200		{object}	domain.Tenant
//	@Failure		403		{object}	map[string]string
//	@Router			/tenants/{id}/support-access [put]
func (h *TenantHandler) SetSupportAccess(c *gin.Context) {
	var req dto.SupportAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tid, ok := pathTenantID(c)
	if !ok {
		return
	}
	t, err := h.uc.SetSupportAccess(c.Request.Context(), tid, req.SupportAccess)
	audit.Record(c, audit_connectors.Event{
		Action:       "tenant.support_access.set",
		ResourceType: "tenant",
		ResourceID:   c.Param("id"),
		Metadata:     map[string]any{"level": string(req.SupportAccess)},
	}, audit_domain.TENANT_SUPPORT_ACCESS_ATTEMPT, audit_domain.TENANT_SUPPORT_ACCESS_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, t)
}

func tenantID(t *domain.Tenant) string {
	if t == nil {
		return ""
	}
	return t.ID.String()
}

// pathTenantID refuses anything that is not a uuid before it reaches a query, so
// a malformed path is a bad request rather than a database error.
func pathTenantID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return uuid.Nil, false
	}
	return id, true
}

// updateMetadata records what the request asked to change, not the whole tenant:
// an audit entry is for answering what was touched.
func updateMetadata(req dto.UpdateRequest) map[string]any {
	m := map[string]any{}
	if req.Name != "" {
		m["name"] = req.Name
	}
	if req.Domain != "" {
		m["domain"] = req.Domain
	}
	if req.Status != "" {
		m["status"] = string(req.Status)
	}
	// Recorded as sent: "removed" and "never mentioned" have to read
	// differently in the audit trail.
	if limit, present, err := req.AILimit(); err == nil && present {
		if limit == nil {
			m["maxAIRequests"] = "removed"
		} else {
			m["maxAIRequests"] = *limit
		}
	}
	return m
}
