package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	VerdictClean     = "clean"
	VerdictMalicious = "malicious"
	VerdictUnknown   = "unknown"
)

type VerdictRecord struct {
	SHA256       string `gorm:"primaryKey"`
	Verdict      string
	Signature    string
	SigDBVersion string
	FirstSeen    time.Time
	LastSeen     time.Time
	Stale        bool
}

type QuarantineRecord struct {
	QuarantineID  string `gorm:"primaryKey"`
	OriginalPath  string
	SHA256        string
	Detection     string
	Engine        string
	QuarantinedAt time.Time
	Restorable    bool
	Restored      bool
	// Purged marks a record whose stored file was permanently deleted (manual
	// admin purge or retention sweep). The audit record is kept; the file is gone.
	Purged   bool
	PurgedAt time.Time
}

type Cache struct {
	db *gorm.DB
	mu sync.RWMutex
}

func Open(dbFile string) (*Cache, error) {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&VerdictRecord{}, &QuarantineRecord{}, &USNCursor{}, &CanaryRecord{}, &RansomwareIncident{}, &IndicatorRecord{}, &NetBlockRecord{}, &FeedState{}); err != nil {
		return nil, err
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Lookup(sha256 string) (VerdictRecord, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var r VerdictRecord
	err := c.db.First(&r, "sha256 = ?", sha256).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return VerdictRecord{}, false, nil
	}
	if err != nil {
		return VerdictRecord{}, false, err
	}
	return r, true, nil
}

func (c *Cache) Store(r VerdictRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	if r.FirstSeen.IsZero() {
		r.FirstSeen = now
	}
	r.LastSeen = now
	return c.db.Save(&r).Error
}

func (c *Cache) StoreQuarantine(r QuarantineRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.QuarantinedAt.IsZero() {
		r.QuarantinedAt = time.Now().UTC()
	}
	return c.db.Save(&r).Error
}

func (c *Cache) GetQuarantine(id string) (QuarantineRecord, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var r QuarantineRecord
	err := c.db.First(&r, "quarantine_id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return QuarantineRecord{}, false, nil
	}
	if err != nil {
		return QuarantineRecord{}, false, err
	}
	return r, true, nil
}

// RecentQuarantineByPath returns the most recent unrestored quarantine whose
// OriginalPath matches path and that was quarantined within `within`. Used to
// recover a malicious verdict for a running process whose on-disk image was just
// quarantined (moved) out from under it, so the process guard can still kill it.
func (c *Cache) RecentQuarantineByPath(path string, within time.Duration) (QuarantineRecord, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var r QuarantineRecord
	cutoff := time.Now().UTC().Add(-within)
	err := c.db.Where("original_path = ? AND restored = ? AND quarantined_at >= ?", path, false, cutoff).
		Order("quarantined_at DESC").First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return QuarantineRecord{}, false, nil
	}
	if err != nil {
		return QuarantineRecord{}, false, err
	}
	return r, true, nil
}

func (c *Cache) ListQuarantine() ([]QuarantineRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var rs []QuarantineRecord
	err := c.db.Find(&rs).Error
	return rs, err
}

func (c *Cache) MarkRestored(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Model(&QuarantineRecord{}).Where("quarantine_id = ?", id).Update("restored", true).Error
}

// MarkPurged flags a record as permanently purged (its stored file deleted). The
// row is kept for audit.
func (c *Cache) MarkPurged(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Model(&QuarantineRecord{}).Where("quarantine_id = ?", id).
		Updates(map[string]any{"purged": true, "purged_at": time.Now().UTC()}).Error
}

// MarkStaleBySigDB marks clean records from an older signature DB as stale so
// they are re-scanned; malicious verdicts are left untouched (they persist).
func (c *Cache) MarkStaleBySigDB(currentSigDB string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := c.db.Model(&VerdictRecord{}).
		Where("verdict = ? AND sig_db_version <> ?", VerdictClean, currentSigDB).
		Update("stale", true)
	return res.RowsAffected, res.Error
}

func (c *Cache) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
