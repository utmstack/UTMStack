package updates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/updater/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/http"
	"github.com/utmstack/UTMStack/shared/logger"
	"github.com/utmstack/UTMStack/shared/svc"
)

const (
	checkEvery = 5 * time.Minute
)

// updateHealthWait is how long the updater waits after starting the new
// binary before treating the service as healthy. Package var so tests can
// shorten it; 30s matches the original agent behavior.
var updateHealthWait = 30 * time.Second

// edrBaseName mirrors dependency.EDRFile's naming: utmstack_edr_<os>_<arch>.
// The dependency package cannot be imported here (it imports agent/agent), so
// the convention is mirrored, exactly like legacyServiceFile for the agent.
const edrBaseName = "utmstack_edr"

// edrServiceName mirrors agent/edr/config.ServiceName — the EDR kardianos
// service name. Mirrored for the same module-boundary reason.
const edrServiceName = "UTMStackEDR"

// edrServiceFile returns the EDR binary name with OS/arch suffix and the given
// role suffix: "" (installed), "_new" (downloaded), ".old" (backup).
func edrServiceFile(suffix string) string {
	name := fmt.Sprintf("%s_%s_%s%s", edrBaseName, runtime.GOOS, runtime.GOARCH, suffix)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// updateTarget describes one updatable service: its binary files and service
// name. The swap logic is identical for every target — only these names
// differ. Version tracking lives in the shared version.json; the swap
// optionally promotes it (see promoteVersion in runUpdate).
type updateTarget struct {
	ServiceName string
	OldBin      string
	NewBin      string
	BackupBin   string
	Migrate     func(basePath string) error // legacy naming migration; nil if none
}

// agentTarget describes the agent binary update (unchanged file names).
func agentTarget() updateTarget {
	return updateTarget{
		ServiceName: config.SERV_AGENT_NAME,
		OldBin:      config.ServiceFile(""),
		NewBin:      config.ServiceFile("_new"),
		BackupBin:   config.ServiceFile(".old"),
		Migrate:     migrateLegacyAgent,
	}
}

// edrTarget describes the EDR binary update. Both binaries live in the same
// install dir and share the version.json file (each reads its own field);
// only binary names and the service differ.
func edrTarget() updateTarget {
	return updateTarget{
		ServiceName: edrServiceName,
		OldBin:      edrServiceFile(""),
		NewBin:      edrServiceFile("_new"),
		BackupBin:   edrServiceFile(".old"),
		Migrate:     migrateLegacyEDR,
	}
}

func migrateLegacyAgent(basePath string) error {
	t := agentTarget()
	if fs.Exists(filepath.Join(basePath, t.OldBin)) {
		return nil
	}
	legacyPath := filepath.Join(basePath, legacyServiceFile())
	if !fs.Exists(legacyPath) {
		return nil
	}
	logger.Info("Migrating legacy agent binary from %s to %s", legacyServiceFile(), t.OldBin)
	if err := os.Rename(legacyPath, filepath.Join(basePath, t.OldBin)); err != nil {
		return fmt.Errorf("error migrating legacy binary: %v", err)
	}
	return nil
}

func migrateLegacyEDR(basePath string) error {
	t := edrTarget()
	if fs.Exists(filepath.Join(basePath, t.OldBin)) {
		return nil
	}
	legacyPath := filepath.Join(basePath, edrBaseName)
	if !fs.Exists(legacyPath) {
		return nil
	}
	logger.Info("Migrating legacy EDR binary from %s to %s", edrBaseName, t.OldBin)
	if err := os.Rename(legacyPath, filepath.Join(basePath, t.OldBin)); err != nil {
		return fmt.Errorf("error migrating legacy binary: %v", err)
	}
	return nil
}

// Version represents one field of the shared version.json file published by
// the server. Every updatable component carries its own field; a reader
// unmarshals only the field it cares about (Go ignores the rest).
type Version struct {
	Version        string `json:"version"`
	UpdaterVersion string `json:"updater_version"`
	EDRVersion     string `json:"edr_version"`
}

// updateHoldState mirrors updateHoldState in agent/agent/updatehold.go —
// the file is the only thing shared between the two, since they are
// separate binaries/modules.
type updateHoldState struct {
	Hold bool `json:"hold"`
}

func updateHeld() bool {
	var state updateHoldState
	if err := fs.ReadJSON(config.UpdateHoldFile, &state); err != nil {
		return false
	}
	return state.Hold
}

func legacyServiceFile() string {
	if runtime.GOOS == "windows" {
		return "utmstack_agent_service.exe"
	}
	return "utmstack_agent_service"
}

var currentVersion = Version{}

// remoteVersions downloads the server-published version.json into
// version_new.json and returns it. The agent's and the EDR's update passes
// share the result so a single poll reads one file.
func remoteVersions(cnf *config.Config, basePath string) (Version, error) {
	if err := http.DownloadFile(
		fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, "version.json"),
		nil, "version_new.json", basePath, cnf.SkipCertValidation,
	); err != nil {
		return Version{}, fmt.Errorf("error downloading version.json: %v", err)
	}
	v := Version{}
	if err := fs.ReadJSON(filepath.Join(basePath, "version_new.json"), &v); err != nil {
		return Version{}, fmt.Errorf("error reading version_new.json: %v", err)
	}
	return v, nil
}

