package updater

import (
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

type PendingUpdate struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Edition     string `json:"edition"`
	Changelog   string `json:"changelog"`
	UpdateLocks string `json:"update_locks"`
}

func GetPendingUpdate() (*PendingUpdate, error) {
	if !utils.CheckIfPathExist(config.PendingUpdatesPath) {
		return nil, nil
	}

	var update PendingUpdate
	if err := utils.ReadJson(config.PendingUpdatesPath, &update); err != nil {
		return nil, err
	}

	return &update, nil
}

func SavePendingUpdate(update PendingUpdate) error {
	return utils.WriteJSON(config.PendingUpdatesPath, update)
}

func ClearPendingUpdate() error {
	if utils.CheckIfPathExist(config.PendingUpdatesPath) {
		return utils.Remove(config.PendingUpdatesPath)
	}
	return nil
}
