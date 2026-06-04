package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

type clusterUsecase interface {
	ClusterStatus(ctx context.Context) (*dto.ClusterStatusResponse, error)
}

type ClusterHandler struct {
	uc clusterUsecase
}

func NewClusterHandler(uc clusterUsecase) *ClusterHandler {
	return &ClusterHandler{uc: uc}
}

// @Summary     Get cluster status
// @Description Returns OpenSearch cluster health information (name, status, nodes, shards).
// @Tags        OpenSearch
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.ClusterStatusResponse
// @Success     204 "No cluster information available"
// @Failure     500 {object} map[string]string
// @Router      /opensearch/cluster/status [get]
func (h *ClusterHandler) Status(c *gin.Context) {
	resp, err := h.uc.ClusterStatus(c.Request.Context())
	if err != nil {
		writeOSError(c, err)
		return
	}

	if resp == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, resp)
}
