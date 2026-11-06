package database

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDatabase also points config.LogsDBFile (what GetDatabaseSizeInMB
// reads) at the same temp file, since DeleteOld's cap checks go through that
// package-level path rather than the *Database it's called on.
func newTestDatabase(t *testing.T) *Database {
	t.Helper()
	path := filepath.Join(t.TempDir(), "logs.db")
	previous := config.LogsDBFile
	config.LogsDBFile = path
	t.Cleanup(func() { config.LogsDBFile = previous })

	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&models.Log{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &Database{db: conn}
}

// ING-F7: deleteOldestBatch is what DeleteOld relies on to evict "just
// enough" rows to get back under the hard disk cap. Plain SQLite (including
// the pure-Go glebarez/sqlite driver this agent uses) doesn't support a
// DELETE ... LIMIT clause — a prior version of this code called
// .Limit(n).Delete(data) directly, which GORM/SQLite silently ignore the
// Limit on, deleting every matching row regardless of n. The fix wraps the
// bound in a SELECT subquery (LIMIT works there) and deletes WHERE id IN
// (that subquery) — this test proves exactly that bound holds today, with
// far more candidate rows than the batch size so a regression back to
// "delete everything" would be caught immediately.
func TestDeleteOldestBatchBoundsRowCount(t *testing.T) {
	db := newTestDatabase(t)
	const total = deleteBatchSize*2 + 200

	base := time.Now().Add(-time.Duration(total) * time.Second)
	for i := 0; i < total; i++ {
		row := models.Log{
			ID:        fmt.Sprintf("id-%04d", i),
			CreatedAt: base.Add(time.Duration(i) * time.Second), // ascending: id-0000 is oldest
			Processed: false,
		}
		if err := db.db.Create(&row).Error; err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	result := db.deleteOldestBatch(&models.Log{}, "processed = ?", false)
	if result.Error != nil {
		t.Fatalf("deleteOldestBatch: %v", result.Error)
	}
	if result.RowsAffected != int64(deleteBatchSize) {
		t.Fatalf("deleted %d rows in one call, want exactly %d (an unbounded DELETE would have wiped all %d)",
			result.RowsAffected, deleteBatchSize, total)
	}

	var remaining int64
	if err := db.db.Model(&models.Log{}).Count(&remaining).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if want := int64(total - deleteBatchSize); remaining != want {
		t.Fatalf("remaining rows = %d, want %d", remaining, want)
	}

	// Oldest-first: the surviving rows must be the newest ones, i.e. id-0000
	// (the very oldest) must be gone and the newest row must still be there.
	var oldestStillThere, newestStillThere int64
	db.db.Model(&models.Log{}).Where("id = ?", "id-0000").Count(&oldestStillThere)
	db.db.Model(&models.Log{}).Where("id = ?", fmt.Sprintf("id-%04d", total-1)).Count(&newestStillThere)
	if oldestStillThere != 0 {
		t.Fatalf("the oldest row survived a batch delete meant to evict oldest-first")
	}
	if newestStillThere != 1 {
		t.Fatalf("the newest row did not survive a batch delete meant to evict oldest-first")
	}
}

// DeleteOld must stop issuing hard-cap deletes once the remaining rows no
// longer exceed the cap, rather than always running until nothing is left —
// the other half of ING-F7's "not just enough" complaint.
func TestDeleteOldDoesNotDeleteEverythingWhenNothingIsOverTheSoftCap(t *testing.T) {
	db := newTestDatabase(t)
	row := models.Log{ID: "only-row", CreatedAt: time.Now(), Processed: true}
	if err := db.db.Create(&row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	// A huge retention figure means both the soft cap and the hard cap
	// (retention * HardCapMultiplier) are far above this tiny table's size.
	processedDeleted, unprocessedDeleted, err := db.DeleteOld(&models.Log{}, 1_000_000)
	if err != nil {
		t.Fatalf("DeleteOld: %v", err)
	}
	if processedDeleted != 0 || unprocessedDeleted != 0 {
		t.Fatalf("deleted %d/%d rows, want 0/0 (nothing is over either cap)", processedDeleted, unprocessedDeleted)
	}

	var remaining int64
	db.db.Model(&models.Log{}).Count(&remaining)
	if remaining != 1 {
		t.Fatalf("remaining rows = %d, want 1", remaining)
	}
}
