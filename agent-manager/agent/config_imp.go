package agent

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AgentService) loadConfigRegistry() error {
	var revisions []models.AgentConfigRevision
	if _, err := s.DBConnection.GetAll(&revisions, ""); err != nil {
		return err
	}
	revisionByKey := make(map[string]uint64, len(revisions))
	for _, r := range revisions {
		revisionByKey[r.Key] = r.Revision
	}

	var rows []models.AgentConfig
	if _, err := s.DBConnection.GetAll(&rows, ""); err != nil {
		return err
	}
	reg := make(map[string]map[uint]string, len(rows))
	for _, row := range rows {
		if reg[row.Key] == nil {
			reg[row.Key] = make(map[uint]string)
		}
		reg[row.Key][row.AgentID] = row.Content
	}

	s.configMutex.Lock()
	s.configRevisions = revisionByKey
	s.configRegistry = reg
	s.configMutex.Unlock()

	return nil
}

func (s *AgentService) SetAgentConfig(ctx context.Context, req *SetAgentConfigRequest) (*SetAgentConfigResponse, error) {
	key := req.GetKey()
	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	agentID := uint(req.GetAgentId())

	row := models.AgentConfig{Key: key, AgentID: agentID, Content: req.GetContent()}
	if err := s.DBConnection.Upsert(&row, "key = ? AND agent_id = ?",
		map[string]interface{}{"content": req.GetContent()}, key, agentID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist config %q for agent %d: %v", key, agentID, err)
	}

	revision, err := s.bumpConfigRevision(key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revision config %q: %v", key, err)
	}

	s.configMutex.Lock()
	if s.configRegistry[key] == nil {
		s.configRegistry[key] = make(map[uint]string)
	}
	s.configRegistry[key][agentID] = req.GetContent()
	s.configRevisions[key] = revision
	s.configMutex.Unlock()

	catcher.Info("Agent config updated", map[string]any{"key": key, "agent_id": agentID, "revision": revision, "process": "agent-manager"})

	return &SetAgentConfigResponse{Revision: revision}, nil
}

func (s *AgentService) bumpConfigRevision(key string) (uint64, error) {
	s.configMutex.RLock()
	revision := s.configRevisions[key] + 1
	s.configMutex.RUnlock()

	row := models.AgentConfigRevision{Key: key, Revision: revision}
	if err := s.DBConnection.Upsert(&row, "key = ?", map[string]interface{}{"revision": revision}, key); err != nil {
		return 0, err
	}
	return revision, nil
}

func (s *AgentService) diffConfig(agentID uint, clientRevisions map[string]uint64) []*ConfigUpdate {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()

	var updates []*ConfigUpdate
	for key, revision := range s.configRevisions {
		if clientRevisions[key] >= revision {
			continue
		}
		scoped := s.configRegistry[key]
		content, ok := scoped[agentID]
		if !ok {
			content, ok = scoped[models.GlobalAgentID]
		}
		if !ok {
			continue
		}
		updates = append(updates, &ConfigUpdate{Key: key, Revision: revision, Content: content})
	}
	return updates
}
