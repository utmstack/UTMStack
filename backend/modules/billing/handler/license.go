package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/billing/domain"
)

type licenseProvider interface {
	Current() domain.License
}

type LicenseHandler struct {
	license licenseProvider
}

func NewLicenseHandler(license licenseProvider) *LicenseHandler {
	return &LicenseHandler{license: license}
}

// Get godoc
//
//	@Summary     Get the evaluated license / edition
//	@Description Authoritative edition (community|enterprise) and capabilities,
//	@Description derived from the signed LICENSE envelope.
//	@Tags        Billing
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {object} domain.License
//	@Router      /billing/license [get]
func (h *LicenseHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, h.license.Current())
}
