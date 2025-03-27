package service

import (
	"errors"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
	"github.com/utmstack/UTMStack/agent-manager/util"
)

type LastSeenService struct {
	mu     sync.RWMutex
	cache  map[string]time.Time
	repo   *repository.LastSeenRepository
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewLastSeenService() *LastSeenService {
	repo := repository.GetLastSeenRepository()
	return &LastSeenService{
		cache:  make(map[string]time.Time),
		repo:   repo,
		ticker: time.NewTicker(30 * time.Second),
		stopCh: make(chan struct{}),
	}
}

func (s *LastSeenService) Start() {
	pings, err := s.repo.GetAll()
	if err != nil {
		util.Logger.ErrorF("Failed to populate LastSeen cache: %v", err)
	} else {
		s.Populate(pings)
	}

	// Start the background process to flush the cache to the database
	go s.flushCachePeriodically()
}

func (s *LastSeenService) Stop() {
	// Stop the background process
	s.ticker.Stop()
	s.stopCh <- struct{}{}
}

func (s *LastSeenService) flushCacheToDB() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSeenObj := &models.LastSeen{}
	// Update the existing records in the database based on the cache data
	for key, lastSeen := range s.cache {
		lastSeenObj.Key = key
		lastSeenObj.LastPing = lastSeen
		err := s.repo.Update(lastSeenObj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *LastSeenService) flushCachePeriodically() {
	for {
		select {
		case <-s.ticker.C:
			// Flush the cache to the database
			err := s.flushCacheToDB()
			if err != nil {
				util.Logger.ErrorF("Failed to flush LastSeen cache to database: %v", err)
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *LastSeenService) Set(key string, lastSeen time.Time) error {
	s.mu.Lock()
	s.cache[key] = lastSeen
	s.mu.Unlock()
	return nil
}

func (s *LastSeenService) Get(key string) (models.LastSeen, error) {
	s.mu.Lock()
	last, ok := s.cache[key]
	if !ok {
		return models.LastSeen{}, errors.New("error getting value from cache")
	}
	s.mu.Unlock()
	return models.LastSeen{
		Key:      key,
		LastPing: last,
	}, nil
}

func (s *LastSeenService) Populate(lastPings []models.LastSeen) {
	for _, lastSeen := range lastPings {
		err := s.Set(lastSeen.Key, lastSeen.LastPing)
		if err != nil {
			continue
		}
	}
}

func (s *LastSeenService) GetLastSeen(key string) (models.LastSeen, error) {
	return s.Get(key)
}

func (s *LastSeenService) GetStatus(key string) (models.Status, string) {
	lastSeen, err := s.GetLastSeen(key)
	lastPing := lastSeen.LastPing.Format("2006-01-02 15:04:05")
	if err != nil {
		return models.Offline, lastPing
	}
	duration := time.Since(lastSeen.LastPing)
	if duration > time.Minute {
		return models.Offline, lastPing
	}
	return models.Online, lastPing
}
