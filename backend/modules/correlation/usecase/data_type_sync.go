package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

const (
	dataTypeSyncInterval     = 60 * time.Second
	dataTypeSyncInitialDelay = 10 * time.Second

	ibmAS400Type = "ibm-as400"
)

type DataTypeScheduler struct {
	repo   connectors.DataTypeRepository
	reader connectors.DataInputStatusReader
}

func NewDataTypeScheduler(repo connectors.DataTypeRepository, reader connectors.DataInputStatusReader) *DataTypeScheduler {
	return &DataTypeScheduler{repo: repo, reader: reader}
}

func (s *DataTypeScheduler) Start(ctx context.Context) {
	logger.Info("data-type-sync scheduler: starting (initial delay 10s)")

	select {
	case <-time.After(dataTypeSyncInitialDelay):
	case <-ctx.Done():
		logger.Info("data-type-sync scheduler: cancelled during initial delay — stopped")
		return
	}

	ticker := time.NewTicker(dataTypeSyncInterval)
	defer ticker.Stop()

	logger.Info("data-type-sync scheduler: running")

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			logger.Info("data-type-sync scheduler: context cancelled — stopped")
			return
		}
	}
}

func (s *DataTypeScheduler) tick(ctx context.Context) {
	newSources, err := s.reader.FindDataSourcesToConfigure(ctx, ibmAS400Type)
	if err != nil {
		logger.Error("data-type-sync: FindDataSourcesToConfigure error: " + err.Error())
		return
	}

	// Step 2: find non-system data types that are no longer reported as active
	// sources.
	orphans, err := s.repo.FindOrphanDataSourceConfigurations(ctx)
	if err != nil {
		logger.Error("data-type-sync: FindOrphanDataSourceConfigurations error: " + err.Error())
		return
	}

	// Step 3: insert new data types (included=true, systemOwner=false).
	if len(newSources) > 0 {
		toInsert := make([]domain.UtmDataTypes, 0, len(newSources))
		for _, src := range newSources {
			toInsert = append(toInsert, domain.UtmDataTypes{
				DataType:            src,
				DataTypeName:        src,
				DataTypeDescription: "Synchronized datatype",
				Included:            true,
				SystemOwner:         false,
			})
		}
		if err := s.repo.SaveAll(ctx, toInsert); err != nil {
			logger.Error("data-type-sync: SaveAll (new sources) error: " + err.Error())
			return
		}
		logger.Info("data-type-sync: inserted " + itoa(len(toInsert)) + " new data type(s)")
	}

	// Step 4: set included=false on orphan data types.
	if len(orphans) > 0 {
		for i := range orphans {
			orphans[i].Included = false
		}
		if err := s.repo.SaveAll(ctx, orphans); err != nil {
			logger.Error("data-type-sync: SaveAll (orphans) error: " + err.Error())
			return
		}
		logger.Info("data-type-sync: disabled " + itoa(len(orphans)) + " orphan data type(s)")
	}
}

func itoa(n int) string {
	return intToStr(n)
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
