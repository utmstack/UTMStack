package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/self/configuration"
	"github.com/utmstack/UTMStack/agent/self/utils"
)

func UpdateService() error {
	path := utils.GetMyPath()

	newbin := configuration.GetAgentBin("new")
	oldbin := configuration.GetAgentBin("old")
	err := os.Remove(filepath.Join(path, oldbin))
	if err != nil {
		return fmt.Errorf("error removing old %s: %v", oldbin, err)
	}

	err = os.Rename(filepath.Join(path, newbin), filepath.Join(path, oldbin))
	if err != nil {
		return fmt.Errorf("error renaming new %s: %v", newbin, err)
	}

	return nil
}
