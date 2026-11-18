package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

// engineCheckResult is the engine-check report. UnknownWhenReached counts
// the loaded rules that need analysis this engine does not perform, by
// reason; RepresentationsNotProduced lists the object representations the
// pipeline does not produce yet. Both make objects incomplete, never clean.
type engineCheckResult struct {
	Implementation             string         `json:"implementation"`
	Compatibility              string         `json:"compatibility"`
	Healthy                    bool           `json:"healthy"`
	Generation                 string         `json:"generation"`
	Signatures                 int            `json:"signatures"`
	EffectiveTuning            engine.Tuning  `json:"effective_tuning"`
	FileTypes                  int            `json:"file_types"`
	Skipped                    map[string]int `json:"skipped,omitempty"`
	UnknownWhenReached         map[string]int `json:"unknown_when_reached,omitempty"`
	RepresentationsNotProduced []string       `json:"representations_not_produced"`
	Error                      string         `json:"error,omitempty"`
}

func engineCheckReport(e *engine.Engine, err error) engineCheckResult {
	loaded := e.LoadReport()
	r := engineCheckResult{
		Implementation: engine.ImplementationVersion, Compatibility: "partial-signature-and-format-support",
		Healthy: err == nil, Generation: e.Version(), Signatures: e.SignatureCount(),
		EffectiveTuning: e.Tuning(),
		FileTypes:       loaded.FileTypes, Skipped: loaded.Skipped, UnknownWhenReached: loaded.Unknown,
		RepresentationsNotProduced: engine.RepresentationsNotProduced(),
	}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

// runEngineCheck validates the configured local generation without scanning or
// quarantining endpoint files. Exit 1 means the database cannot be activated.
func runEngineCheck() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "UTMStack EDR: invalid configuration:", err)
		os.Exit(1)
	}
	e := engine.New(cfg)
	err = e.EnsureRunning()
	_ = json.NewEncoder(os.Stdout).Encode(engineCheckReport(e, err))
	if err != nil {
		os.Exit(1)
	}
}
