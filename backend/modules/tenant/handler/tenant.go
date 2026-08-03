package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
//	@Description	Creates a platform tenant. Internal-only: a tenant is created by whatever sells the subscription, never from inside an instance.
//	@Tags			Tenants
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
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, t)
}

// Update godoc
//
//	@Summary		Update a tenant
//	@Description	Renames a tenant, corrects its domain or changes its status.
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

	t, err := h.uc.Update(c.Request.Context(), c.Param("id"), req)
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
	t, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
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
	if err := h.uc.Terminate(c.Request.Context(), c.Param("id")); err != nil {
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
	case errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrDomainRequired),
		errors.Is(err, domain.ErrDomainInvalid), errors.Is(err, domain.ErrStatusInvalid),
		errors.Is(err, domain.ErrAlreadyTerminated):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
