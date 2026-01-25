package main

import (
	"github.com/threatwinds/go-sdk/catcher"
)

func main() {
	catcher.Info("Starting Compliance Orchestrator", map[string]any{
		"process": "compliance-orchestrator",
	})

	backend, err := bootstrap()
	if err != nil {
		return
	} else {
		catcher.Info("Compliance Orchestrator bootstrapped successfully", nil)
	}

	/*startWorkers()
	startLoop(backend)*/
}