// downloadBinary fetches the named dependency into basePath under newBinName,
// with checksum verification when the server publishes one, and chmod 755 on
// unix.
func downloadBinary(cnf *config.Config, basePath, binName, newBinName string) error {
	verified, err := http.DownloadFileAndVerify(
		fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, binName),
		nil, newBinName, basePath, cnf.SkipCertValidation,
	)
	if err != nil {
		os.Remove(filepath.Join(basePath, newBinName))
		return fmt.Errorf("error downloading or verifying %s: %v", binName, err)
	}
	if verified {
		logger.Info("checksum verified for %s", binName)
	} else {
		logger.Info("no checksum published for %s, installing unverified", binName)
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if err := exec.Run("chmod", basePath, "755", filepath.Join(basePath, newBinName)); err != nil {
			return fmt.Errorf("error executing chmod: %v", err)
		}
	}
	return nil
}

// UpdateDependencies is the main 5-minute loop in the updater service. Each
// iteration it polls the server once for version.json, then checks both the
// agent's and the EDR's version fields against the local one and swaps the
// binaries that changed. The same pause/hold gates apply to both.
func UpdateDependencies(cnf *config.Config) {
	basePath := fs.GetExecutablePath()

	if fs.Exists(config.VersionPath) {
		if err := fs.ReadJSON(config.VersionPath, &currentVersion); err != nil {
			logger.Error("error reading version file: %v", err)
		}
	}

	autoUpdatePausedLogged := false
	remoteHoldLogged := false

	for {
		time.Sleep(checkEvery)

		if cnf.PauseAutoUpdate {
			if !autoUpdatePausedLogged {
				logger.Info("auto-update is paused (pause-auto-update), staying on version %s", currentVersion.Version)
				autoUpdatePausedLogged = true
			}
			continue
		}
		autoUpdatePausedLogged = false

		if updateHeld() {
			if !remoteHoldLogged {
				logger.Info("update held by agent-manager (update_hold config), staying on version %s", currentVersion.Version)
				remoteHoldLogged = true
			}
			continue
		}
		remoteHoldLogged = false

		remote, err := remoteVersions(cnf, basePath)
		if err != nil {
			logger.Error("%v", err)
			continue
		}

		// Snapshot the local version once per iteration: the agent pass may
		// promote version.json mid-loop, and the EDR pass must still compare
		// against what was installed when the poll started.
		local := currentVersion

		if err := agentUpdatePass(cnf, local, remote, basePath); err != nil {
			logger.Error("error updating agent: %v", err)
		}
		if err := edrUpdatePass(cnf, local, remote, basePath); err != nil {
			logger.Error("error updating EDR: %v", err)
		}
	}
}

// agentUpdatePass swaps in the agent binary when the server's "version"
// field differs from the local one. The agent swap always promotes the
// shared version.json (already downloaded into version_new.json) so the new
// binary sees the new version on first boot.
func agentUpdatePass(cnf *config.Config, local, remote Version, basePath string) error {
	if remote.Version == local.Version {
		return nil
	}
	logger.Info("New version of agent found: %s", remote.Version)
	if err := downloadBinary(cnf, basePath, config.ServiceFile(""), config.ServiceFile("_new")); err != nil {
		return err
	}
	if err := runUpdate(basePath, agentTarget(), true); err != nil {
		os.Remove(filepath.Join(basePath, config.ServiceFile("_new")))
		return err
	}
	logger.Info("Agent update completed successfully")
	if err := fs.ReadJSON(config.VersionPath, &currentVersion); err != nil {
		return fmt.Errorf("error reading updated version file: %v", err)
	}
	return nil
}

// edrInstalled reports whether an EDR binary is present in basePath.
func edrInstalled(basePath string) bool {
	return fs.Exists(filepath.Join(basePath, edrServiceFile("")))
}

// edrUpdatePass swaps in the EDR binary when the server's "edr_version"
// field differs from the local one. Endpoints without an EDR binary are
// skipped: installing the module is an explicit operator/manager action
// (enable-edr), not an auto-update. When the agent pass did not promote
// version.json (only the EDR changed), the EDR swap promotes it instead —
// exactly one successful swap per iteration owns the promotion, so the
// version file never claims a state that did not reach disk.
func edrUpdatePass(cnf *config.Config, local, remote Version, basePath string) error {
	if remote.EDRVersion == local.EDRVersion {
		return nil
	}
	if !edrInstalled(basePath) {
		logger.Info("EDR version %s published but EDR not installed, skipping", remote.EDRVersion)
		return nil
	}
	logger.Info("New version of EDR found: %s", remote.EDRVersion)
	if err := downloadBinary(cnf, basePath, edrServiceFile(""), edrServiceFile("_new")); err != nil {
		return err
	}
	agentPromoted := local.Version != remote.Version
	if err := runUpdate(basePath, edrTarget(), !agentPromoted); err != nil {
		os.Remove(filepath.Join(basePath, edrServiceFile("_new")))
		return err
	}
	logger.Info("EDR update completed successfully")
	if err := fs.ReadJSON(config.VersionPath, &currentVersion); err != nil {
		return fmt.Errorf("error reading updated version file: %v", err)
	}
	return nil
}

