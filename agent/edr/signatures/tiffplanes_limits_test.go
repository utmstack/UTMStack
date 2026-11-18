package signatures

import (
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

func TestTIFFPlanesWorkCancellationAndInput(t *testing.T) {
	root := filepath.Join("testdata", "tiff_planes")
	names := []string{"planes-0-p1-r0-s0.tiff", "planes-6-p0-r2-s0.tiff", "planes-7-p1-r0-s0.tiff", "planes-9-p0-r0-s0.tiff", "planes-11-p0-r2-s1.tiff", "planes-12-p1-r0-s0.tiff", "planes-15-p1-r0-s1.tiff", "planes-17-p0-r2-s0.tiff"}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		before := sha256.Sum256(data)
		ctx, cancel := context.WithCancel(context.Background())
		work := &matchWork{ctx: ctx, limit: 1 << 40}
		p, err := readTIFFProfile(data, work)
		if err != nil || p.jpeg == nil {
			cancel()
			t.Fatalf("%s profile: %v", name, err)
		}
		work.ctx = &tiffFloatCancelContext{Context: ctx, cancel: cancel, calls: 5}
		if _, err := p.decode(); err == nil || ctx.Err() != context.Canceled {
			cancel()
			t.Fatalf("%s mid-decode cancellation: %v", name, err)
		}
		cancel()
		work.ctx = context.Background()
		work.limit = work.steps + 1
		if _, err := p.decode(); err == nil || work.steps > work.limit {
			t.Fatalf("%s budget: %v", name, err)
		}
		if sha256.Sum256(data) != before {
			t.Fatalf("%s changed input", name)
		}
	}
}
