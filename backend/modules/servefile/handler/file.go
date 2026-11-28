package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/servefile/connectors"
	"github.com/utmstack/utmstack/backend/modules/servefile/domain"
	"github.com/utmstack/utmstack/backend/modules/servefile/usecase"
)

const maxUploadBytes = 10 << 20 // 10 MiB

type FileHandler struct {
	uc connectors.ServedFileUsecase
}

func NewFileHandler(uc connectors.ServedFileUsecase) *FileHandler {
	return &FileHandler{uc: uc}
}

// Upload godoc
//
//	@Summary		Upload a file
//	@Description	Stores an encrypted file for the acting tenant. Sniffs the
//	@Tags			ServedFiles
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"File (pdf/png/jpeg/gif/webp, ≤10MB)"
//	@Success		201		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		413		{object}	map[string]string
//	@Failure		415		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/file [post]
func (h *FileHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+512)
	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file is too large (max 10MB)"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}
	if file.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file is too large (max 10MB)"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not open uploaded file"})
		return
	}
	defer src.Close()

	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	ct := usecase.DetectContentType(head[:n])
	if !usecase.ContentTypeAllowed(ct) {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "only pdf/png/jpeg/gif/webp files are allowed"})
		return
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read file"})
		return
	}
	data, err := io.ReadAll(src)
	if err != nil {
		_ = catcher.Error("servefile: read uploaded file failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read file"})
		return
	}

	name := filepath.Base(file.Filename)
	row, err := h.uc.Store(c.Request.Context(), name, ct, data)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"url": "/api/v1/file/" + usecase.FileName(row.ID, row.TenantID)})
}

// Download godoc
//
//	@Summary		Download a file
//	@Description	Retrieves and decrypts a stored file for the acting tenant.
//	@Tags			ServedFiles
//	@Security		BearerAuth
//	@Produce		binary
//	@Param			filename	path		string	true	"<fileID>-<tenant>.enc"
//	@Success		200
//	@Failure		403
//	@Failure		404
//	@Failure		500
//	@Router			/file/{filename} [get]
func (h *FileHandler) Download(c *gin.Context) {
	filename := c.Param("filename")
	if !strings.HasSuffix(filename, ".enc") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	fileID, tenant, ok := usecase.SplitFileName(filename)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	actorTenant, actorErr := uuid.Parse(c.GetString("tenant_id"))
	if actorErr != nil || actorTenant != tenant {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	data, row, err := h.uc.Fetch(c.Request.Context(), fileID, actorTenant.String())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		writeError(c, err)
		return
	}

	c.Header("Content-Disposition", contentDisposition(row.Filename))
	c.Data(http.StatusOK, row.ContentType, data)
}

func contentDisposition(name string) string {
	var ascii strings.Builder
	for _, r := range name {
		switch {
		case r == '"' || r == '\'' || r == '\r' || r == '\n':
		case r > unicode.MaxASCII:
			ascii.WriteRune('_')
		default:
			ascii.WriteRune(r)
		}
	}
	if ascii.Len() == 0 {
		ascii.WriteString("file")
	}
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, ascii.String(), url.PathEscape(name))
}
