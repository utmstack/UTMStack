//go:build !windows

package engine

// freeDiskBytes is a no-op off Windows (the engine only runs on Windows in
// Phase 1). Returning 0 makes DeriveTuning skip the temp-based clamp on the
// dev host, where tier caps already bound the footprint.
func freeDiskBytes(path string) uint64 { return 0 }
