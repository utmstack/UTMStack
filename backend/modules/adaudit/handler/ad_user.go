package handler

import (
	"net/http"
	"strconv"

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
//	@Param			search		query		string	false	"Substring on samAccountName/sid"
//	@Param			tenantId	query		string	false	"Filter by tenant"
//	@Param			active		query		bool	false	"Filter by active"
//	@Param			page		query		int		false	"Page (0-based)"
//	@Param			size		query		int		false	"Page size"
//	@Success		200			{array}		domain.ADUser
//	@Header			200			{string}	X-Total-Count	"Total records"
//	@Failure		500			{object}	map[string]string
//	@Router			/ad-audit/users [get]
func (h *ADUserHandler) List(c *gin.Context) {
	var f dto.ADUserFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// Sync godoc
//
//	@Summary		Export all AD users (internal)
//	@Description	Internal endpoint the ad-audit plugin calls at startup to seed its in-memory cache.
//	@Tags			AD Audit
//	@Produce		json
//	@Success		200	{array}	domain.ADUser
//	@Failure		500	{object}	map[string]string
//	@Router			/ad-audit/users/sync [get]
func (h *ADUserHandler) Sync(c *gin.Context) {
	users, err := h.uc.All(c.Request.Context())
	if err != nil {
		_ = catcher.Error("adaudit: sync failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not export users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
