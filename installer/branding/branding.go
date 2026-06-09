package branding

import (
	"strings"
	"sync"

	"path/filepath"

	"github.com/utmstack/UTMStack/installer/utils"
)

const DefaultName = "UTMStack"

// Brand mirrors the text fields the backend branding accepts; image files
// (logo, logoDark, favicon, reportLogo, reportCover) live alongside brand.yaml
// and are discovered by name (see seed.go).
type Brand struct {
	ProductName string `yaml:"productName"`
}

var (
	brand     *Brand
	brandOnce sync.Once
)

func Dir() string {
	return filepath.Join(utils.GetMyPath(), "branding")
}

func configPath() string {
	return filepath.Join(Dir(), "brand.yaml")
}

func LicensePath() string {
	return filepath.Join(utils.GetMyPath(), "LICENSE")
}

func Get() *Brand {
	brandOnce.Do(func() {
		b := &Brand{ProductName: DefaultName}
		if utils.CheckIfPathExist(configPath()) {
			_ = utils.ReadYAML(configPath(), b)
			if strings.TrimSpace(b.ProductName) == "" {
				b.ProductName = DefaultName
			}
		}
		brand = b
	})
	return brand
}

func Name() string {
	return Get().ProductName
}

func IsCustom() bool {
	return utils.CheckIfPathExist(configPath())
}
