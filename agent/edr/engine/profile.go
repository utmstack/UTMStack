package engine

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/logger"
)

// PlanTuning profiles the host and derives the engine tuning. If the host
// cannot be profiled for ANY reason (probe error, zero/garbage values, or a
// panic), it falls back to the failsafe SafeDefaultTuning() — a lean,
// functional scan budget. Compiled database memory is measured separately.
func PlanTuning(cfg config.EDRConfig) (result Tuning) {
	return planTuningWithProfiler(cfg, profileHost)
}

func planTuningWithProfiler(cfg config.EDRConfig, probe func() (uint64, int, error)) (result Tuning) {
	profileAvailable := false
	defer func() {
		if r := recover(); r != nil {
			logger.Error("UTMStack EDR: host profiling panicked (%v); using failsafe engine config", r)
			result = SafeDefaultTuning()
		}
		result = ApplyScanLimits(result, cfg, profileAvailable)
	}()

	ram, cores, err := probe()
	if err != nil || ram == 0 || cores < 1 {
		logger.Error("UTMStack EDR: host profiling unavailable (%v); using failsafe engine config", err)
		return SafeDefaultTuning()
	}
	profileAvailable = true
	return DeriveTuning(ram, cores, cfg.TierOverride)
}
