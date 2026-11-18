package cache

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// QuarantineRestoreReceipt records a verified publication, atomically with the
// restored flag. File identities bind replay to the independently copied output
// and its pinned directory; the original snapshot is retained until commit.
type QuarantineRestoreReceipt struct {
	OperationID  string `gorm:"primaryKey"`
	QuarantineID string `gorm:"index"`
	RecordSHA256 string
	ParentID     string
	OutputID     string
	CompletedAt  time.Time
}

func restoreRecordDigest(rec QuarantineRecord) (string, error) {
	rec.Restored = false
	return purgeRecordDigest(rec)
}

func validateQuarantineRestore(db *gorm.DB, rec QuarantineRecord, operationID, parentID, outputID string) (bool, error) {
	if operationID == "" || parentID == "" || outputID == "" || rec.QuarantineID == "" || !rec.Restorable || rec.Restored || rec.Purged {
		return false, errors.New("invalid approved restore identity")
	}
	var current QuarantineRecord
	if err := db.First(&current, "quarantine_id = ?", rec.QuarantineID).Error; err != nil {
		return false, err
	}
	digest, err := restoreRecordDigest(rec)
	if err != nil {
		return false, err
	}
	currentDigest, err := restoreRecordDigest(current)
	if err != nil {
		return false, err
	}
	if current.Purged || currentDigest != digest {
		return false, errors.New("quarantine record changed during restore")
	}
	var receipt QuarantineRestoreReceipt
	err = db.First(&receipt, "operation_id = ?", operationID).Error
	if err == nil {
		if receipt.QuarantineID != rec.QuarantineID || receipt.RecordSHA256 != digest || receipt.ParentID != parentID || receipt.OutputID != outputID || receipt.CompletedAt.IsZero() || !current.Restored {
			return false, errors.New("restore receipt conflicts with the approved publication")
		}
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if current.Restored {
		return false, errors.New("quarantine was restored without this operation's receipt")
	}
	return false, nil
}

func (c *Cache) ValidateQuarantineRestore(rec QuarantineRecord, operationID, parentID, outputID string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return validateQuarantineRestore(c.db, rec, operationID, parentID, outputID)
}

// CommitQuarantineRestore is called only after verified no-replace publication.
// A replay preserves the first completion time and never adopts another nonce.
func (c *Cache) CommitQuarantineRestore(rec QuarantineRecord, operationID, parentID, outputID string, at time.Time) error {
	if at.IsZero() {
		return errors.New("restore completion time is required")
	}
	digest, err := restoreRecordDigest(rec)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Transaction(func(tx *gorm.DB) error {
		committed, err := validateQuarantineRestore(tx, rec, operationID, parentID, outputID)
		if err != nil || committed {
			return err
		}
		changed := tx.Model(&QuarantineRecord{}).Where("quarantine_id = ?", rec.QuarantineID).Update("restored", true)
		if changed.Error != nil {
			return changed.Error
		}
		if changed.RowsAffected != 1 {
			return errors.New("restore audit row disappeared")
		}
		return tx.Create(&QuarantineRestoreReceipt{OperationID: operationID, QuarantineID: rec.QuarantineID, RecordSHA256: digest, ParentID: parentID, OutputID: outputID, CompletedAt: at.UTC()}).Error
	})
}
