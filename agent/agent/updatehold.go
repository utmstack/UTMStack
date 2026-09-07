package agent

import (
	"strconv"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/fs"
)

type updateHoldState struct {
	Hold bool `json:"hold"`
}

func SetUpdateHold(content string) error {
	hold, err := strconv.ParseBool(content)
	if err != nil {
		return err
	}
	return fs.WriteJSON(config.UpdateHoldFile, updateHoldState{Hold: hold})
}
