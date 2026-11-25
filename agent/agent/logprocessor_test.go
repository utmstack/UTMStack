package agent

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/database"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/go-sdk/plugins"
)

func TestDequeueBatchPrioritizesPlatformQueue(t *testing.T) {
	priorityQ := make(chan *plugins.Log, 10)
	standardQ := make(chan *plugins.Log, 10)

	for i := 0; i < 5; i++ {
		standardQ <- &plugins.Log{DataType: "syslog", Raw: "syslog-event"}
	}
	priorityQ <- &plugins.Log{DataType: "linux_agent", Raw: "auditd-1"}
	priorityQ <- &plugins.Log{DataType: "linux_agent", Raw: "auditd-2"}

	batch := dequeueBatch(context.Background(), priorityQ, standardQ, 4)
	if len(batch) != 4 {
		t.Fatalf("len(batch) = %d, want 4", len(batch))
	}
	if batch[0].Raw != "auditd-1" || batch[1].Raw != "auditd-2" {
		t.Fatalf("priority items were not dequeued first: got %q, %q", batch[0].Raw, batch[1].Raw)
	}
	if batch[2].DataType != "syslog" || batch[3].DataType != "syslog" {
		t.Fatalf("expected standard queue items after priority queue drained, got %q, %q", batch[2].DataType, batch[3].DataType)
	}
}

func TestDequeueBatchReturnsPromptlyOnContextCancel(t *testing.T) {
	priorityQ := make(chan *plugins.Log, 1)
	standardQ := make(chan *plugins.Log, 1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	batch := dequeueBatch(ctx, priorityQ, standardQ, 64)
	if len(batch) != 0 {
		t.Fatalf("expected empty batch on cancel, got %d items", len(batch))
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("dequeueBatch took %v after cancel, want < 500ms", elapsed)
	}
}

func TestFlushAcknowledgementsBatchesAndFlushesOnClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ack_test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(&models.Log{}); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	now := time.Now().UTC()
	if err := db.CreateBatch([]*models.Log{
		{ID: "a1", CreatedAt: now, DataSource: "h1", Type: "linux_agent", Log: "1", Processed: false},
		{ID: "a2", CreatedAt: now, DataSource: "h1", Type: "linux_agent", Log: "2", Processed: false},
		{ID: "a3", CreatedAt: now, DataSource: "h1", Type: "linux_agent", Log: "3", Processed: false},
	}); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	lp := &LogProcessor{db: db}
	ackCh := make(chan string, 4)
	ackCh <- "a1"
	ackCh <- "a3"
	close(ackCh)

	lp.flushAcknowledgements(ackCh)

	var remaining []models.Log
	if _, err := db.Find(&remaining, "processed", false); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "a2" {
		t.Fatalf("remaining unprocessed = %+v, want only a2", remaining)
	}
}
