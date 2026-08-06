package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
)

type queryRunner interface {
	Run(ctx context.Context, spec domain.Spec) (*usecase.Result, error)
}

type QueryHandler struct{ uc queryRunner }

func NewQueryHandler(uc queryRunner) *QueryHandler { return &QueryHandler{uc: uc} }

// Run godoc
//
//	@Summary		Answer a visualization
//	@Description	Takes the question a widget asks — dataset, aggregation, breakdown, filters — and answers it against the event store. The tenant comes from the session and cannot be named in the request.
//	@Tags			Dashboards
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			spec	body		domain.Spec	true	"What to ask"
//	@Success		200		{object}	usecase.Result
//	@Failure		400		{object}	map[string]string
//	@Router			/visualizations/query [post]
func (h *QueryHandler) Run(c *gin.Context) {
	var spec domain.Spec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.uc.Run(c.Request.Context(), spec)
	if err != nil {
		if isSpecError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_ = catcher.Error("dashboards: query failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "the query could not be run"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// isSpecError separates "you asked for something that makes no sense" from "the
// store could not answer": the first is the caller's to fix, the second is ours.
func isSpecError(err error) bool {
	for _, e := range []error{
		domain.ErrDatasetRequired, domain.ErrUnknownDataset, domain.ErrUnknownChart,
		domain.ErrDimensionRequired, domain.ErrFieldRequired, domain.ErrUnknownAgg,
		domain.ErrUnknownOp,
	} {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
