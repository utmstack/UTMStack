package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/models"
)

func TestOpenConfiguresWALAndBatchOperations(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "agent_logs.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var journalMode string
	if err := db.db.Raw("PRAGMA journal_mode;").Scan(&journalMode).Error; err != nil {
		t.Fatalf("failed to read journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var busyTimeout int
	if err := db.db.Raw("PRAGMA busy_timeout;").Scan(&busyTimeout).Error; err != nil {
		t.Fatalf("failed to read busy_timeout: %v", err)
	}
	if busyTimeout < 5000 {
		t.Fatalf("busy_timeout = %d, want >= 5000", busyTimeout)
	}

	if err := db.Migrate(&models.Log{}); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	now := time.Now().UTC()
	batch := []*models.Log{
		{ID: "log-1", CreatedAt: now, DataSource: "host-a", Type: "linux_agent", Log: "msg-1", Processed: false},
		{ID: "log-2", CreatedAt: now, DataSource: "host-a", Type: "linux_agent", Log: "msg-2", Processed: false},
		{ID: "log-3", CreatedAt: now, DataSource: "host-a", Type: "syslog", Log: "msg-3", Processed: false},
	}
	if err := db.CreateBatch(batch); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	if err := db.MarkProcessedBatch([]string{"log-1", "log-3"}); err != nil {
		t.Fatalf("MarkProcessedBatch failed: %v", err)
	}

	var remaining []models.Log
	if _, err := db.Find(&remaining, "processed", false); err != nil {
		t.Fatalf("Find unprocessed failed: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "log-2" {
		t.Fatalf("unprocessed logs = %+v, want only log-2", remaining)
	}
}
