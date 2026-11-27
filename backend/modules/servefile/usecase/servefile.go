package usecase

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/servefile/connectors"
	"github.com/utmstack/utmstack/backend/modules/servefile/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type fileUsecase struct {
	repo       connectors.ServedFileRepository
	storageDir string
	cipher     *fileCipher
}

func NewServedFileUsecase(repo connectors.ServedFileRepository, storageDir, encryptionKey string) connectors.ServedFileUsecase {
	c, _ := newFileCipher(encryptionKey)
	return &fileUsecase{repo: repo, storageDir: storageDir, cipher: c}
}

func hex(u uuid.UUID) string { return strings.ReplaceAll(u.String(), "-", "") }

func (u *fileUsecase) Store(ctx context.Context, filename, contentType string, data []byte) (*domain.ServedFile, error) {
	tenant, err := uuid.Parse(authz.TenantIDFromContext(ctx))
	if err != nil {
		return nil, domain.ErrTenantMismatch
	}

	fileID := uuid.New()
	enc, err := u.cipher.Encrypt(data)
	if err != nil {
		return nil, err
	}

	path := blobPath(u.storageDir, tenant, fileID)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		_ = catcher.Error("servefile: mkdir tenant dir failed", err, nil)
		return nil, err
	}
	if err := os.WriteFile(path, enc, 0o600); err != nil {
		_ = catcher.Error("servefile: write encrypted blob failed", err, nil)
		return nil, err
	}

	row := &domain.ServedFile{
		ID:          fileID,
		Filename:    filename,
		ContentType: contentType,
		Size:        int64(len(data)),
	}
	if err := u.repo.Save(ctx, row); err != nil {
		// Remove the now-orphaned blob so a failed insert leaves no unregistered
		// file on disk.
		if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			_ = catcher.Error("servefile: remove orphaned blob failed", rmErr, nil)
		}
		return nil, err
	}
	row.TenantID = tenant
	return row, nil
}

func (u *fileUsecase) Fetch(ctx context.Context, fileID uuid.UUID, tenantID string) ([]byte, *domain.ServedFile, error) {
	tenant, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}

	row, err := u.repo.FindByID(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}
	if row == nil {
		return nil, nil, domain.ErrNotFound
	}
	if row.TenantID != tenant {
		return nil, nil, domain.ErrTenantMismatch
	}

	enc, err := os.ReadFile(blobPath(u.storageDir, tenant, fileID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, domain.ErrNotFound
		}
		_ = catcher.Error("servefile: read encrypted blob failed", err, nil)
		return nil, nil, err
	}
	plain, err := u.cipher.Decrypt(enc)
	if err != nil {
		_ = catcher.Error("servefile: decrypt blob failed", err, nil)
		return nil, nil, err
	}
	return plain, row, nil
}

func blobPath(storageDir string, tenant, fileID uuid.UUID) string {
	return filepath.Join(storageDir, hex(tenant), FileName(fileID, tenant))
}

func FileName(fileID, tenant uuid.UUID) string {
	return hex(fileID) + "-" + hex(tenant) + ".enc"
}

func SplitFileName(name string) (fileID, tenant uuid.UUID, ok bool) {
	const encSuffix = ".enc"
	if !strings.HasSuffix(name, encSuffix) {
		return uuid.Nil, uuid.Nil, false
	}
	inner := name[:len(name)-len(encSuffix)]
	i := strings.LastIndex(inner, "-")
	if i <= 0 || i == len(inner)-1 {
		return uuid.Nil, uuid.Nil, false
	}
	f, err1 := uuid.Parse(inner[:i])
	t, err2 := uuid.Parse(inner[i+1:])
	if err1 != nil || err2 != nil {
		return uuid.Nil, uuid.Nil, false
	}
	return f, t, true
}

var allowedContentTypes = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"image/gif":       true,
	"image/webp":      true,
}

func ContentTypeAllowed(ct string) bool { return allowedContentTypes[ct] }

func DetectContentType(head []byte) string {
	if len(head) >= 12 &&
		string(head[0:4]) == "RIFF" &&
		string(head[8:12]) == "WEBP" {
		return "image/webp"
	}
	return http.DetectContentType(head)
}
