package agent

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *AgentService) loadConnectionKeys() error {
	var rows []models.ConnectionKey
	if _, err := s.DBConnection.GetAll(&rows, ""); err != nil {
		return fmt.Errorf("failed to load connection keys: %v", err)
	}

	byTenant := make(map[string]models.ConnectionKey, len(rows)+1)
	for _, row := range rows {
		// Rows from before enrolment was per-tenant carry no tenant; they are
		// the default tenant's key.
		if row.TenantID == "" {
			row.TenantID = config.DefaultTenantID
		}
		byTenant[row.TenantID] = row
	}

	s.connKeyMutex.Lock()
	s.connKeys = byTenant
	s.connKeyMutex.Unlock()

	if _, ok := byTenant[config.DefaultTenantID]; !ok {
		if _, err := s.issueConnectionKey(config.DefaultTenantID); err != nil {
			return err
		}
		catcher.Info("Generated the default tenant's connection key", map[string]any{"process": "agent-manager"})
	}

	return nil
}

// issueConnectionKey creates or replaces a tenant's key and returns it.
func (s *AgentService) issueConnectionKey(tenantID string) (string, error) {
	key, err := generateConnectionKey()
	if err != nil {
		return "", err
	}

	s.connKeyMutex.RLock()
	existing, ok := s.connKeys[tenantID]
	s.connKeyMutex.RUnlock()

	row := models.ConnectionKey{TenantID: tenantID, Key: key}
	if ok {
		row.Model = gorm.Model{ID: existing.ID}
		if err := s.DBConnection.Upsert(&row, "id = ?",
			map[string]interface{}{"key": key, "tenant_id": tenantID}, existing.ID); err != nil {
			return "", fmt.Errorf("failed to persist connection key: %v", err)
		}
	} else if err := s.DBConnection.Create(&row); err != nil {
		return "", fmt.Errorf("failed to create connection key: %v", err)
	}

	s.connKeyMutex.Lock()
	s.connKeys[tenantID] = row
	s.connKeyMutex.Unlock()

	return key, nil
}

// TenantForConnectionKey reports which tenant a presented key enrols into.
// Comparison is constant-time against every key rather than a map lookup: a
// map would leak, through timing, whether a guessed prefix exists.
func (s *AgentService) TenantForConnectionKey(key string) (string, bool) {
	if key == "" {
		return "", false
	}

	s.connKeyMutex.RLock()
	defer s.connKeyMutex.RUnlock()

	for tenantID, row := range s.connKeys {
		if subtle.ConstantTimeCompare([]byte(key), []byte(row.Key)) == 1 {
			return tenantID, true
		}
	}
	return "", false
}

// ValidateConnectionKey reports whether the key enrols into any tenant.
func (s *AgentService) ValidateConnectionKey(key string) bool {
	_, ok := s.TenantForConnectionKey(key)
	return ok
}

func (s *AgentService) GetConnectionKey(ctx context.Context, req *ConnectionKeyRequest) (*ConnectionKeyResponse, error) {
	tenantID := tenantOrDefault(req.GetTenantId())

	s.connKeyMutex.RLock()
	row, ok := s.connKeys[tenantID]
	s.connKeyMutex.RUnlock()

	if ok {
		return &ConnectionKeyResponse{ConnectionKey: row.Key}, nil
	}

	// A tenant created after this service started has no key yet, and asking
	// for it is how one comes to exist.
	key, err := s.issueConnectionKey(tenantID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &ConnectionKeyResponse{ConnectionKey: key}, nil
}

func (s *AgentService) RotateConnectionKey(ctx context.Context, req *ConnectionKeyRequest) (*ConnectionKeyResponse, error) {
	tenantID := tenantOrDefault(req.GetTenantId())

	key, err := s.issueConnectionKey(tenantID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	catcher.Info("Rotated a tenant's connection key", map[string]any{
		"process": "agent-manager",
		"tenant":  tenantID,
	})
	return &ConnectionKeyResponse{ConnectionKey: key}, nil
}

func tenantOrDefault(tenantID string) string {
	if tenantID == "" {
		return config.DefaultTenantID
	}
	return tenantID
}

func generateConnectionKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %v", err)
	}
	return hex.EncodeToString(b), nil
}
