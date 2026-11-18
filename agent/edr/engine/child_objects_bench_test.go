package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// These bounded microbenchmarks isolate allocation changes in child handling.
// They are not full-feed throughput or operating-system peak-RSS measurements.
func BenchmarkChildObjects(b *testing.B) {
	db, err := signatures.Compile(map[string][]byte{"child.ndb": []byte("Child.Marker:0:*:" + hx(formatCoverageMarker) + "\n")})
	if err != nil {
		b.Fatal(err)
	}
	state := func() *scanState {
		return &scanState{ctx: context.Background(), db: db, maxFile: 64 << 20, remaining: 256 << 20, maxFiles: 1000, maxDepth: 16}
	}
	b.Run("StoredZIP1MiB", func(b *testing.B) {
		payload := bytes.Repeat([]byte{0}, 1<<20)
		archive := formatRawZipFixture(b, zip0Stored, payload, payload)
		b.ReportAllocs()
		b.SetBytes(int64(len(payload)))
		b.ResetTimer()
		for b.Loop() {
			if sig, err := state().zip(archive, 0); sig != "" || err != nil {
				b.Fatalf("scan: %q, %v", sig, err)
			}
		}
	})
	b.Run("CSSFirstImageDetection12x64KiB", func(b *testing.B) {
		payload := append([]byte(formatCoverageMarker), bytes.Repeat([]byte{0}, (64<<10)-len(formatCoverageMarker))...)
		uri := "url(data:image/gif;base64," + base64.StdEncoding.EncodeToString(payload) + ")"
		page := []byte("<html><style>" + strings.Repeat("a{background:"+uri+"}", 12) + "</style></html>")
		b.ReportAllocs()
		b.SetBytes(int64(len(page)))
		b.ResetTimer()
		for b.Loop() {
			if sig, err := state().htmlViews(page, 0, signatures.TypeHTML); sig != "Child.Marker" || err != nil {
				b.Fatalf("scan: %q, %v", sig, err)
			}
		}
	})
}
