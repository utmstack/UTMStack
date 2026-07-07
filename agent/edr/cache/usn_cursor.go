package cache

type USNCursor struct {
	Volume    string `gorm:"primaryKey"`
	JournalID uint64
	NextUSN   int64
}

func (c *Cache) LoadUSN(volume string) (uint64, int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var cur USNCursor
	if err := c.db.First(&cur, "volume = ?", volume).Error; err != nil {
		return 0, 0
	}
	return cur.JournalID, cur.NextUSN
}

func (c *Cache) SaveUSN(volume string, journalID uint64, nextUSN int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.db.Save(&USNCursor{Volume: volume, JournalID: journalID, NextUSN: nextUSN}).Error
}
