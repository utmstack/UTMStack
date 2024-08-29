package agent

import (
	"fmt"
	"io"
	"strconv"
	sync "sync"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	LastSeenServ    *LastSeenService
	lastSeenOnce    sync.Once
	LastSeenChannel chan models.LastSeen
)

type LastSeenService struct {
	UnimplementedPingServiceServer

	CacheAgentLastSeen          map[uint]models.LastSeen
	CacheAgentLastSeenMutex     sync.Mutex
	CacheCollectorLastSeen      map[uint]models.LastSeen
	CacheCollectorLastSeenMutex sync.Mutex

	DBConnection *database.DB
}

func InitLastSeenService() *LastSeenService {
	lastSeenOnce.Do(func() {
		LastSeenServ = &LastSeenService{
			CacheAgentLastSeen:     make(map[uint]models.LastSeen),
			CacheCollectorLastSeen: make(map[uint]models.LastSeen),
			DBConnection:           database.GetDB(),
		}
		LastSeenChannel = make(chan models.LastSeen, 1000)
		LastSeenServ.InitPingSync()
	})
	return LastSeenServ
}

func (s *LastSeenService) InitPingSync() {
	pings := []models.LastSeen{}
	go s.processPings()
	for {
		_, err := s.DBConnection.GetAll(&pings, "")
		if err != nil {
			utils.ALogger.ErrorF("failed to get LastSeen items: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	for _, ping := range pings {
		LastSeenChannel <- ping
	}

	go s.flushLastSeenToDB()
}

func (s *LastSeenService) processPings() {
	for ping := range LastSeenChannel {
		if ping.ConnectorType == "agent" {
			s.CacheAgentLastSeenMutex.Lock()
			s.CacheAgentLastSeen[ping.ConnectorID] = ping
			s.CacheAgentLastSeenMutex.Unlock()
		} else if ping.ConnectorType == "collector" {
			s.CacheCollectorLastSeenMutex.Lock()
			s.CacheCollectorLastSeen[ping.ConnectorID] = ping
			s.CacheCollectorLastSeenMutex.Unlock()
		}
	}
}

func (s *LastSeenService) flushLastSeenToDB() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		pings := []models.LastSeen{}

		s.CacheAgentLastSeenMutex.Lock()
		for _, lastSeen := range s.CacheAgentLastSeen {
			pings = append(pings, lastSeen)
		}
		s.CacheAgentLastSeenMutex.Unlock()

		s.CacheCollectorLastSeenMutex.Lock()
		for _, lastSeen := range s.CacheCollectorLastSeen {
			pings = append(pings, lastSeen)
		}
		s.CacheCollectorLastSeenMutex.Unlock()

		for _, ping := range pings {
			err := s.DBConnection.Upsert(&ping, "connector_id = ?", nil, ping.ConnectorID)
			if err != nil {
				utils.ALogger.ErrorF("failed to save LastSeen item: %v", err)
			}
		}
	}
}

func (s *LastSeenService) Ping(stream PingService_PingServer) error {
	id, _, typ, err := utils.GetItemsFromContext(stream.Context())
	if err != nil {
		return err
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid id")
	}

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		LastSeenChannel <- models.LastSeen{
			ConnectorID:   uint(idInt),
			ConnectorType: typ,
			LastPing:      time.Now(),
		}
	}
}

func (s *LastSeenService) GetLastSeenStatus(id uint, typ string) (Status, string, error) {
	lastSeen := models.LastSeen{}
	ok := false
	switch typ {
	case "agent":
		s.CacheAgentLastSeenMutex.Lock()
		lastSeen, ok = s.CacheAgentLastSeen[id]
		s.CacheAgentLastSeenMutex.Unlock()
	case "collector":
		s.CacheCollectorLastSeenMutex.Lock()
		lastSeen, ok = s.CacheCollectorLastSeen[id]
		s.CacheCollectorLastSeenMutex.Unlock()
	default:
		return Status_OFFLINE, "", status.Error(codes.InvalidArgument, "invalid type")
	}

	if !ok {
		return Status_OFFLINE, "", fmt.Errorf("failed to get last seen for %s %d", typ, id)
	}

	lastPing := lastSeen.LastPing.Format("2006-01-02 15:04:05")
	duration := time.Since(lastSeen.LastPing)
	if duration > time.Minute {
		return Status_OFFLINE, lastPing, nil
	}
	return Status_ONLINE, lastPing, nil
}
