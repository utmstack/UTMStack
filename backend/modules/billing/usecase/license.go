package usecase

import (
	"context"
	"encoding/json"
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

	instanceID := s.instanceID()
	if instanceID == "" || s.publicKey == "" || s.salt == "" {
		return domain.Community()
	}

	decrypted, err := lm.DecryptAndVerifyFromBase64(strings.TrimSpace(string(envelope)), []string{instanceID, s.salt}, s.publicKey)
	if err != nil {
		_ = catcher.Error("billing: license decrypt/verify failed", err, nil)
		return domain.Community()
	}

	var inner licenseInner
	if err := json.Unmarshal([]byte(decrypted), &inner); err != nil {
		_ = catcher.Error("billing: cannot parse license payload", err, nil)
		return domain.Community()
	}

	if time.Now().After(inner.ExpiresAt) {
		return domain.Community() // expired → community
	}

	return domain.License{
		Edition:     domain.EditionEnterprise,
		MSSP:        inner.MSSP,
		Datasources: inner.Datasources,
		Type:        inner.Type,
		ExpiresAt:   inner.ExpiresAt,
	}
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
