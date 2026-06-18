package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/billing/domain"
)

type licenseProvider interface {
	Current() domain.License
	Replace(envelope []byte) (domain.License, error)
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

// Upload godoc
//
//	@Summary     Upload/replace the signed LICENSE envelope (admin only)
//	@Description Validates the uploaded LICENSE against this instance and, only if
//	@Description valid (signature + not expired), atomically replaces the stored
//	@Description license and returns the newly evaluated edition.
//	@Tags        Billing
//	@Security    BearerAuth
//	@Accept      multipart/form-data
//	@Produce     json
//	@Param       file formData file true "LICENSE envelope file"
//	@Success     200 {object} domain.License
//	@Failure     400 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /billing/license [post]
func (h *LicenseHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "license file is required (multipart field 'file')"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer func() { _ = f.Close() }()

	envelope, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read uploaded file"})
		return
	}

	lic, err := h.license.Replace(envelope)
	audit.Record(c, audit_connectors.Event{Action: "license.upload", ResourceType: "license"},
		audit_domain.LICENSE_UPLOAD_ATTEMPT, audit_domain.LICENSE_UPLOAD_SUCCESS, err)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLicense) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_ = catcher.Error("billing: license upload failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to install license"})
		return
	}
	c.JSON(http.StatusOK, lic)
}
