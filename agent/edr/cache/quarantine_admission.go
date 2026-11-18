package cache

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (c *Cache) QuarantineByCaptureIdentity(id string) (QuarantineRecord, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var rec QuarantineRecord
	err := c.db.Where("source_capture_dir_id = ? AND (purged = ? OR purged IS NULL)", id, false).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return rec, false, nil
	}
	return rec, err == nil, err
}

// CommitQuarantine is idempotent across filesystem-journal replay. Recovery must
// never replace a previously restored/purged row with an older admission row.
func (c *Cache) CommitQuarantine(rec QuarantineRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rec).Error; err != nil {
		return err
	}
	var current QuarantineRecord
	if err := c.db.First(&current, "quarantine_id = ?", rec.QuarantineID).Error; err != nil {
		return err
	}
	if current.OriginalPath != rec.OriginalPath || current.SHA256 != rec.SHA256 || current.Detection != rec.Detection || current.SourceCaptureDir != rec.SourceCaptureDir || current.SourceCaptureDirID != rec.SourceCaptureDirID || current.SourceFileID != rec.SourceFileID || current.OriginalSecurity != rec.OriginalSecurity {
		return fmt.Errorf("quarantine id %q has conflicting admission metadata", rec.QuarantineID)
	}
	return nil
}
