package service

import (
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CollectorService struct {
	repo            *repository.CollectorRepository
	lastSeenService *LastSeenService
}

func NewCollectorService(lastSeenService *LastSeenService) *CollectorService {
	repo := repository.NewCollectorRepository()
	return &CollectorService{repo: repo, lastSeenService: lastSeenService}
}

func (s *CollectorService) Create(collector *models.Collector) error {
	return s.repo.CreateCollector(collector)
}

func (s *CollectorService) Update(collector *models.Collector) error {
	return s.repo.UpdateCollector(collector)
}

func (s *CollectorService) Delete(key uuid.UUID, deletedBy string) (uint, error) {
	return s.repo.DeleteCollectorByKey(key, deletedBy)
}

func (s *CollectorService) FinByID(id uint) (*models.Collector, error) {
	return s.repo.GetCollectorById(id)
}
func (s *CollectorService) GetByKey(token string) (*models.Collector, error) {
	return s.repo.GetByKey(uuid.MustParse(token))
}
func (s *CollectorService) FindAll() ([]models.Collector, error) {
	return s.repo.GetAllCollectors()
}

// ListCollectors retrieves a paginated list of collectors based on the provided search criteria.
func (s *CollectorService) ListCollectors(p util.Pagination, f []util.Filter) ([]models.Collector, int64, error) {
	collectors, totalCount, err := s.repo.GetCollectorByFilter(p, f)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to retrieve collectors: %v", err)
	}
	return collectors, totalCount, err

}

func (s *CollectorService) GetCollectorLastSeen(collector models.Collector) (models.LastSeen, error) {
	return s.lastSeenService.Get(collector.CollectorKey)
}

func (s *CollectorService) GetCollectorStatus(collector models.Collector) (models.Status, string) {
	lastSeen, err := s.GetCollectorLastSeen(collector)
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
