package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
)

type ConfigurationHandler struct {
	usecase connectors.ConfigUsecase
}

func NewConfigurationHandler(uc connectors.ConfigUsecase) *ConfigurationHandler {
	return &ConfigurationHandler{usecase: uc}
}

// @Summary Update module-group configuration keys
// @Tags ModulesConfig
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body dto.UpdateGroupConfigurationRequest true "Configuration keys"
// @Success 200
// @Router /module-group-configurations/update [put]
func (h *ConfigurationHandler) Update(c *gin.Context) {
	var req dto.UpdateGroupConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.Update(c.Request.Context(), req); err != nil {
		writeError(c, "configuration.update", err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ConfigurationHandler) ListByGroupID(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Query("groupId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid groupId"})
		return
	}
	resp, err := h.usecase.ListByGroupID(c.Request.Context(), groupID)
	if err != nil {
		writeError(c, "configuration.list", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigurationHandler) GetByGroupAndKey(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Query("groupId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid groupId"})
		return
	}
	confKey := c.Query("confKey")
	if confKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing confKey"})
		return
	}
	resp, err := h.usecase.GetByGroupAndKey(c.Request.Context(), groupID, confKey)
	if err != nil {
		writeError(c, "configuration.get", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
