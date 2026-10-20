package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/archive"
	"github.com/utmstack/UTMStack/shared/fs"
	sharedhttp "github.com/utmstack/UTMStack/shared/http"
	"github.com/utmstack/UTMStack/shared/logger"
)

const EnginePackageVersion = "1.4.6"

const (
	provisionerName   = "utmstack_scan_engine"
	provisionerExt    = ".zip"
	versionFileName   = "version.json"
	defaultDaemonName = "clamd.exe"
)

func (e *Engine) provision() error {
	if e.cfg.ClamdPath != "" {
		return nil
	}
	if !provisionerSupported {
		return fmt.Errorf("UTMStack EDR: scan engine package not available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if e.cfg.Server == "" {
		return fmt.Errorf("UTMStack EDR: scan engine missing and no server registered to fetch it from")
	}

	binPath := filepath.Join(config.EngineDir, defaultDaemonName)
	versionPath := filepath.Join(config.EngineDir, versionFileName)
	if engineUpToDate(binPath, versionPath, EnginePackageVersion) {
		logger.Info("UTMStack EDR: scan engine already provisioned (version %s)", EnginePackageVersion)
		return nil
	}

	filename := e.packageFilename()
	url := fmt.Sprintf("https://%s:%s/private/dependencies/agent/%s", e.cfg.Server, config.MirrorPort, filename)
	verified, err := sharedhttp.DownloadFileAndVerify(url, nil, filename, config.InstallDir, false)
	if !verified || err != nil {
		_ = os.Remove(filepath.Join(config.InstallDir, filename))
		if err == nil {
			err = fmt.Errorf("download checksum not verified for %s; refusing to unpack", filename)
		}
		return err
	}
	defer os.Remove(filepath.Join(config.InstallDir, filename))

	if err := archive.Unzip(filepath.Join(config.InstallDir, filename), config.InstallDir); err != nil {
		return fmt.Errorf("unpacking scan engine package: %w", err)
	}
	if !fs.Exists(binPath) {
		return fmt.Errorf("scan engine package unpacked but %s is missing", defaultDaemonName)
	}
	if !engineUpToDate(binPath, versionPath, EnginePackageVersion) {
		return fmt.Errorf("scan engine package version mismatch: expected %s", EnginePackageVersion)
	}
	logger.Info("UTMStack EDR: scan engine provisioned (version %s)", EnginePackageVersion)
	return nil
}

func (e *Engine) packageFilename() string {
	return fmt.Sprintf("%s_%s_%s_%s%s", provisionerName, EnginePackageVersion, runtime.GOOS, runtime.GOARCH, provisionerExt)
}

func engineUpToDate(binPath, versionPath, expected string) bool {
	if !fs.Exists(binPath) {
		return false
	}
	raw, err := os.ReadFile(versionPath)
	if err != nil {
		return false
	}
	var v struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return false
	}
	return v.Version == expected
}
