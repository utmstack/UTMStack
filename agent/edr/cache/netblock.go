// edr/cache/netblock.go
package cache

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MaxIndicatorRecords = 262144

// ReplaceIndicators makes removals durable as well as additions. The indicators
// and optional source receipts commit together; a failed insert restores the
// previous complete snapshot. Fixed bounds prevent oversized SQL batches/cache
// growth, including through callers outside the network-feed parser.
func (c *Cache) ReplaceIndicators(recs []IndicatorRecord, states ...FeedState) error {
	if len(recs) > MaxIndicatorRecords || len(states) > 64 {
		return errors.New("network indicator cache budget exceeded")
	}
	var total int64
	for _, r := range recs {
		if r.Value == "" || len(r.Value) > 1<<20 || len(r.ListName) > 256 {
			return errors.New("invalid network indicator cache record")
		}
		switch r.Type {
		case "ip", "cidr", "domain", "hostname":
		default:
			return errors.New("invalid network indicator cache type")
		}
		total += int64(len(r.Value) + len(r.Type) + len(r.ListName))
		if total > 128<<20 {
			return errors.New("network indicator cache byte budget exceeded")
		}
	}
	for _, s := range states {
		if s.Feed == "" || len(s.Feed) > 256 || len(s.Checksum) > 256 {
			return errors.New("invalid network indicator source receipt")
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&IndicatorRecord{}).Error; err != nil {
			return err
		}
		if len(recs) > 0 {
			if err := tx.CreateInBatches(&recs, 512).Error; err != nil {
				return err
			}
		}
		if len(states) > 0 {
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&states, 64).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type IndicatorRecord struct {
	Value     string `gorm:"primaryKey"`
	Type      string `gorm:"primaryKey"` // ip | cidr | domain | hostname
	Level     int
	ListName  string
	FirstSeen time.Time
	LastSeen  time.Time
	Removed   bool
}

type NetBlockRecord struct {
	ID          uint `gorm:"primaryKey"`
	Time        time.Time
	Direction   string
	RemoteIP    string
	RemotePort  int
	PID         int
	ProcessPath string
	Matched     string
	MatchedType string
	Level       int
	Action      string
	Domain      string
}

type FeedState struct {
	Feed         string `gorm:"primaryKey"`
	LastFullSync time.Time
	LastDelta    time.Time
	Checksum     string
	IndicatorNum int
}

func (c *Cache) PutIndicators(recs []IndicatorRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(recs) == 0 {
		return nil
	}
	return c.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&recs).Error
}

func (c *Cache) DeleteIndicator(value, typ string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Delete(&IndicatorRecord{}, "value = ? AND type = ?", value, typ).Error
}

func (c *Cache) ListIndicators() ([]IndicatorRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []IndicatorRecord
	err := c.db.Where("removed = ?", false).Find(&out).Error
	return out, err
}

func (c *Cache) RecordNetBlock(r NetBlockRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.Time.IsZero() {
		r.Time = time.Now().UTC()
	}
	return c.db.Create(&r).Error
}

func (c *Cache) SaveFeedState(f FeedState) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&f).Error
}

func (c *Cache) GetFeedState(feed string) (FeedState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var f FeedState
	err := c.db.First(&f, "feed = ?", feed).Error
	return f, err
}
