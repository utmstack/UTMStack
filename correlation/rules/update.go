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
		cnf := utils.GetConfig()

		if cnf.ConnectionMode == utils.ConnModeOffline {
			if _, err := os.Stat(cnf.RulesFolder + "system"); err == nil {
				log.Println("Offline mode: rules folder exists, skipping git clone")
				if first {
					first = false
					updateReady <- true
				}
				time.Sleep(48 * time.Hour)
				continue
			}
		}

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

		err = utils.RunCommand("git", "clone", "--depth", "1", "https://github.com/utmstack/rules.git", cnf.RulesFolder+"system")
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
