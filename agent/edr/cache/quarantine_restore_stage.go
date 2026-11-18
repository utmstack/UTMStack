package cache

import (
	"bytes"
	"errors"

	"gorm.io/gorm"
)

// QuarantineRestoreStage is a bounded lookup hint, never a restore receipt.
// The caller must corroborate it with the private journal and live directory.
// RootPath is a blob to preserve arbitrary Unix filename bytes.
type QuarantineRestoreStage struct {
	StageID      string `gorm:"primaryKey"`
	OperationID  string `gorm:"uniqueIndex"`
	QuarantineID string `gorm:"index"`
	RootID       string
	RootPath     []byte
}

func validRestoreStage(v QuarantineRestoreStage) error {
	for _, id := range []string{v.StageID, v.OperationID, v.QuarantineID, v.RootID} {
		if len(id) == 0 || len(id) > 256 {
			return errors.New("invalid restore stage lookup identity")
		}
	}
	if len(v.RootPath) == 0 || len(v.RootPath) > 128<<10 {
		return errors.New("invalid restore stage lookup root")
	}
	return nil
}
func sameRestoreStage(a, b QuarantineRestoreStage) bool {
	return a.StageID == b.StageID && a.OperationID == b.OperationID && a.QuarantineID == b.QuarantineID && a.RootID == b.RootID && bytes.Equal(a.RootPath, b.RootPath)
}
func (c *Cache) TrackQuarantineRestoreStage(v QuarantineRestoreStage) error {
	if err := validRestoreStage(v); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Transaction(func(db *gorm.DB) error {
		var old QuarantineRestoreStage
		err := db.First(&old, "stage_id = ?", v.StageID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&v).Error
		}
		if err != nil {
			return err
		}
		if !sameRestoreStage(old, v) {
			return errors.New("restore stage lookup conflicts with another operation")
		}
		return nil
	})
}
func (c *Cache) QuarantineRestoreStageByIdentity(id string) (QuarantineRestoreStage, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var v QuarantineRestoreStage
	err := c.db.First(&v, "stage_id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return v, false, nil
	}
	if err != nil {
		return v, false, err
	}
	if err = validRestoreStage(v); err != nil {
		return v, false, err
	}
	return v, true, nil
}

// Forget removes only this exact operation's hint. Another operation's entry
// is preserved, including if a deleted directory's identity has been reused.
func (c *Cache) ForgetQuarantineRestoreStage(v QuarantineRestoreStage) error {
	if err := validRestoreStage(v); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Where("stage_id = ? AND operation_id = ? AND quarantine_id = ? AND root_id = ? AND root_path = ?", v.StageID, v.OperationID, v.QuarantineID, v.RootID, v.RootPath).Delete(&QuarantineRestoreStage{}).Error
}
