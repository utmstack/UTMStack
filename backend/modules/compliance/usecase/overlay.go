package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"syscall"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	SystemSubdir   = "system"
	UserSubdir     = "user"
	fileExt        = ".yaml"
	disabledSuffix = ".disabled"
)

type scannedFile struct {
	relPath string
	data    []byte
	enabled bool
	system  bool
}

func safeID(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	return !strings.ContainsAny(id, "/\\") && !strings.Contains(id, "..")
}

func tenantDir(ctx context.Context) (string, error) {
	tid := authz.TenantIDFromContext(ctx)
	if tid == "" {

		if tenancy.Enabled() {
			return "", ErrNoTenant
		}
		tid = authz.DefaultTenantID
	}
	if _, err := uuid.Parse(tid); err != nil {
		return "", domain.ErrInvalidID
	}
	return tid, nil
}

var ErrNoTenant = errors.New("compliance: no tenant in scope")

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func withTenantLock(dir string, fn func() error) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	lf, err := os.OpenFile(filepath.Join(dir, ".compliance.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lf.Close()

	if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	return fn()
}

func scanYAML(dir string, system bool) ([]scannedFile, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	var out []scannedFile
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		base, enabled := path, true
		if strings.HasSuffix(base, disabledSuffix) {
			base, enabled = strings.TrimSuffix(base, disabledSuffix), false
		}
		if filepath.Ext(base) != fileExt {
			return nil
		}
		rel, err := filepath.Rel(dir, base)
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		out = append(out, scannedFile{relPath: filepath.ToSlash(rel), data: data, enabled: enabled, system: system})
		return nil
	})
	return out, err
}

func tenantDirs(userRoot string) []string {
	entries, err := os.ReadDir(userRoot)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}
