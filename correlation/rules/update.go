package rules

import (
	"log"
	"os"
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
		err := rm.Run()
		if err != nil {
			log.Printf("Could not remove rules folder: %v", err)
			os.Exit(1)
		}

		clone := exec.Command("git", "clone", "https://github.com/utmstack/rules.git", cnf.RulesFolder+"system")
		err = clone.Run()
		if err != nil {
			log.Printf("Could not clone rules: %v", err)
			os.Exit(1)
		}

		if first {
			first = false
			updateReady <- true
		}

		log.Println("Rules updated, waiting for 72 hours to update again")
		time.Sleep(72 * time.Hour)
	}
}
