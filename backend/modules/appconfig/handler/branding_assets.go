package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
	"github.com/utmstack/utmstack/backend/modules/appconfig/usecase"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
)

const (
	brandingSubdir    = "branding"
	brandingMaxBytes  = 5 << 20 // 5 MiB per image
	brandingURLPrefix = "/uploads/" + brandingSubdir
)

var brandingRasterTypes = map[string]string{
	"image/png":                ".png",
	"image/jpeg":               ".jpg",
	"image/gif":                ".gif",
	"image/webp":               ".webp",
	"image/x-icon":             ".ico",
	"image/vnd.microsoft.icon": ".ico",
}

var validBrandingSlots = map[string]bool{
	usecase.AssetLogo:        true,
	usecase.AssetLogoDark:    true,
	usecase.AssetFavicon:     true,
	usecase.AssetReportLogo:  true,
	usecase.AssetReportCover: true,
}

func brandingExt(head []byte, origName string) (string, bool) {
	if ext, ok := brandingRasterTypes[http.DetectContentType(head)]; ok {
		return ext, true
	}
	if strings.HasSuffix(strings.ToLower(origName), ".svg") && bytes.Contains(head, []byte("<svg")) {
		return ".svg", true
	}
	return "", false
}

func (h *BrandingHandler) storeBrandingFile(slot string, fh *multipart.FileHeader) (string, error) {
	if fh.Size > brandingMaxBytes {
		return "", fmt.Errorf("image %q is too large (max 5MB)", slot)
	}
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	ext, ok := brandingExt(head[:n], fh.Filename)
	if !ok {
		return "", fmt.Errorf("image %q must be png/jpeg/webp/gif/svg/ico", slot)
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	rnd := make([]byte, 8)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s-%s%s", slot, hex.EncodeToString(rnd), ext)
	dir := filepath.Join(h.uploadDir, brandingSubdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, filename)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, src); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return "", err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(dst)
		return "", err
	}

	// Best-effort: drop older files for this slot so they don't accumulate.
	deleteBrandingFilesExcept(dir, slot+"-", filename)
	return brandingURLPrefix + "/" + filename, nil
}

func deleteBrandingFilesExcept(dir, prefix, keep string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if name := e.Name(); strings.HasPrefix(name, prefix) && name != keep {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
}

// UploadAsset godoc
//
//	@Summary		Upload a white-label branding image
//	@Description	Stores a branding image (logo/logoDark/favicon/loginBackground/reportLogo/reportCover) and points the matching branding URL at it. MSSP feature.
//	@Tags			Branding
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			slot	path		string	true	"Asset slot: logo|logoDark|favicon|loginBackground|reportLogo|reportCover"
//	@Param			file	formData	file	true	"Image file (png/jpg/webp/gif/svg/ico, ≤5MB)"
//	@Success		200		{object}	dto.BrandingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		415		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/branding/assets/{slot} [post]
func (h *BrandingHandler) UploadAsset(c *gin.Context) {
	slot := c.Param("slot")
	if !validBrandingSlots[slot] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown asset slot"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, brandingMaxBytes+512)
	fh, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 5MB)"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}
	url, err := h.storeBrandingFile(slot, fh)
	if err != nil {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.SetAsset(c.Request.Context(), c.GetString("user_email"), slot, url)
	audit.Record(c, audit_connectors.Event{Action: "branding.asset.uploaded", ResourceType: "branding", ResourceID: slot},
		audit_domain.CONFIG_CHANGED, audit_domain.CONFIG_CHANGED, err)
	if errors.Is(err, usecase.ErrUnknownAssetSlot) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		_ = catcher.Error("branding asset upload failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save asset"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// brandingBool parses an optional form boolean, defaulting to def when absent.
func brandingBool(v string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return def
	}
}

// Seed godoc
//
//	@Summary		Seed white-label branding at install time
//	@Description	Internal-only. The installer posts the customer's brand fields and image files; the brand is persisted (enabled) and rendered once the MSSP license is present.
//	@Tags			Branding
//	@Accept			multipart/form-data
//	@Produce		json
//	@Success		200	{object}	dto.BrandingResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/branding/seed [post]
func (h *BrandingHandler) Seed(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(8 * brandingMaxBytes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart payload"})
		return
	}
	req := dto.BrandingRequest{
		Enabled:     brandingBool(c.PostForm("enabled"), true),
		ProductName: strings.TrimSpace(c.PostForm("productName")),
	}

	// Optional image files, one per slot. The setter points the matching URL at
	// the freshly stored file.
	slots := map[string]*string{
		usecase.AssetLogo:        &req.LogoURL,
		usecase.AssetLogoDark:    &req.LogoDarkURL,
		usecase.AssetFavicon:     &req.FaviconURL,
		usecase.AssetReportLogo:  &req.ReportLogoURL,
		usecase.AssetReportCover: &req.ReportCoverURL,
	}
	for slot, target := range slots {
		fh, err := c.FormFile(slot)
		if err != nil || fh == nil {
			continue
		}
		url, err := h.storeBrandingFile(slot, fh)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		*target = url
	}

	resp, err := h.usecase.Seed(c.Request.Context(), req)
	if err != nil {
		_ = catcher.Error("branding seed failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not seed branding"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
