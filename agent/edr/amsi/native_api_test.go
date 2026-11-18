package amsi

import (
	"context"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

func snapshotForTest(scan func(context.Context, []byte) (bool, string, error)) func(context.Context, []byte) (bool, string, string, error) {
	return func(ctx context.Context, data []byte) (bool, string, string, error) {
		clean, signature, err := scan(ctx, data)
		return clean, signature, "test-generation", err
	}
}

func TestIncompleteReportUsesActualScanGeneration(t *testing.T) {
	s, path := newTestScanner(t, config.Default())
	generation := "generation-a"
	s.scanSnapshot = func(context.Context, []byte) (bool, string, string, error) {
		return false, "", generation, engine.ErrUnsupported
	}
	for i := 0; i < 2; i++ {
		s.ScanBuffer("test-host", []byte("same content"))
	}
	generation = "generation-b"
	s.ScanBuffer("test-host", []byte("same content"))
	if got := len(spoolLines(t, path)); got != 2 {
		t.Fatalf("different actual generations must not share report suppression: %d", got)
	}
}

func TestBufferContextReachesNativeEngine(t *testing.T) {
	cfg := config.Default()
	cfg.FailMode = "closed"
	s, _ := newTestScanner(t, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.scanSnapshot = func(actual context.Context, _ []byte) (bool, string, string, error) {
		if actual != ctx {
			t.Fatal("caller context was replaced")
		}
		return false, "", "generation-a", actual.Err()
	}
	if !s.ScanBufferContext(ctx, "test-host", []byte("inert content")) {
		t.Fatal("fail-closed policy must also apply to cancelled analysis")
	}
}
