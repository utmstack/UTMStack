package rules

import (
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/utils"
)

func Update(updateReady chan bool) {
	first := true
	for {
		mu.Lock()
		h.Info("Downloading rules")
		cnf := utils.GetConfig()
		clone := exec.Command("git", "clone", "https://github.com/AtlasInsideCorp/UTMStackCorrelationRules.git", cnf.RulesFolder+"system")
		_ = clone.Run()

		pull := exec.Command("git", "pull")
		pull.Dir = filepath.Join(cnf.RulesFolder, "system")
		if err := pull.Run(); err != nil {
			h.Error("Could not update ruleset: %v", err)
		}

		if first {
			first = false
			updateReady <- true
		}
		h.Info("Rules updated")
		mu.Unlock()
		time.Sleep(24 * time.Hour)
	}
}
