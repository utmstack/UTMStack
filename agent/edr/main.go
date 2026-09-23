package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/scanner"
	"github.com/utmstack/UTMStack/agent/edr/service"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
)

func main() {
	logger.Init(config.LogFile, logger.LevelInfo)

	// --json may appear anywhere; strip it so verb args stay clean, then run
	// the verb with jsonMode set for the rest of the process.
	args := os.Args[1:]
	jsonMode, args = parseJSONFlag(args)

	if len(args) > 0 {
		switch args[0] {
		case "install":
			runInstall()
			return
		case "uninstall":
			runUninstall()
			return
		case "scan":
			if len(args) < 2 {
				failUsage("usage: utmstack_edr scan <path>")
			}
			runScan(args[1])
			return
		case "restore":
			if len(args) < 2 {
				failUsage("usage: utmstack_edr restore <id>")
			}
			runRestore(args[1])
			return
		case "config":
			runConfig(args[1:])
			return
		case "allow":
			runAllow(args[1:])
			return
		case "quarantine":
			runQuarantine(args[1:])
			return
		case "help", "-h", "--help":
			runHelp()
			return
		case "status":
			runStatus()
			return
		}
	}
	service.RunService()
}

func runInstall() {
	service.InstallService()
	if jsonMode {
		data, _ := installData()
		emitJSON(true, "", data)
		return
	}
	fmt.Println("UTMStack EDR service installed")
}

func runUninstall() {
	service.UninstallService()
	if jsonMode {
		data, _ := uninstallData()
		emitJSON(true, "", data)
		return
	}
	fmt.Println("UTMStack EDR service uninstalled")
}

// runStatus shows the running service status: the human path prints the raw
// status file line-by-line, the --json path parses it into the envelope.
func runStatus() {
	if jsonMode {
		data, err := statusData()
		if err != nil {
			failJSON(err.Error())
		}
		emitJSON(true, "", data)
		return
	}
	if fs.Exists(config.StatusFile) {
		lines, _ := fs.ReadLines(config.StatusFile)
		for _, l := range lines {
			fmt.Println(l)
		}
	} else {
		fmt.Println("UTMStack EDR: not running")
	}
}

// statusData is the --json core of `status`: it parses status.json (written by
// the service with json.MarshalIndent) into a plain object. A missing file
// means the service is not running.
func statusData() (interface{}, error) {
	if !fs.Exists(config.StatusFile) {
		return map[string]interface{}{"running": false}, nil
	}
	b, err := os.ReadFile(config.StatusFile)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	return m, nil
}

// installData / uninstallData are the --json data shapes for the lifecycle
// verbs. The service registration itself happens in the entry points because
// the service package owns that process lifecycle (it exits on failure).
func installData() (interface{}, error) {
	return map[string]interface{}{"action": "install"}, nil
}

func uninstallData() (interface{}, error) {
	return map[string]interface{}{"action": "uninstall"}, nil
}

// helpData is the --json core of `help`: the exact usage text.
func helpData() (interface{}, error) {
	return map[string]interface{}{"usage": usageText()}, nil
}

func runHelp() {
	if jsonMode {
		data, _ := helpData()
		emitJSON(true, "", data)
		return
	}
	printUsage()
}

// failUsage prints the usage line (human) or an error envelope (--json) and
// exits non-zero. It is shared by the scan/restore argument guards so both
// modes fail identically.
func failUsage(msg string) {
	if jsonMode {
		failJSON(msg)
	}
	fmt.Println(msg)
	os.Exit(1)
}

func runScan(path string) {
	verdict, sig, err := scanData(path)
	if err != nil {
		if jsonMode {
			failJSON(err.Error())
		}
		fmt.Println(err)
		os.Exit(1)
	}
	if jsonMode {
		emitJSON(true, "", map[string]interface{}{"verdict": verdict, "signature": sig})
		return
	}
	fmt.Printf("UTMStack EDR verdict: %s %s\n", verdict, sig)
}

// scanData is the --json core of `scan <path>`: it wires the cache, spool,
// quarantine store and scanner and runs the scan, returning the verdict and
// signature without printing. Each failure keeps the human-facing prefix so
// the non-JSON path is byte-identical to the original inline version.
func scanData(path string) (string, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", fmt.Errorf("config: %w", err)
	}
	c, err := cache.Open(config.DBFile)
	if err != nil {
		return "", "", fmt.Errorf("cache: %w", err)
	}
	defer c.Close()
	sp, err := event.OpenSpool(config.SpoolFile, 8<<20)
	if err != nil {
		return "", "", fmt.Errorf("spool: %w", err)
	}
	defer sp.Close()

	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		return "", "", fmt.Errorf("quarantine: %w", err)
	}
	s := scanner.New(cfg, c, sp, store)
	verdict, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil {
		return "", "", fmt.Errorf("scan error: %w", err)
	}
	return verdict, sig, nil
}

func runRestore(id string) {
	if err := restoreData(id); err != nil {
		if jsonMode {
			failJSON(err.Error())
		}
		fmt.Println(err)
		os.Exit(1)
	}
	if jsonMode {
		emitJSON(true, "", map[string]interface{}{"restored": id})
		return
	}
	fmt.Println("UTMStack EDR: restored", id)
}

// restoreData is the --json core of `restore <id>`: it wires the quarantine
// store and restores the item without printing; each failure keeps the
// human-facing prefix so the non-JSON path is byte-identical to the original
// inline version.
func restoreData(id string) error {
	cfg, _ := config.Load()
	c, err := cache.Open(config.DBFile)
	if err != nil {
		return fmt.Errorf("cache: %w", err)
	}
	defer c.Close()
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		return fmt.Errorf("quarantine: %w", err)
	}
	if err := store.Restore(id); err != nil {
		return fmt.Errorf("restore error: %w", err)
	}
	return nil
}
