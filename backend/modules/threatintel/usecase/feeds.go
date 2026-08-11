package usecase

import (
	"github.com/utmstack/utmstack/backend/modules/threatintel/dto"
	"github.com/utmstack/utmstack/backend/modules/threatintel/repository"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

// FeedsService owns what the feeds plugin runs on. The plugin reads the file
// and never asks for it: a secret served over the config API comes back
// withheld, which is what left the plugin re-registering with ThreatWinds on
// every boot.
type FeedsService struct {
	store  *repository.ConfigStore
	cipher *secret.Cipher
}

func NewFeedsService(store *repository.ConfigStore, cipher *secret.Cipher) *FeedsService {
	return &FeedsService{store: store, cipher: cipher}
}

func (s *FeedsService) Status() (dto.FeedsStatus, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return dto.FeedsStatus{}, err
	}
	return dto.FeedsStatus{
		Enabled:    cfg.Enabled,
		Configured: cfg.APIKey != "" && cfg.APISecret != "",
	}, nil
}

// SetEnabled decides whether this instance's incidents leave it. Credentials
// are kept either way, so turning it back on does not mean registering again.
func (s *FeedsService) SetEnabled(enabled bool) error {
	return s.store.Update(func(c *repository.FeedsConfig) { c.Enabled = enabled })
}

func (s *FeedsService) SaveCredentials(apiKey, apiSecret string) error {
	key, err := s.cipher.Encrypt(apiKey)
	if err != nil {
		return err
	}
	sec, err := s.cipher.Encrypt(apiSecret)
	if err != nil {
		return err
	}
	return s.store.Update(func(c *repository.FeedsConfig) {
		c.APIKey = key
		c.APISecret = sec
	})
}
