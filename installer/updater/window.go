package updater

import (
	"encoding/json"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
)

type MaintenanceWindow struct {
	Days      []int  `json:"days"`      // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime string `json:"startTime"` // Format: "HH:MM"
	EndTime   string `json:"endTime"`   // Format: "HH:MM"
}

var windowConfig *MaintenanceWindow

func UpdateWindowConfig() {
	for {
		window, _ := getWindowMaintaince()
		if window != nil {
			windowConfig = window
		}

		time.Sleep(config.CheckUpdatesEvery)
	}
}

func IsInMaintenanceWindow() bool {
	if windowConfig == nil {
		return true
	}

	if len(windowConfig.Days) == 0 {
		return true
	}

	if windowConfig.StartTime == "" || windowConfig.EndTime == "" {
		return true
	}

	startTime, err := time.Parse("15:04", windowConfig.StartTime)
	if err != nil {
		config.Logger().ErrorF("Error parsing start time %s: %v", windowConfig.StartTime, err)
		return true
	}

	endTime, err := time.Parse("15:04", windowConfig.EndTime)
	if err != nil {
		config.Logger().ErrorF("Error parsing end time %s: %v", windowConfig.EndTime, err)
		return true
	}

	now := time.Now()

	currentDay := int(now.Weekday())
	dayAllowed := false
	for _, day := range windowConfig.Days {
		if day == currentDay {
			dayAllowed = true
			break
		}
	}

	if !dayAllowed {
		return false
	}

	currentTime, _ := time.Parse("15:04", now.Format("15:04"))

	if startTime.Before(endTime) || startTime.Equal(endTime) {
		return !currentTime.Before(startTime) && !currentTime.After(endTime)
	}

	return !currentTime.Before(startTime) || !currentTime.After(endTime)
}

func getWindowMaintaince() (*MaintenanceWindow, error) {
	backConf, err := getConfigFromBackend(8)
	if err != nil {
		return nil, err
	}

	var window MaintenanceWindow
	err = json.Unmarshal([]byte(backConf[0].ConfParamValue), &window)
	if err != nil {
		return nil, err
	}

	return &window, nil
}
