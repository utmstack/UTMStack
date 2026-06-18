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

type AssetGroupHandler struct {
	uc connectors.AssetGroupUsecase
}

func NewAssetGroupHandler(uc connectors.AssetGroupUsecase) *AssetGroupHandler {
	return &AssetGroupHandler{uc: uc}
}

func (h *AssetGroupHandler) List(c *gin.Context) {
	var req common_models.ListRequest
	_ = c.ShouldBindQuery(&req)
	res, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		writeError(c, "list groups", err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AssetGroupHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	g, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "get group", err)
		return
	}
	c.JSON(http.StatusOK, dto.ToAssetGroupDTO(g, 0))
}

func (h *AssetGroupHandler) Create(c *gin.Context) {
	var in dto.AssetGroupDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	out, err := h.uc.Create(c.Request.Context(), dto.FromAssetGroupDTO(in))
	ev := audit_connectors.Event{Action: "datasource.group.create", ResourceType: "datasource_group"}
	if out != nil {
		ev.ResourceID = strconv.FormatUint(out.ID, 10)
	}
	audit.Record(c, ev, audit_domain.DATASOURCE_GROUP_CREATE_ATTEMPT, audit_domain.DATASOURCE_GROUP_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, "create group", err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToAssetGroupDTO(out, 0))
}

func (h *AssetGroupHandler) Update(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var in dto.AssetGroupDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	in.ID = id
	out, err := h.uc.Update(c.Request.Context(), dto.FromAssetGroupDTO(in))
	audit.Record(c, audit_connectors.Event{Action: "datasource.group.update", ResourceType: "datasource_group", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.DATASOURCE_GROUP_UPDATE_ATTEMPT, audit_domain.DATASOURCE_GROUP_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, "update group", err)
		return
	}
	c.JSON(http.StatusOK, dto.ToAssetGroupDTO(out, 0))
}

func (h *AssetGroupHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "datasource.group.delete", ResourceType: "datasource_group", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.DATASOURCE_GROUP_DELETE_ATTEMPT, audit_domain.DATASOURCE_GROUP_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, "delete group", err)
		return
	}
	c.Status(http.StatusNoContent)
}
