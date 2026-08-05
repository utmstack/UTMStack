package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
)

type ADUserHandler struct {
	uc connectors.ADUserUsecase
}

func NewADUserHandler(uc connectors.ADUserUsecase) *ADUserHandler {
	return &ADUserHandler{uc: uc}
}

// Ingest godoc
//
//	@Summary		Ingest AD users (internal)
//	@Description	Internal endpoint the ad-audit plugin calls to upsert the users that changed since its last flush.
//	@Tags			AD Audit
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.IngestRequest	true	"Users to upsert"
//	@Success		200		{object}	map[string]int
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/ad-audit/users [post]
func (h *ADUserHandler) Ingest(c *gin.Context) {
	var req dto.IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.uc.Ingest(c.Request.Context(), req)
	if err != nil {
		_ = catcher.Error("adaudit: ingest failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not ingest users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accepted": n})
}

// List godoc
//
//	@Summary		List AD users
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Param			search		query		string	false	"Substring on samAccountName/sid/username"
//	@Param			source		query		string	false	"Filter by source: windows | linux (omit for all)"
//	@Param			active		query		bool	false	"Filter by active"
//	@Param			status		query		string	false	"Lifecycle bucket: active|disabled|deleted|stale|service (overrides active)"
//	@Param			sort		query		string	false	"Sort: recent (last seen) or name (default)"
//	@Param			page		query		int		false	"Page (0-based)"
//	@Param			size		query		int		false	"Page size"
//	@Success		200			{array}		domain.ADUser
//	@Header			200			{string}	X-Total-Count	"Total records"
//	@Failure		400			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/ad-audit/users [get]
func (h *ADUserHandler) List(c *gin.Context) {
	var f dto.ADUserFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if f.Source != "" && f.Source != "windows" && f.Source != "linux" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source must be 'windows', 'linux', or omitted"})
		return
	}
	res, err := h.uc.List(c.Request.Context(), f)
	if err != nil {
		_ = catcher.Error("adaudit: list failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list users"})
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(res.Total, 10))
	c.JSON(http.StatusOK, res.Items)
}

// Stats godoc
//
//	@Summary		AD user inventory stats
//	@Description	Roll-up for the User Auditor overview: lifecycle counts and by-domain breakdown, scoped to the caller's tenant.
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200			{object}	dto.ADUserStats
//	@Failure		500			{object}	map[string]string
//	@Router			/ad-audit/stats [get]
func (h *ADUserHandler) Stats(c *gin.Context) {
	res, err := h.uc.Stats(c.Request.Context())
	if err != nil {
		_ = catcher.Error("adaudit: stats failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load stats"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Sync godoc
//
//	@Summary		Export all AD users (internal)
//	@Description	Internal endpoint the ad-audit plugin calls at startup to seed its in-memory cache.
//	@Description	Accepts an optional `source` filter so the plugin can seed its Windows and Linux
//	@Tags			AD Audit
//	@Produce		json
//	@Param			source	query		string	false	"Filter by source: windows | linux (omit for all)"
//	@Success		200		{array}		domain.ADUser
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/ad-audit/users/sync [get]
func (h *ADUserHandler) Sync(c *gin.Context) {
	source := c.Query("source")
	if source != "" && source != "windows" && source != "linux" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source must be 'windows', 'linux', or omitted"})
		return
	}
	c.Header("Content-Type", "application/x-ndjson")
	c.Status(http.StatusOK)

	enc := json.NewEncoder(c.Writer)
	err := h.uc.Each(c.Request.Context(), source, func(u domain.ADUser) error {
		return enc.Encode(u)
	})
	if err != nil {
		_ = catcher.Error("adaudit: sync failed part-way", err, nil)
	}
}

// Resolve godoc
//
//	@Summary		Resolve provisional Linux user rows (internal)
//	@Description	Internal endpoint the ad-audit plugin calls once it learns the
//	@Description	machine-id for a host that already has provisional Linux user rows
//	@Description	(machine_id IS NULL).
//	@Tags			AD Audit
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.ResolveLinuxIdentityRequest	true	"Resolution payload"
//	@Success		200		{object}	map[string]int64
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/ad-audit/users/resolve [post]
func (h *ADUserHandler) Resolve(c *gin.Context) {
	var req dto.ResolveLinuxIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.uc.ResolveLinuxIdentity(c.Request.Context(), req)
	if err != nil {
		_ = catcher.Error("adaudit: resolve failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not resolve provisional rows"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"resolved": n})
}
