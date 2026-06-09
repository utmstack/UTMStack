package usecase

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/billing/dto"
	"gopkg.in/yaml.v3"
)

const (
	versionFileName  = "version.json"
	instanceFileName = "instance-config.yml"
)

type VersionUsecase struct {
	updatesDir string
}

func NewVersionUsecase(updatesDir string) *VersionUsecase {
	return &VersionUsecase{updatesDir: updatesDir}
}

type versionFile struct {
	Version   string `json:"version"`
	Edition   string `json:"edition"`
	Changelog string `json:"changelog"`
}

type instanceFile struct {
	Server     string `yaml:"server"`
	InstanceID string `yaml:"instance_id"`
}

func (u *VersionUsecase) Info() (dto.VersionInfo, error) {
	data, err := os.ReadFile(filepath.Join(u.updatesDir, versionFileName))
	if err != nil {
		return dto.VersionInfo{}, catcher.Error("billing: cannot read version file", err, nil)
	}
	var vf versionFile
	if err := json.Unmarshal(data, &vf); err != nil {
		return dto.VersionInfo{}, catcher.Error("billing: cannot parse version file", err, nil)
	}

	info := dto.VersionInfo{
		Version:   vf.Version,
		Edition:   vf.Edition,
		Changelog: vf.Changelog,
	}

	if data, err := os.ReadFile(filepath.Join(u.updatesDir, instanceFileName)); err == nil {
		var inf instanceFile
		if err := yaml.Unmarshal(data, &inf); err == nil {
			info.Server = inf.Server
			info.InstanceID = inf.InstanceID
		}
	}

	return info, nil
}
