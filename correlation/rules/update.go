package rules

import (
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/correlation/utils"
)

func Update(updateReady chan bool) {
	first := true
	for {
		log.Println("Downloading rules")
		cnf := utils.GetConfig()
		clone := exec.Command("git", "clone", "https://github.com/AtlasInsideCorp/UTMStackCorrelationRules.git", cnf.RulesFolder+"system")
		_ = clone.Run()

		pull := exec.Command("git", "pull")
		pull.Dir = filepath.Join(cnf.RulesFolder, "system")
		if err := pull.Run(); err != nil {
			log.Printf("Could not update ruleset: %v", err)
		}

		if first {
			first = false
			updateReady <- true
		}
		log.Println("Rules updated")
		time.Sleep(24 * time.Hour)
	}
}

