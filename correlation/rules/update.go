package rules

import (
	"log"
	"os/exec"
	"time"

	"github.com/utmstack/UTMStack/correlation/utils"
)

func Update(updateReady chan bool) {
	first := true
	for {
		log.Println("Downloading rules")
		cnf := utils.GetConfig()

		rm := exec.Command("rm", "-R", cnf.RulesFolder+"system")
		_ = rm.Run()
		
		clone := exec.Command("git", "clone", "https://github.com/utmstack/rules.git", cnf.RulesFolder+"system")
		_ = clone.Run()

		if first {
			first = false
			updateReady <- true
		}
		
		log.Println("Rules updated")
		time.Sleep(24 * time.Hour)
	}
}

