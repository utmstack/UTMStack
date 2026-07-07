package cache

import "time"

// CanaryRecord is a planted decoy file. A write/rename/delete on Path is a
// high-confidence ransomware signal.
type CanaryRecord struct {
	Path     string `gorm:"primaryKey"`
	SHA256   string
	Volume   string
	PlacedAt time.Time
}

// RansomwareIncident is the audit trail of a contained (or alerted) ransomware
// escalation.
type RansomwareIncident struct {
	ID           string `gorm:"primaryKey"`
	PID          int
	Image        string
	Cmdline      string
	Signals      string // comma-separated contributing signal kinds
	Score        int
	Action       string // "alerted" | "suspended" | "contained"
	DetectedAt   time.Time
	QuarantineID string
}

func (c *Cache) StoreCanary(r CanaryRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.PlacedAt.IsZero() {
		r.PlacedAt = time.Now().UTC()
	}
	return c.db.Save(&r).Error
}

func (c *Cache) ListCanaries() ([]CanaryRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var rs []CanaryRecord
	err := c.db.Find(&rs).Error
	return rs, err
}

func (c *Cache) StoreIncident(r RansomwareIncident) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.DetectedAt.IsZero() {
		r.DetectedAt = time.Now().UTC()
	}
	return c.db.Save(&r).Error
}
