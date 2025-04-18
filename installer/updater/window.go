package updater

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/utmstack/UTMStack/installer/config"
)

var windowConfig string

func UpdateWindowConfig() {
	for {
		window, err := getWindowMaintaince()
		if err != nil {
			config.Logger().ErrorF("Error getting maintenance window config: %v", err)
		}

		if window != "" {
			windowConfig = window
			config.Logger().Info("Updated maintenance window config: %s", windowConfig)
		}

		time.Sleep(config.CheckUpdatesEvery)
	}
}

func IsInMaintenanceWindow() bool {
	if windowConfig == "" {
		config.Logger().Info("Maintenance window config not set in backend, presuming 24/7 maintenance window")
		return true
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	schedule, err := parser.Parse(windowConfig)
	if err != nil {
		config.Logger().ErrorF("Error parsing cron expression %s: %v", windowConfig, err)
		return false
	}

	now := time.Now().Truncate(time.Minute)
	prev := schedule.Next(now.Add(-1 * time.Minute))

	return prev.Equal(now)
}

func getWindowMaintaince() (string, error) {
	backConf, err := getConfigFromBackend(8)
	if err != nil {
		return "", err
	}

	return backConf[0].ConfParamValue, nil
}
