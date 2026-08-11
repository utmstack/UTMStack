package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/storage/connectors"
	"github.com/utmstack/utmstack/backend/modules/storage/domain"
	"github.com/utmstack/utmstack/backend/modules/storage/dto"
)

type Handler struct{ uc connectors.Usecase }

func New(uc connectors.Usecase) *Handler { return &Handler{uc: uc} }

// Retentions godoc
//
//	@Summary		How long each dataset is kept
//	@Description	Per dataset: the day a record is deleted, and the day it moves to cold storage when the instance has any.
//	@Tags			Storage
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}	dto.RetentionResponse
//	@Router			/storage/retention [get]
func (h *Handler) Retentions(c *gin.Context) {
	out, err := h.uc.Retentions(c.Request.Context())
	if err != nil {
		writeError(c, err, "could not read the retention")
		return
	}
	c.JSON(http.StatusOK, dto.FromRetentions(out))
}

// SetRetention godoc
//
//	@Summary		Set how long a dataset is kept
//	@Description	keepDays is the whole lifetime; coldDays is when records move to object storage and must be shorter. A dataset that already moves records cannot go back to local-only without rebuilding its table.
//	@Tags			Storage
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.RetentionRequest	true	"Dataset and its days"
//	@Success		200		{object}	dto.RetentionResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Router			/storage/retention [put]
func (h *Handler) SetRetention(c *gin.Context) {
	var req dto.RetentionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	got, err := h.uc.SetRetention(c.Request.Context(), req.ToDomain())
	audit.Record(c,
		audit_connectors.Event{Action: "retention.update", ResourceType: "dataset", ResourceID: req.Dataset},
		audit_domain.RETENTION_UPDATE_ATTEMPT, audit_domain.RETENTION_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err, "could not set the retention")
		return
	}
	c.JSON(http.StatusOK, dto.FromRetention(got))
}

// Usage godoc
//
//	@Summary		What each dataset holds
//	@Tags			Storage
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}	dto.UsageResponse
//	@Router			/storage/usage [get]
func (h *Handler) Usage(c *gin.Context) {
	out, err := h.uc.Usage(c.Request.Context())
	if err != nil {
		writeError(c, err, "could not read the usage")
		return
	}
	c.JSON(http.StatusOK, dto.FromUsage(out))
}

// Health godoc
//
//	@Summary		The state of the disk underneath
//	@Tags			Storage
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.HealthResponse
//	@Router			/storage/health [get]
func (h *Handler) Health(c *gin.Context) {
	out, err := h.uc.Health(c.Request.Context())
	if err != nil {
		writeError(c, err, "could not read the store health")
		return
	}
	c.JSON(http.StatusOK, dto.FromHealth(out))
}

// Tiering godoc
//
//	@Summary		Whether this instance has cold storage
//	@Description	Configured means a bucket is written; ready means the store has picked it up. Only ready allows a dataset to move records there. The credentials are never returned.
//	@Tags			Storage
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.TieringResponse
//	@Router			/storage/tiering [get]
func (h *Handler) Tiering(c *gin.Context) {
	out, err := h.uc.Tiering(c.Request.Context())
	if err != nil {
		writeError(c, err, "could not read the cold storage configuration")
		return
	}
	c.JSON(http.StatusOK, dto.FromTiering(out))
}

// EnableTiering godoc
//
//	@Summary		Point cold storage at an object store
//	@Description	Writes the bucket into the event store's configuration and makes it read it again. The bucket cannot be changed once records live in it, because the parts already moved carry it; the credentials can.
//	@Tags			Storage
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.ObjectStoreRequest	true	"Bucket URL and credentials"
//	@Success		200		{object}	dto.TieringResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Failure		503		{object}	map[string]string
//	@Router			/storage/tiering [put]
func (h *Handler) EnableTiering(c *gin.Context) {
	var req dto.ObjectStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	got, err := h.uc.EnableTiering(c.Request.Context(), req.ToDomain())
	// The endpoint is recorded and the credentials are not: the audit says
	// where the records were sent, not how to read them.
	audit.Record(c,
		audit_connectors.Event{Action: "cold_storage.set", ResourceType: "object_store", ResourceID: req.Endpoint},
		audit_domain.COLD_STORAGE_SET_ATTEMPT, audit_domain.COLD_STORAGE_SET_SUCCESS, err)
	if err != nil {
		writeError(c, err, "could not enable cold storage")
		return
	}
	c.JSON(http.StatusOK, dto.FromTiering(got))
}

// A refused request is the caller's to fix and says why; anything else is ours
// and says nothing beyond that it failed.
func writeError(c *gin.Context, err error, msg string) {
	switch {
	case isBadRequest(err):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case isConflict(err):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrColdNotReady):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		_ = catcher.Error("storage: "+msg, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
	}
}

func isBadRequest(err error) bool {
	for _, e := range []error{
		domain.ErrUnknownDataset, domain.ErrKeepRequired, domain.ErrColdBeforeDelete,
		domain.ErrColdNegative, domain.ErrEndpointRequired, domain.ErrEndpointNotURL,
		domain.ErrEndpointNoBucket, domain.ErrCredentialsNeeded, domain.ErrCacheNegative,
		domain.ErrColdRefused,
	} {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

func isConflict(err error) bool {
	for _, e := range []error{
		domain.ErrTieringRequired, domain.ErrTieringPermanent, domain.ErrEndpointLocked,
	} {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
