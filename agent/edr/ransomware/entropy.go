package ransomware

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// FileReader is the injection point for entropy sampling. The default
// implementation stats the file, opens it, and does bounded ReadAt calls.
type FileReader func(path string, offset int64, size int) ([]byte, error)

// EntropySensor samples a rewritten file at a few offsets and measures the
// Shannon entropy of the concatenated samples. Encrypted content is
// statistically flat (~8 bits/byte); normal text/code stays well below the
// default threshold (7.5). Total I/O is bounded by
// EntropySampleOffsets*EntropySampleSize (16 KiB at the defaults) — the
// sensor never reads whole files.
type EntropySensor struct {
	cfg    config.EDRConfig
	now    func() time.Time
	reader FileReader
	mu     sync.Mutex
	// lastFire maps "pid\x00path" → when the per-path cooldown expires.
	lastFire map[string]time.Time
}

// NewEntropySensor builds the sensor. A nil reader selects the default OS
// implementation; a nil now falls back to time.Now.
func NewEntropySensor(cfg config.EDRConfig, now func() time.Time, reader FileReader) *EntropySensor {
	if now == nil {
		now = time.Now
	}
	if reader == nil {
		reader = defaultFileReader
	}
	return &EntropySensor{cfg: cfg, now: now, reader: reader, lastFire: map[string]time.Time{}}
}

// defaultFileReader stats + ReadAt-bounded read of one sample window.
func defaultFileReader(path string, offset int64, size int) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if offset >= fi.Size() {
		return nil, nil
	}
	if int64(size) > fi.Size()-offset {
		size = int(fi.Size() - offset)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, size)
	n, err := f.ReadAt(buf, offset)
	if err != nil && n == 0 {
		return nil, err
	}
	return buf[:n], nil
}

// shannonEntropy returns the entropy of data in bits per byte.
func shannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	var freq [256]int
	for _, b := range data {
		freq[b]++
	}
	n := float64(len(data))
	ent := 0.0
	for _, c := range freq {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		ent -= p * math.Log2(p)
	}
	return ent
}

// Record samples one written file and returns evidence if its sampled
// entropy clears the threshold. It is a no-op for ops other than OpWrite,
// for paths that stat fails, and for paths inside the per-path cooldown.
func (e *EntropySensor) Record(ev FileEvent) *Evidence {
	if ev.Op != OpWrite {
		return nil
	}
	b := e.cfg.Ransomware.Behavior
	cooldown := time.Duration(b.CooldownMs) * time.Millisecond
	now := e.now()

	key := fmt.Sprintf("%d\x00%s", ev.PID, ev.Path)
	e.mu.Lock()
	if until, ok := e.lastFire[key]; ok && now.Before(until) {
		e.mu.Unlock()
		return nil
	}
	e.mu.Unlock()

	fi, err := os.Stat(ev.Path)
	if err != nil {
		return nil
	}
	sampleSize := b.EntropySampleSize
	if sampleSize <= 0 {
		sampleSize = 4096
	}
	offsets := b.EntropySampleOffsets
	if offsets <= 0 {
		offsets = 1
	}
	if fi.Size() < int64(sampleSize) {
		offsets = 1
	}

	var all []byte
	step := fi.Size() / int64(offsets+1)
	for i := 0; i < offsets; i++ {
		off := int64(i) * step
		if off < 0 {
			off = 0
		}
		data, err := e.reader(ev.Path, off, sampleSize)
		if err != nil {
			return nil
		}
		all = append(all, data...)
	}
	// Any successful sample puts the path in cooldown so a long-lived
	// high-entropy file re-samples at most once per CooldownMs.
	e.mu.Lock()
	e.lastFire[key] = now.Add(cooldown)
	e.mu.Unlock()

	ent := shannonEntropy(all)
	if ent < b.EntropyBitsPerByte {
		return nil
	}
	return &Evidence{
		PID:    ev.PID,
		Kind:   KindEntropy,
		Weight: float64(e.cfg.Ransomware.FuzzyWeight("entropy")),
		Detail: fmt.Sprintf("%.2f bits/byte", ent),
		TS:     now,
	}
}
