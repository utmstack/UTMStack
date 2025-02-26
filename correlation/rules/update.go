package rules

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/utmstack/UTMStack/correlation/utils"
)

func Update(updateReady chan bool) {
	first := true
	for {
		log.Println("Downloading rules")
		cnf := utils.GetConfig()

		f, err := os.Stat(cnf.RulesFolder + "system")
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatalf("Could not get rules folder: %v", err)
			}
		}

		if f != nil {
			err := utils.RunCommand("rm", "-R", cnf.RulesFolder+"system")
			if err != nil {
				log.Fatalf("Could not remove rules folder: %v", err)
			}
		}

		err = utils.RunCommand("git", "clone", "https://github.com/utmstack/rules.git", cnf.RulesFolder+"system")
		if err != nil {
			log.Fatalf("Could not clone rules: %v", err)
		}

		if first {
			first = false
			updateReady <- true
		}

		log.Println("Rules updated, waiting for 72 hours to update again")
		time.Sleep(72 * time.Hour)
	}
}
