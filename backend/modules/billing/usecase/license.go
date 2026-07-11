package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	lm "github.com/utmstack/license-manager-sdk"
	"github.com/utmstack/utmstack/backend/modules/billing/domain"
	"gopkg.in/yaml.v3"
)

// errVerifyNotConfigured means the build has no signing key/salt injected (dev
// builds; prod injects them via ldflags). It is not a license problem, so we fall
// back to Community without log spam — the refresh ticker would otherwise log an
// ERROR every interval.
var errVerifyNotConfigured = errors.New("license verification not configured")

const (
	licenseFileName = "LICENSE"
	refreshEvery    = 5 * time.Minute
)

type licenseInner struct {
	InstanceID  string    `json:"instance_id"`
	MSSP        bool      `json:"mssp"`
	Type        string    `json:"type"`
	Datasources int64     `json:"datasources"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type LicenseService struct {
	licenseFile  string
	instanceFile string
	publicKey    string
	salt         string
	communityCap int64 // datasource cap for community (no enterprise license)

	mu      sync.RWMutex
	current domain.License
}

func NewLicenseService(updatesDir, publicKey, salt string, communityCap int64) *LicenseService {
	return &LicenseService{
		licenseFile:  filepath.Join(updatesDir, licenseFileName),
		instanceFile: filepath.Join(updatesDir, instanceFileName),
		publicKey:    publicKey,
		salt:         salt,
		communityCap: communityCap,
		current:      domain.Community(),
	}
}

func (s *LicenseService) DatasourceCap() (limit int64, unlimited bool) {
	lic := s.Current()
	if !lic.IsEnterprise() {
		return s.communityCap, false
	}
	if lic.Datasources == 0 {
		return 0, true
	}
	return lic.Datasources, false
}

func (s *LicenseService) Current() domain.License {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

func (s *LicenseService) Start(ctx context.Context) {
	s.Refresh()
	go func() {
		t := time.NewTicker(refreshEvery)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.Refresh()
			}
		}
	}()
}

func (s *LicenseService) Refresh() domain.License {
	lic := s.evaluate()
	s.mu.Lock()
	s.current = lic
	s.mu.Unlock()
	return lic
}

func (s *LicenseService) evaluate() domain.License {
	envelope, err := os.ReadFile(s.licenseFile)
	if err != nil {
		return domain.Community() // no license installed → community
	}
	if len(strings.TrimSpace(string(envelope))) == 0 {
		return domain.Community() // empty LICENSE file → community, not an error
	}
	lic, err := s.validateAndParse(envelope)
	if err != nil {
		// A dev build (no injected key/salt) can't verify a LICENSE that happens to
		// be mounted — that's expected, not a license problem, so don't log it.
		if !errors.Is(err, errVerifyNotConfigured) {
			_ = catcher.Error("billing: license invalid", err, nil)
		}
		return domain.Community()
	}
	return lic
}

func (s *LicenseService) validateAndParse(envelope []byte) (domain.License, error) {
	instanceID := s.instanceID()
	if instanceID == "" || s.publicKey == "" || s.salt == "" {
		return domain.License{}, errVerifyNotConfigured
	}

	decrypted, err := lm.DecryptAndVerifyFromBase64(strings.TrimSpace(string(envelope)), []string{instanceID, s.salt}, s.publicKey)
	if err != nil {
		return domain.License{}, fmt.Errorf("decrypt/verify failed: %w", err)
	}

	var inner licenseInner
	if err := json.Unmarshal([]byte(decrypted), &inner); err != nil {
		return domain.License{}, fmt.Errorf("cannot parse license payload: %w", err)
	}

	if time.Now().After(inner.ExpiresAt) {
		return domain.License{}, fmt.Errorf("license expired at %s", inner.ExpiresAt.Format(time.RFC3339))
	}

	return domain.License{
		Edition:     domain.EditionEnterprise,
		MSSP:        inner.MSSP,
		Datasources: inner.Datasources,
		Type:        inner.Type,
		ExpiresAt:   inner.ExpiresAt,
	}, nil
}

func (s *LicenseService) Replace(envelope []byte) (domain.License, error) {
	if _, err := s.validateAndParse(envelope); err != nil {
		return domain.License{}, fmt.Errorf("%w: %v", domain.ErrInvalidLicense, err)
	}
	if err := s.writeLicenseFile(envelope); err != nil {
		return domain.License{}, fmt.Errorf("billing: cannot write license file: %w", err)
	}
	return s.Refresh(), nil
}

func (s *LicenseService) writeLicenseFile(envelope []byte) error {
	data := []byte(strings.TrimSpace(string(envelope)) + "\n")
	tmp := s.licenseFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.licenseFile)
}

func (s *LicenseService) instanceID() string {
	data, err := os.ReadFile(s.instanceFile)
	if err != nil {
		return ""
	}
	var inf struct {
		InstanceID string `yaml:"instance_id"`
	}
	if err := yaml.Unmarshal(data, &inf); err != nil {
		return ""
	}
	return inf.InstanceID
}
