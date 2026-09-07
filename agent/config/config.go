package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

type InstallationUUID struct {
	UUID string `yaml:"uuid"`
}

type Config struct {
	Server             string `yaml:"server"`
	AgentID            uint   `yaml:"agent-id"`
	AgentKey           string `yaml:"agent-key"`
	SkipCertValidation bool   `yaml:"insecure"`
	NoRemoteControl    bool   `yaml:"no-remote-control"`
	// PauseAutoUpdate freezes the agent at whatever version it's currently
	// running — the updater checks this and skips applying anything the
	// server offers. Deliberately named/framed as "pause", not "pin to
	// version X": the dependencies endpoint always serves the current
	// binary (config.DependUrl has no version segment), so there is no
	// way to fetch a specific historical version to actually pin to. See
	// C4 in GAPS_AND_IMPROVEMENTS.md.
	PauseAutoUpdate bool `yaml:"pause-auto-update"`
}

var (
	cnf                = Config{}
	confOnce           sync.Once
	installationId     = ""
	installationIdOnce sync.Once
)

func GetCurrentConfig() (*Config, error) {
	var errR error
	confOnce.Do(func() {
		var encryptConfig Config
		if err := fs.ReadYAML(ConfigurationFile, &encryptConfig); err != nil {
			errR = fmt.Errorf("error reading config file: %v", err)
			return
		}

		id, err := GetUUID()
		if err != nil {
			errR = fmt.Errorf("failed to get uuid: %v", err)
			return
		}

		agentKey, err := utils.DecryptAES(encryptConfig.AgentKey, REPLACE_KEY, id)
		if err != nil {
			errR = fmt.Errorf("error decrypting agent key: %v", err)
			return
		}

		cnf.Server = encryptConfig.Server
		cnf.AgentID = encryptConfig.AgentID
		cnf.AgentKey = agentKey
		cnf.SkipCertValidation = encryptConfig.SkipCertValidation
		// NoRemoteControl was previously dropped here (SaveConfig didn't
		// persist it either — see the matching fix there): a customer
		// installing with --no-remote-control got local enforcement only
		// for the lifetime of the install process, and it silently
		// reverted to false on every subsequent service start, since
		// nothing ever wrote or read it back from config.yml.
		cnf.NoRemoteControl = encryptConfig.NoRemoteControl
		cnf.PauseAutoUpdate = encryptConfig.PauseAutoUpdate

		applyEnvOverrides(&cnf)
	})
	if errR != nil {
		return nil, errR
	}
	return &cnf, nil
}

// applyEnvOverrides lets a small set of non-secret settings be overridden
// without editing config.yml — useful for config-managed deployments
// (Ansible, GPO, RMM) that prefer environment variables. AgentID/AgentKey
// are deliberately not overridable this way; they come only from
// registration.
func applyEnvOverrides(cnf *Config) {
	if v, ok := os.LookupEnv("UTMSTACK_SERVER"); ok && v != "" {
		cnf.Server = v
	}
	if v, ok := os.LookupEnv("UTMSTACK_SKIP_CERT_VALIDATION"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cnf.SkipCertValidation = b
		} else {
			utils.Logger.ErrorF("invalid UTMSTACK_SKIP_CERT_VALIDATION value %q, ignoring", v)
		}
	}
	if v, ok := os.LookupEnv("UTMSTACK_NO_REMOTE_CONTROL"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cnf.NoRemoteControl = b
		} else {
			utils.Logger.ErrorF("invalid UTMSTACK_NO_REMOTE_CONTROL value %q, ignoring", v)
		}
	}
	if v, ok := os.LookupEnv("UTMSTACK_PAUSE_AUTO_UPDATE"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cnf.PauseAutoUpdate = b
		} else {
			utils.Logger.ErrorF("invalid UTMSTACK_PAUSE_AUTO_UPDATE value %q, ignoring", v)
		}
	}
}

func SaveConfig(cnf *Config) error {
	id, err := GenerateNewUUID()
	if err != nil {
		return fmt.Errorf("failed to generate uuid: %v", err)
	}

	agentKey, err := utils.EncryptAES(cnf.AgentKey, REPLACE_KEY, id)
	if err != nil {
		return fmt.Errorf("error encrypting agent key: %v", err)
	}

	encryptConf := &Config{
		Server:             cnf.Server,
		AgentID:            cnf.AgentID,
		AgentKey:           agentKey,
		SkipCertValidation: cnf.SkipCertValidation,
		NoRemoteControl:    cnf.NoRemoteControl,
		PauseAutoUpdate:    cnf.PauseAutoUpdate,
	}

	if err := fs.WriteYAML(ConfigurationFile, encryptConf); err != nil {
		return err
	}
	// fs.WriteString (used by WriteYAML) defaults to 0644 — fine for most
	// of its callers, but this file carries the encrypted agent key. Lock
	// it down to owner-only (the agent runs as root/SYSTEM) so a local
	// non-privileged user can't read it. Real ACL tightening on Windows
	// (os.Chmod there only toggles the read-only attribute, it isn't a
	// full ACL change) is a separate, larger piece of work — see B4 in
	// GAPS_AND_IMPROVEMENTS.md.
	if err := os.Chmod(ConfigurationFile, 0o600); err != nil {
		utils.Logger.ErrorF("error restricting permissions on %s: %v", ConfigurationFile, err)
	}
	return nil
}

func GenerateNewUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %v", err)
	}

	InstallationUUID := InstallationUUID{
		UUID: id.String(),
	}

	if err = fs.WriteYAML(UUIDFileName, InstallationUUID); err != nil {
		return "", fmt.Errorf("error writing uuid file: %v", err)
	}
	// Part of the agent-key decryption material (see GetCurrentConfig) —
	// same reasoning as ConfigurationFile above.
	if err := os.Chmod(UUIDFileName, 0o600); err != nil {
		utils.Logger.ErrorF("error restricting permissions on %s: %v", UUIDFileName, err)
	}

	return InstallationUUID.UUID, nil
}

func GetUUID() (string, error) {
	var errR error
	installationIdOnce.Do(func() {
		var id = InstallationUUID{}
		if err := fs.ReadYAML(UUIDFileName, &id); err != nil {
			errR = fmt.Errorf("error reading uuid file: %v", err)
			return
		}

		installationId = id.UUID
	})

	if errR != nil {
		return "", errR
	}

	return installationId, nil
}
