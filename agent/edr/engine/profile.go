package engine

import (
	"runtime"

	sysinfo "github.com/elastic/go-sysinfo"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/logger"
)

// PlanTuning profiles the host and derives the engine tuning. If the host
// cannot be profiled for ANY reason (probe error, zero/garbage values, or a
// panic), it falls back to the failsafe SafeDefaultTuning() — a lean,
// functional config that is very unlikely to strain even a low-spec host.
func PlanTuning(cfg config.EDRConfig) (result Tuning) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("UTMStack EDR: host profiling panicked (%v); using failsafe engine config", r)
			result = SafeDefaultTuning()
		}
	}()

	ram, cores, freeTemp, err := profileHost()
	if err != nil || ram == 0 || cores < 1 {
		logger.Error("UTMStack EDR: host profiling unavailable (%v); using failsafe engine config", err)
		return SafeDefaultTuning()
	}
	return DeriveTuning(ram, cores, freeTemp, cfg.TierOverride)
}

func profileHost() (ramBytes uint64, cores int, freeTempBytes uint64, err error) {
	cores = runtime.NumCPU()
	host, err := sysinfo.Host()
	if err != nil {
		return 0, cores, 0, err
	}
	mem, err := host.Memory()
	if err != nil {
		return 0, cores, 0, err
	}
	// freeDiskBytes returns 0 when it can't measure; DeriveTuning then simply
	// skips the temp-based clamp (tier caps still bound the footprint).
	return mem.Total, cores, freeDiskBytes(tempDir()), nil
}
