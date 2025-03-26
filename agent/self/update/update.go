package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/self/config"
	"github.com/utmstack/UTMStack/agent/self/utils"
)

func RunUpdate() error {
	path := utils.GetMyPath()

	newbin := fmt.Sprintf(config.ServiceFile, "_new")
	oldbin := fmt.Sprintf(config.ServiceFile, "")
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
