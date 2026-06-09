package agent

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *AgentService) loadConnectionKey() error {
	var rows []models.ConnectionKey
	if _, err := s.DBConnection.GetAll(&rows, ""); err != nil {
		return fmt.Errorf("failed to load connection key: %v", err)
	}

	if len(rows) > 0 {
		s.connKeyMutex.Lock()
		s.connKeyID = rows[0].ID
		s.connKey = rows[0].Key
		s.connKeyMutex.Unlock()
		return nil
	}

	key, err := generateConnectionKey()
	if err != nil {
		return err
	}
	row := models.ConnectionKey{Key: key}
	if err := s.DBConnection.Create(&row); err != nil {
		return fmt.Errorf("failed to create connection key: %v", err)
	}
	s.connKeyMutex.Lock()
	s.connKeyID = row.ID
	s.connKey = row.Key
	s.connKeyMutex.Unlock()
	catcher.Info("Generated agent connection key", map[string]any{"process": "agent-manager"})
	return nil
}

// ValidateConnectionKey reports whether the presented key matches the current
// connection key.
func (s *AgentService) ValidateConnectionKey(key string) bool {
	s.connKeyMutex.RLock()
	defer s.connKeyMutex.RUnlock()
	if s.connKey == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(key), []byte(s.connKey)) == 1
}

func (s *AgentService) GetConnectionKey(ctx context.Context, req *ConnectionKeyRequest) (*ConnectionKeyResponse, error) {
	s.connKeyMutex.RLock()
	key := s.connKey
	s.connKeyMutex.RUnlock()
	return &ConnectionKeyResponse{ConnectionKey: key}, nil
}

func (s *AgentService) RotateConnectionKey(ctx context.Context, req *ConnectionKeyRequest) (*ConnectionKeyResponse, error) {
	key, err := generateConnectionKey()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to generate connection key: %v", err))
	}

	s.connKeyMutex.RLock()
	id := s.connKeyID
	s.connKeyMutex.RUnlock()

	if err := s.DBConnection.Upsert(&models.ConnectionKey{Model: gorm.Model{ID: id}, Key: key}, "id = ?", map[string]interface{}{"key": key}, id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to persist connection key: %v", err))
	}

	s.connKeyMutex.Lock()
	s.connKey = key
	s.connKeyMutex.Unlock()

	catcher.Info("Rotated agent connection key", map[string]any{"process": "agent-manager"})
	return &ConnectionKeyResponse{ConnectionKey: key}, nil
}

func generateConnectionKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %v", err)
	}
	return hex.EncodeToString(b), nil
}
