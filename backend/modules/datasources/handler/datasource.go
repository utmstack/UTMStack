package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type DatasourceHandler struct {
	uc      connectors.DatasourceUsecase
	license connectors.LicenseCapProvider
}

func NewDatasourceHandler(uc connectors.DatasourceUsecase, license connectors.LicenseCapProvider) *DatasourceHandler {
	return &DatasourceHandler{uc: uc, license: license}
}

// Usage godoc
//
//	@Summary     Datasource usage vs license limit (for the usage banner)
//	@Tags        Datasources
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {object} dto.UsageResponse
//	@Failure     500 {object} map[string]string
//	@Router      /datasources/usage [get]
func (h *DatasourceHandler) Usage(c *gin.Context) {
	count, err := h.uc.Count(c.Request.Context())
	if err != nil {
		writeError(c, "count datasources", err)
		return
	}
	limit, unlimited := h.license.DatasourceCap()
	c.JSON(http.StatusOK, dto.UsageResponse{
		Count:     count,
		Limit:     limit,
		Unlimited: unlimited,
		OverLimit: !unlimited && count > limit,
	})
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

// Enrichment is the internal feed the alerts plugin caches to enrich alerts with
// their datasource's group/labels. Not audited: internal, high-frequency poll.
func (h *DatasourceHandler) Enrichment(c *gin.Context) {
	out, err := h.uc.Enrichment(c.Request.Context())
	if err != nil {
		writeError(c, "datasource enrichment", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *DatasourceHandler) UpdateGroup(c *gin.Context) {
	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	err := h.uc.UpdateGroup(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "datasource.group.assign", ResourceType: "datasource"},
		audit_domain.DATASOURCE_GROUP_ASSIGN_ATTEMPT, audit_domain.DATASOURCE_GROUP_ASSIGN_SUCCESS, err)
	if err != nil {
		writeError(c, "update group", err)
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
	audit.Record(c, audit_connectors.Event{Action: "datasource.labels.update", ResourceType: "datasource", ResourceID: strconv.FormatUint(req.ID, 10)},
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
	audit.Record(c, audit_connectors.Event{Action: "datasource.sensitivity.update", ResourceType: "datasource", ResourceID: strconv.FormatUint(req.ID, 10)},
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
	audit.Record(c, audit_connectors.Event{Action: "datasource.delete", ResourceType: "datasource", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.DATASOURCE_DELETE_ATTEMPT, audit_domain.DATASOURCE_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, "delete datasource", err)
		return
	}
	c.Status(http.StatusNoContent)
}
