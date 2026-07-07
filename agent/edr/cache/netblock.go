// edr/cache/netblock.go
package cache

import (
	"time"

	"gorm.io/gorm/clause"
)

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
