package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/socai/dto"
	"github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/modules/socai/verifier"
	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

var ErrVerificationFailed = errors.New("connection verification failed")

// threatWindsAIPath is the CM proxy path for the AI chat-completions endpoint
// (see backend/modules/threatintel: same reverse-proxy target, same id/key auth).
const threatWindsAIPath = "/proxy/api/ai/v1/chat/completions"
const threatWindsDefaultModel = "silas-1.0"

var ErrInstanceNotRegistered = errors.New("instance not registered yet — cannot use the ThreatWinds provider")

const ensureDefaultRetryInterval = 30 * time.Second

// StartEnsureDefaultLoop provisions the default ThreatWinds config in the
// background, retrying until it succeeds. A fresh install's backend can
// start before the updater service (a separate host process installed after
// the Docker stack — see installer/install.go) finishes registering the
// instance and writing instance-config.yml, so a single boot-time attempt
// isn't enough; this keeps checking every 30s until the instance is
// registered or a config already exists.
func (s *ConfigService) StartEnsureDefaultLoop() {
	if s.EnsureDefault() {
		return
	}
	go func() {
		for range time.Tick(ensureDefaultRetryInterval) {
			if s.EnsureDefault() {
				return
			}
		}
	}()
}

func (s *ConfigService) EnsureDefault() bool {
	existing, err := s.store.Load()
	if err != nil {
		catcher.Warn("socai: failed to check existing config for default provisioning", map[string]any{"error": err.Error()})
		return false
	}
	if existing != nil {
		return true
	}

	inst := instanceconfig.Get()
	if inst == nil || inst.Server == "" || inst.InstanceID == "" || inst.InstanceKey == "" {
		return false
	}

	idEnc, err := s.cipher.Encrypt(inst.InstanceID)
	if err != nil {
		catcher.Warn("socai: failed to encrypt default ThreatWinds credentials", map[string]any{"error": err.Error()})
		return false
	}
	keyEnc, err := s.cipher.Encrypt(inst.InstanceKey)
	if err != nil {
		catcher.Warn("socai: failed to encrypt default ThreatWinds credentials", map[string]any{"error": err.Error()})
		return false
	}

	fc := &repository.FileConfig{
		Provider:          "threatwinds",
		Model:             threatWindsDefaultModel,
		URL:               strings.TrimRight(inst.Server, "/") + threatWindsAIPath,
		AuthType:          "none",
		CustomHeaders:     map[string]string{"id": idEnc, "key": keyEnc},
		MaxTokens:         4096,
		MaxToolIterations: 12,
		AutoAnalyze:       true,
	}
	if err := s.store.Save(fc); err != nil {
		catcher.Warn("socai: failed to save default ThreatWinds config", map[string]any{"error": err.Error()})
		return false
	}
	catcher.Info("socai: auto-configured default ThreatWinds provider", nil)
	return true
}

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
		Configured:        true,
		Provider:          fc.Provider,
		Model:             fc.Model,
		URL:               fc.URL,
		APIKeySet:         fc.APIKey != "",
		AuthType:          fc.AuthType,
		AuthHeaderName:    fc.AuthHeaderName,
		CustomHeaders:     maskedHeaders,
		MaxTokens:         fc.MaxTokens,
		MaxToolIterations: fc.MaxToolIterations,
		AutoAnalyze:       fc.AutoAnalyze,
		Capabilities:      fc.Capabilities,
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

	url := req.URL
	authHeaderName := req.AuthHeaderName

	if req.Provider == "threatwinds" {
		inst := instanceconfig.Get()
		if inst == nil || inst.Server == "" || inst.InstanceID == "" || inst.InstanceKey == "" {
			return nil, ErrInstanceNotRegistered
		}
		url = strings.TrimRight(inst.Server, "/") + threatWindsAIPath
		authType = "none"
		authHeaderName = ""
		apiKeyPlain = ""
		headersPlain = map[string]string{"id": inst.InstanceID, "key": inst.InstanceKey}
	}

	// 2. Verify the connection before writing anything.
	if err := s.verifier.Verify(ctx, verifier.Config{
		Provider:       req.Provider,
		URL:            url,
		Model:          req.Model,
		APIKey:         apiKeyPlain,
		AuthType:       authType,
		AuthHeaderName: authHeaderName,
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
		Provider:          req.Provider,
		Model:             req.Model,
		URL:               url,
		APIKey:            apiKeyEnc,
		AuthType:          authType,
		AuthHeaderName:    authHeaderName,
		CustomHeaders:     headersEnc,
		MaxTokens:         maxTokens,
		MaxToolIterations: maxIters,
		AutoAnalyze:       req.AutoAnalyze,
		Capabilities:      req.Capabilities,
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
