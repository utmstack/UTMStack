package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/socai/dto"
	"github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/modules/socai/verifier"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

var ErrVerificationFailed = errors.New("connection verification failed")

// connectionVerifier pings the LLM endpoint with the candidate credentials.
type connectionVerifier interface {
	Verify(ctx context.Context, c verifier.Config) error
}

type ConfigService struct {
	store    *repository.ConfigStore
	cipher   *secret.Cipher
	verifier connectionVerifier
}

func NewConfigService(store *repository.ConfigStore, cipher *secret.Cipher, v connectionVerifier) *ConfigService {
	return &ConfigService{store: store, cipher: cipher, verifier: v}
}

// Get returns the masked current config (never exposes stored secrets).
func (s *ConfigService) Get(_ context.Context) (*dto.ConfigResponse, error) {
	fc, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	if fc == nil {
		return &dto.ConfigResponse{Configured: false}, nil
	}

	maskedHeaders := make(map[string]string, len(fc.CustomHeaders))
	for k := range fc.CustomHeaders {
		maskedHeaders[k] = dto.MaskedValue
	}

	return &dto.ConfigResponse{
		Configured:                true,
		Provider:                  fc.Provider,
		Model:                     fc.Model,
		URL:                       fc.URL,
		APIKeySet:                 fc.APIKey != "",
		AuthType:                  fc.AuthType,
		AuthHeaderName:            fc.AuthHeaderName,
		CustomHeaders:             maskedHeaders,
		MaxTokens:                 fc.MaxTokens,
		MaxToolIterations:         fc.MaxToolIterations,
		AutoAnalyze:               fc.AutoAnalyze,
		ChangeAlertStatus:         fc.ChangeAlertStatus,
		AutomaticIncidentCreation: fc.AutomaticIncidentCreation,
		AllowedWriteTools:         fc.AllowedWriteTools,
	}, nil
}

// Update persists the config. Flow (mirrors the integrations tenant save):
// resolve masked secrets to plaintext → verify the LLM connection (fail fast,
// nothing written) → encrypt secrets → write the plugin YAML.
func (s *ConfigService) Update(ctx context.Context, req dto.ConfigRequest) (*dto.ConfigResponse, error) {
	existing, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	var prevAPIKey string
	var prevHeaders map[string]string
	if existing != nil {
		prevAPIKey = existing.APIKey
		prevHeaders = existing.CustomHeaders
	}

	// 1. Resolve secrets to plaintext (decrypt stored values left masked/empty).
	apiKeyPlain, err := s.resolvePlain(req.APIKey, prevAPIKey)
	if err != nil {
		return nil, err
	}
	headersPlain := make(map[string]string, len(req.CustomHeaders))
	for k, v := range req.CustomHeaders {
		p, err := s.resolvePlain(v, prevHeaders[k])
		if err != nil {
			return nil, err
		}
		headersPlain[k] = p
	}

	authType := req.AuthType
	if authType == "" {
		authType = "bearer"
	}

	// 2. Verify the connection before writing anything.
	if err := s.verifier.Verify(ctx, verifier.Config{
		Provider:       req.Provider,
		URL:            req.URL,
		Model:          req.Model,
		APIKey:         apiKeyPlain,
		AuthType:       authType,
		AuthHeaderName: req.AuthHeaderName,
		CustomHeaders:  headersPlain,
	}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVerificationFailed, err)
	}

	// 3. Encrypt secrets for storage.
	apiKeyEnc, err := s.cipher.Encrypt(apiKeyPlain)
	if err != nil {
		return nil, err
	}
	var headersEnc map[string]string
	if len(headersPlain) > 0 {
		headersEnc = make(map[string]string, len(headersPlain))
		for k, p := range headersPlain {
			enc, err := s.cipher.Encrypt(p)
			if err != nil {
				return nil, err
			}
			headersEnc[k] = enc
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	maxIters := req.MaxToolIterations
	if maxIters <= 0 {
		maxIters = 12
	}

	fc := &repository.FileConfig{
		Provider:                  req.Provider,
		Model:                     req.Model,
		URL:                       req.URL,
		APIKey:                    apiKeyEnc,
		AuthType:                  authType,
		AuthHeaderName:            req.AuthHeaderName,
		CustomHeaders:             headersEnc,
		MaxTokens:                 maxTokens,
		MaxToolIterations:         maxIters,
		AutoAnalyze:               req.AutoAnalyze,
		ChangeAlertStatus:         req.ChangeAlertStatus,
		AutomaticIncidentCreation: req.AutomaticIncidentCreation,
		AllowedWriteTools:         req.AllowedWriteTools,
	}

	if err := s.store.Save(fc); err != nil {
		return nil, err
	}
	return s.Get(ctx)
}

func (s *ConfigService) resolvePlain(incoming, prevEncrypted string) (string, error) {
	if incoming == "" || incoming == dto.MaskedValue {
		if prevEncrypted == "" {
			return "", nil
		}
		return s.cipher.Decrypt(prevEncrypted)
	}
	return incoming, nil
}
