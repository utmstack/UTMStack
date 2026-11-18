// edr/internal/feeds/sync_test.go
package feeds

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

func TestSyncOnceWritesArtifactsAndIsolatesFailures(t *testing.T) {
	root := t.TempDir()
	s := &Syncer{
		Root:   root,
		Levels: []int{1},
		Names:  []string{"ip", "domain"},
		Rec:    mirror.NewRecorder(root),
		Fetch: func(_ context.Context, level, name string) ([]byte, error) {
			if name == "domain" {
				return nil, errors.New("upstream down")
			}
			return tarGzBytes(t, map[string]string{"m": "1.1.1.1\n"}), nil
		},
	}
	s.SyncOnce(context.Background())

	ipPath := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "accumulative", "ip")
	if _, err := os.Stat(ipPath); err != nil {
		t.Fatalf("ip artifact missing: %v", err)
	}
	if _, err := os.Stat(ipPath + ".sha256"); err != nil {
		t.Fatalf("ip checksum missing: %v", err)
	}
	domainPath := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "accumulative", "domain")
	if _, err := os.Stat(domainPath); !os.IsNotExist(err) {
		t.Fatal("failed feed must not produce an artifact")
	}
}

func TestSyncOnceFailureKeepsPreviousArtifact(t *testing.T) {
	root := t.TempDir()
	good := tarGzBytes(t, map[string]string{"m": "1.1.1.1\n"})
	fail := false
	s := &Syncer{
		Root: root, Levels: []int{1}, Names: []string{"ip"},
		Rec: mirror.NewRecorder(root),
		Fetch: func(_ context.Context, _, _ string) ([]byte, error) {
			if fail {
				return nil, errors.New("down")
			}
			return good, nil
		},
	}
	s.SyncOnce(context.Background())
	fail = true
	s.SyncOnce(context.Background())
	p := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "accumulative", "ip")
	if _, err := os.Stat(p); err != nil {
		t.Fatal("previous artifact must survive a failed sync")
	}
}

func TestSyncOnceWritesDailyDeltaBetweenSnapshots(t *testing.T) {
	root := t.TempDir()
	// First snapshot: 1.1.1.1 + 2.2.2.2. Second: 1.1.1.1 + 3.3.3.3
	// (2.2.2.2 removed, 3.3.3.3 added).
	second := false
	s := &Syncer{
		Root: root, Levels: []int{1}, Names: []string{"ip"},
		Rec: mirror.NewRecorder(root),
		Fetch: func(_ context.Context, _, _ string) ([]byte, error) {
			if second {
				return tarGzBytes(t, map[string]string{"m": "1.1.1.1\n3.3.3.3\n"}), nil
			}
			return tarGzBytes(t, map[string]string{"m": "1.1.1.1\n2.2.2.2\n"}), nil
		},
	}

	dailyPath := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "daily", "ip")

	// First run: prev is nil, so the daily is all-adds.
	s.SyncOnce(context.Background())
	b, err := os.ReadFile(dailyPath)
	if err != nil {
		t.Fatalf("first daily missing: %v", err)
	}
	first := parseDaily(t, b)
	wantFirst := []dailyLine{
		{"1.1.1.1", "ip", "add"},
		{"2.2.2.2", "ip", "add"},
	}
	if !reflect.DeepEqual(first, wantFirst) {
		t.Fatalf("first daily = %+v, want %+v", first, wantFirst)
	}
	// The daily has a verifiable .sha256 sibling.
	if _, err := os.Stat(dailyPath + ".sha256"); err != nil {
		t.Fatalf("first daily checksum missing: %v", err)
	}

	// Second run: exactly one add (3.3.3.3) and one del (2.2.2.2).
	second = true
	s.SyncOnce(context.Background())
	b, err = os.ReadFile(dailyPath)
	if err != nil {
		t.Fatalf("second daily missing: %v", err)
	}
	got := parseDaily(t, b)
	want := []dailyLine{
		{"3.3.3.3", "ip", "add"},
		{"2.2.2.2", "ip", "del"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("second daily = %+v, want %+v", got, want)
	}
}
