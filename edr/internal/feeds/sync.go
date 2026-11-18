// edr/internal/feeds/sync.go
package feeds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

// Syncer refreshes the mirrored indicator lists. Each (level, name) artifact
// is independent: one failure never blocks the others and never removes the
// previous good artifact (stale, never empty).
type Syncer struct {
	Root   string
	Levels []int
	Names  []string
	Fetch  func(ctx context.Context, level, name string) ([]byte, error)
	Rec    *mirror.Recorder
}

func (s *Syncer) SyncOnce(ctx context.Context) {
	for _, lvl := range s.Levels {
		level := fmt.Sprintf("level%d", lvl)
		for _, name := range s.Names {
			key := fmt.Sprintf("%s/%s", level, name)
			upstream, err := s.Fetch(ctx, level, name)
			if err != nil {
				log.Printf("feed %s: fetch failed: %v", key, err)
				s.Rec.FeedFailure(key, err.Error())
				continue
			}
			next, err := extractValues(upstream)
			if err != nil {
				log.Printf("feed %s: normalize failed: %v", key, err)
				s.Rec.FeedFailure(key, err.Error())
				continue
			}
			artifact := gzipValues(next)
			count := len(next)
			accRel := fmt.Sprintf("feeds/v1/download/list/%s/accumulative/%s", level, name)

			// Compute the daily delta from the previous accumulative on disk
			// BEFORE it is overwritten. A missing/unreadable/unparseable prev
			// → prev==nil → all-adds; that's fine because agents bootstrap from
			// the accumulative and dailies are additive. A daily-write failure
			// is best-effort and must not abort the accumulative (the
			// accumulative is the source of truth).
			var prev []string
			if prevBytes, rerr := os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(accRel))); rerr == nil {
				if pv, perr := parseGzipValues(prevBytes); perr == nil {
					prev = pv
				}
			}
			adds, dels := Diff(prev, next)
			dailyRel := fmt.Sprintf("feeds/v1/download/list/%s/daily/%s", level, name)
			if err := mirror.WriteArtifact(s.Root, dailyRel, RenderDaily(adds, dels, name)); err != nil {
				log.Printf("feed %s: daily write failed: %v", key, err)
			}

			if err := mirror.WriteArtifact(s.Root, accRel, artifact); err != nil {
				log.Printf("feed %s: write failed: %v", key, err)
				s.Rec.FeedFailure(key, err.Error())
				continue
			}
			sum := sha256.Sum256(artifact)
			s.Rec.FeedSuccess(key, count, hex.EncodeToString(sum[:]))
			log.Printf("feed %s: %d indicators mirrored (+%d/-%d daily)", key, count, len(adds), len(dels))
		}
	}
}
