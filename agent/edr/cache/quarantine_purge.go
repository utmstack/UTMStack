package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// QuarantinePurgeReceipt commits with the audit row's purged flag. The operation
// nonce binds filesystem-journal replay to this exact approved record, rather
// than treating any later purged flag as evidence that this operation finished.
type QuarantinePurgeReceipt struct {
	OperationID  string `gorm:"primaryKey"`
	QuarantineID string `gorm:"index"`
	RecordSHA256 string
	CompletedAt  time.Time
}

func purgeRecordDigest(rec QuarantineRecord) (string, error) {
	// These are the only fields changed by the purge transaction. Restored and
	// all admission metadata remain part of the approved record's identity.
	body, err := EncodeQuarantinePurgeRecord(rec)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func validateQuarantinePurge(db *gorm.DB, rec QuarantineRecord, operationID string) (bool, error) {
	if operationID == "" || rec.QuarantineID == "" || rec.Purged {
		return false, errors.New("invalid approved purge identity")
	}
	var current QuarantineRecord
	if err := db.First(&current, "quarantine_id = ?", rec.QuarantineID).Error; err != nil {
		return false, err
	}
	digest, err := purgeRecordDigest(rec)
	if err != nil {
		return false, err
	}
	currentDigest, err := purgeRecordDigest(current)
	if err != nil {
		return false, err
	}
	if currentDigest != digest {
		return false, fmt.Errorf("quarantine record %q changed during purge", rec.QuarantineID)
	}
	var receipt QuarantinePurgeReceipt
	err = db.First(&receipt, "operation_id = ?", operationID).Error
	if err == nil {
		if receipt.QuarantineID != rec.QuarantineID || receipt.RecordSHA256 != digest || receipt.CompletedAt.IsZero() || !current.Purged || !current.PurgedAt.Equal(receipt.CompletedAt) {
			return false, errors.New("purge receipt conflicts with the quarantine audit row")
		}
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if current.Purged {
		return false, errors.New("quarantine was purged without this operation's receipt")
	}
	return false, nil
}

// ValidateQuarantinePurge checks that destructive work still refers to the
// approved metadata, or that this exact operation has already committed.
func (c *Cache) ValidateQuarantinePurge(rec QuarantineRecord, operationID string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return validateQuarantinePurge(c.db, rec, operationID)
}

// CommitQuarantinePurge atomically records completion and its replay receipt.
// Replaying the same receipt preserves its original completion timestamp.
func (c *Cache) CommitQuarantinePurge(rec QuarantineRecord, operationID string, completedAt time.Time) error {
	if completedAt.IsZero() {
		return errors.New("purge completion time is required")
	}
	digest, err := purgeRecordDigest(rec)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Transaction(func(tx *gorm.DB) error {
		committed, err := validateQuarantinePurge(tx, rec, operationID)
		if err != nil || committed {
			return err
		}
		completedAt = completedAt.UTC()
		changed := tx.Model(&QuarantineRecord{}).Where("quarantine_id = ?", rec.QuarantineID).
			Updates(map[string]any{"purged": true, "purged_at": completedAt})
		if changed.Error != nil {
			return changed.Error
		}
		if changed.RowsAffected != 1 {
			return errors.New("purge audit row disappeared")
		}
		return tx.Create(&QuarantinePurgeReceipt{OperationID: operationID, QuarantineID: rec.QuarantineID, RecordSHA256: digest, CompletedAt: completedAt}).Error
	})
}
