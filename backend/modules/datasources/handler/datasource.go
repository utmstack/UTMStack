package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type DatasourceHandler struct {
	uc connectors.DatasourceUsecase
}

func NewDatasourceHandler(uc connectors.DatasourceUsecase) *DatasourceHandler {
	return &DatasourceHandler{uc: uc}
}

// Count godoc
//
//	@Summary     Number of configured datasources
//	@Tags        Datasources
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {object} dto.CountResponse
//	@Failure     500 {object} map[string]string
//	@Router      /datasources/count [get]
func (h *DatasourceHandler) Count(c *gin.Context) {
	count, err := h.uc.Count(c.Request.Context())
	if err != nil {
		writeError(c, "count datasources", err)
		return
	}
	c.JSON(http.StatusOK, dto.CountResponse{Count: count})
}

func (h *DatasourceHandler) List(c *gin.Context) {
	var req common_models.ListRequest
	_ = c.ShouldBindQuery(&req)
	res, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		writeError(c, "list datasources", err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *DatasourceHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "get datasource", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Ping is the internal register-or-update batch endpoint (agent-manager / pullers).
// Not audited: internal, high-frequency.
func (h *DatasourceHandler) Ping(c *gin.Context) {
	var req dto.PingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.uc.Ping(c.Request.Context(), req); err != nil {
		writeError(c, "ping datasources", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DatasourceHandler) UpdateLabels(c *gin.Context) {
	var req dto.UpdateLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	err := h.uc.UpdateLabels(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "datasource.labels.update", ResourceType: "datasource", ResourceID: req.ID.String()},
		audit_domain.DATASOURCE_LABELS_UPDATE_ATTEMPT, audit_domain.DATASOURCE_LABELS_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, "update labels", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DatasourceHandler) UpdateSensitivity(c *gin.Context) {
	var req dto.UpdateSensitivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	err := h.uc.UpdateSensitivity(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "datasource.sensitivity.update", ResourceType: "datasource", ResourceID: req.ID.String()},
		audit_domain.DATASOURCE_SENSITIVITY_UPDATE_ATTEMPT, audit_domain.DATASOURCE_SENSITIVITY_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, "update sensitivity", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "datasource.delete", ResourceType: "datasource", ResourceID: id.String()},
		audit_domain.DATASOURCE_DELETE_ATTEMPT, audit_domain.DATASOURCE_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, "delete datasource", err)
		return
	}
	c.Status(http.StatusNoContent)
}