// runUpdate performs the binary swap for one target: stop, migrate legacy
// naming, backup, rename, start, health check — with rollback on any failure
// after the backup exists. When promoteVersion is true, the shared
// version_new.json (downloaded by remoteVersions) is promoted to version.json
// before the service starts and restored on rollback; only the agent swap
// passes true. State files next to the binary (config, spool, quarantine, DB)
// are never touched: the service restarts in the same directory.
func runUpdate(basePath string, t updateTarget, promoteVersion bool) error {
	newBinPath := filepath.Join(basePath, t.NewBin)
	if _, err := os.Stat(newBinPath); err != nil {
		return fmt.Errorf("no %s found to update", t.NewBin)
	}

	if err := svc.Stop(t.ServiceName); err != nil {
		return fmt.Errorf("error stopping %s: %v", t.ServiceName, err)
	}

	if t.Migrate != nil {
		if err := t.Migrate(basePath); err != nil {
			return err
		}
	}

	backupPath := filepath.Join(basePath, t.BackupBin)
	if fs.Exists(backupPath) {
		logger.Info("Removing previous backup: %s", t.BackupBin)
		if err := os.Remove(backupPath); err != nil {
			logger.Error("could not remove old backup: %v", err)
		}
	}

	if err := os.Rename(filepath.Join(basePath, t.OldBin), backupPath); err != nil {
		return fmt.Errorf("error backing up old binary: %v", err)
	}

	if err := os.Rename(newBinPath, filepath.Join(basePath, t.OldBin)); err != nil {
		os.Rename(backupPath, filepath.Join(basePath, t.OldBin))
		return fmt.Errorf("error renaming new binary: %v", err)
	}

	versionNewPath := filepath.Join(basePath, "version_new.json")
	versionPath := filepath.Join(basePath, "version.json")
	versionBackupPath := filepath.Join(basePath, "version.json.old")

	if promoteVersion {
		os.Remove(versionBackupPath)
		versionBackedUp := false
		if fs.Exists(versionPath) {
			if err := os.Rename(versionPath, versionBackupPath); err != nil {
				os.Rename(filepath.Join(basePath, t.OldBin), newBinPath)
				os.Rename(backupPath, filepath.Join(basePath, t.OldBin))
				return fmt.Errorf("error backing up version.json: %v", err)
			}
			versionBackedUp = true
		}
		if err := os.Rename(versionNewPath, versionPath); err != nil {
			if versionBackedUp {
				os.Rename(versionBackupPath, versionPath)
			}
			os.Rename(filepath.Join(basePath, t.OldBin), newBinPath)
			os.Rename(backupPath, filepath.Join(basePath, t.OldBin))
			return fmt.Errorf("error promoting version.json: %v", err)
		}
	}

	if err := svc.Start(t.ServiceName); err != nil {
		rollback(basePath, t, promoteVersion)
		return fmt.Errorf("error starting %s: %v", t.ServiceName, err)
	}

	time.Sleep(updateHealthWait)

	isHealthy, err := svc.IsActive(t.ServiceName)
	if err != nil || !isHealthy {
		logger.Info("New version of %s failed health check, rolling back...", t.ServiceName)
		rollback(basePath, t, promoteVersion)
		return fmt.Errorf("rollback completed: new version of %s failed health check", t.ServiceName)
	}

	logger.Info("Health check passed for %s", t.ServiceName)
	if promoteVersion {
		os.Remove(versionBackupPath)
	}

	return nil
}

// rollback restores the previous binary for a target and restarts its
// service. It also restores version.json from its backup when the failed
// swap was the one that promoted it. After rollback the endpoint is back on
// the old, known-good version with its state untouched.
func rollback(basePath string, t updateTarget, promoteVersion bool) {
	logger.Info("Rolling back %s to previous version...", t.ServiceName)

	svc.Stop(t.ServiceName)

	os.Remove(filepath.Join(basePath, t.OldBin))
	os.Rename(filepath.Join(basePath, t.BackupBin), filepath.Join(basePath, t.OldBin))

	if promoteVersion {
		versionPath := filepath.Join(basePath, "version.json")
		versionBackupPath := filepath.Join(basePath, "version.json.old")
		if fs.Exists(versionBackupPath) {
			os.Remove(versionPath)
			os.Rename(versionBackupPath, versionPath)
		}
	}

	svc.Start(t.ServiceName)

	logger.Info("Rollback completed for %s", t.ServiceName)
}
