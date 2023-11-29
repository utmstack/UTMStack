package service

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
)

type AgentLastSeenService struct {
	mu     sync.RWMutex
	cache  map[string]time.Time
	repo   *repository.AgentLastSeenRepository
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewAgentLastSeenService() *AgentLastSeenService {
	repo := repository.NewAgentLastSeenRepository()
	return &AgentLastSeenService{
		cache:  make(map[string]time.Time),
		repo:   repo,
		ticker: time.NewTicker(30 * time.Second),
		stopCh: make(chan struct{}),
	}
}

func (s *AgentLastSeenService) Start() {
	// Populate cache from the database
	agents, err := s.repo.GetAll()
	if err != nil {
		log.Printf("Failed to populate AgentLastSeen cache: %v", err)
	} else {
		s.Populate(agents)
	}

	// Start the background process to flush the cache to the database
	go s.flushCachePeriodically()
}

func (s *AgentLastSeenService) Stop() {
	// Stop the background process
	s.ticker.Stop()
	s.stopCh <- struct{}{}
}

func (c *AgentLastSeenService) flushCacheToDB() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update the existing records in the database based on the cache data
	for agentKey, lastSeen := range c.cache {
		agentLastSeen := &models.AgentLastSeen{
			AgentKey: agentKey,
			LastPing: lastSeen,
		}
		err := c.repo.Update(agentLastSeen)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AgentLastSeenService) flushCachePeriodically() {
	for {
		select {
		case <-s.ticker.C:
			// Flush the cache to the database
			err := s.flushCacheToDB()
			if err != nil {
				log.Printf("Failed to flush AgentLastSeen cache to database: %v", err)
			}
		case <-s.stopCh:
			return
		}
	}
}

func (c *AgentLastSeenService) Set(agentKey string, lastSeen time.Time) error {
	c.mu.Lock()
	c.cache[agentKey] = lastSeen
	c.mu.Unlock()
	return nil
}

func (c *AgentLastSeenService) Get(agentKey string) (models.AgentLastSeen, error) {
	c.mu.Lock()
	last, ok := c.cache[agentKey]
	if !ok {
		return models.AgentLastSeen{}, errors.New("error getting value from cache")
	}
	c.mu.Unlock()
	return models.AgentLastSeen{
		AgentKey: agentKey,
		LastPing: last,
	}, nil
}

func (c *AgentLastSeenService) Populate(lastPings []models.AgentLastSeen) {
	for _, lastSeen := range lastPings {
		err := c.Set(lastSeen.AgentKey, lastSeen.LastPing)
		if err != nil {
			continue
		}
	}
}
