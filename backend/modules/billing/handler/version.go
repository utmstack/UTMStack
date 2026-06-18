package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/billing/domain"
	"github.com/utmstack/utmstack/backend/modules/billing/dto"
)

type versionProvider interface {
	Info() (dto.VersionInfo, error)
}

type editionProvider interface {
	Current() domain.License
}

type VersionHandler struct {
	usecase versionProvider
	license editionProvider
}

func NewVersionHandler(uc versionProvider, license editionProvider) *VersionHandler {
	return &VersionHandler{usecase: uc, license: license}
}

// Get godoc
//
//	@Summary     Get deployment version and instance info
//	@Description Reads version.json and instance-config.yml from the updates folder.
//	@Tags        Billing
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {object} dto.VersionInfo
//	@Failure     500 {object} map[string]string
//	@Router      /billing/version [get]
func (h *VersionHandler) Get(c *gin.Context) {
	info, err := h.usecase.Info()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Edition is authoritative from the license, not version.json.
	info.Edition = string(h.license.Current().Edition)
	c.JSON(http.StatusOK, info)
}
