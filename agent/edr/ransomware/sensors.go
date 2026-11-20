package ransomware

import (
	"os"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// SensorSet is the four fuzzy (X1) behavioral sensors. Each is fed every
// attributed modifying file event from Guard.OnFileEvent (before the canary
// early-return) and returns ready evidence, which the guard feeds to the
// scorer like any other signal. The sensors only add evidence; the guard's
// escalation ladder and containment handling are untouched.
type SensorSet struct {
	MassRate   *MassRateSensor
	Churn      *ChurnSensor
	Entropy    *EntropySensor
	RansomNote *RansomNoteSensor
}

// NewSensorSet builds the four fuzzy sensors from the guard's config. The
// config is read at construction (structural), matching the Scorer's own
// handling of thresholds; a nil now falls back to time.Now. A nil reader
// selects the default OS implementation for the entropy sensor.
func NewSensorSet(cfg config.EDRConfig, now func() time.Time, reader FileReader) *SensorSet {
	return &SensorSet{
		MassRate:   NewMassRateSensor(cfg, now),
		Churn:      NewChurnSensor(cfg, now),
		Entropy:    NewEntropySensor(cfg, now, reader),
		RansomNote: NewRansomNoteSensor(cfg, now),
	}
}

// Record fans one file event out to the four sensors and returns the evidence
// each of them produced (nil entries are dropped by the caller).
func (s *SensorSet) Record(fe FileEvent) []*Evidence {
	out := make([]*Evidence, 0, 4)
	if ev := s.MassRate.Record(fe); ev != nil {
		out = append(out, ev)
	}
	if ev := s.Churn.Record(fe); ev != nil {
		out = append(out, ev)
	}
	if ev := s.Entropy.Record(fe); ev != nil {
		out = append(out, ev)
	}
	if ev := s.RansomNote.Record(fe); ev != nil {
		out = append(out, ev)
	}
	return out
}

// behavior returns the configured tuning, overlaying A's defaults on any
// zero or negative knob so a config that never went through Load (or a
// hand-edited edr.json that sets only some fields) still behaves.
func behavior(cfg config.EDRConfig) config.Behavior {
	b := cfg.Ransomware.Behavior
	if b.MassRateWindowMs <= 0 {
		b.MassRateWindowMs = 5000
	}
	if b.MassRateMinWrites <= 0 {
		b.MassRateMinWrites = 25
	}
	if b.ChurnWindowMs <= 0 {
		b.ChurnWindowMs = 5000
	}
	if b.ChurnMinRenames <= 0 {
		b.ChurnMinRenames = 10
	}
	if b.ChurnMinDistinct <= 0 {
		b.ChurnMinDistinct = 5
	}
	if b.EntropySampleSize <= 0 {
		b.EntropySampleSize = 4096
	}
	if b.EntropySampleOffsets <= 0 {
		b.EntropySampleOffsets = 4
	}
	if b.EntropyBitsPerByte <= 0 {
		b.EntropyBitsPerByte = 7.5
	}
	if b.CooldownMs <= 0 {
		b.CooldownMs = 10000
	}
	return b
}

// fuzzyWeight returns the configured per-sensor evidence weight, falling back
// to the DefaultWeights entry for the kind when the config value is absent.
func fuzzyWeight(cfg config.EDRConfig, kind SignalKind) float64 {
	if w := cfg.Ransomware.Behavior.FuzzyWeightOf(); w > 0 {
		return float64(w)
	}
	return DefaultWeights[kind]
}

// FileReader is the injection point for entropy sampling. The default
// implementation stats the file, opens it, and does bounded ReadAt calls.
type FileReader func(path string, offset int64, size int) ([]byte, error)

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
