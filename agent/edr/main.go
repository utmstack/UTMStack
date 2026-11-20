package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/scanner"
	"github.com/utmstack/UTMStack/agent/edr/service"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
)

func main() {
	if runInternalBroker() {
		return
	}
	// Inspection must not create a log, database or recovery lock before
	// collecting the operator's evidence.
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "inspect" {
		os.Exit(runQuarantineInspection(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "inspect-fresh" {
		os.Exit(runQuarantineFreshInspection(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "approve-fresh" {
		os.Exit(runQuarantineFreshApproval(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "execute-fresh" {
		os.Exit(runQuarantineFreshExecution(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "inspect-fresh-cancel" {
		os.Exit(runQuarantineFreshCancellationInspection(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "cancel-fresh" {
		os.Exit(runQuarantineFreshCancellation(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "inspect-retirement" {
		os.Exit(runQuarantineRetirementInspection(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "retire-cancelled" {
		os.Exit(runQuarantineRetirement(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "hold" {
		os.Exit(runQuarantineHold(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "release-hold" {
		os.Exit(runQuarantineRelease(os.Args[3:], os.Stdout, os.Stderr))
	}
	if len(os.Args) >= 3 && os.Args[1] == "quarantine" && os.Args[2] == "cancel-restore" {
		os.Exit(runQuarantineCancel(os.Args[3:], os.Stdout, os.Stderr))
	}
	limit := configureGoMemoryLimit()
	logger.Init(config.LogFile, logger.LevelInfo)
	if limit > 0 {
		logger.Info("UTMStack EDR: cgroup Go memory target %d bytes", limit)
	}

	// Strip --json from all positions so verb handlers see clean args; the
	// flag can appear anywhere (the agent appends it at the end).
	args := os.Args[1:]
	jsonMode, args = parseJSONFlag(args)

	if len(args) > 0 {
		switch args[0] {
		case "version":
			if err := runVersion(args[1:], os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "engine-check":
			runEngineCheck()
			return
		case "install":
			service.InstallService()
			if jsonMode {
				emitJSON(true, "", map[string]interface{}{"action": "install"})
			} else {
				fmt.Println("UTMStackEDR service installed")
			}
			return
		case "uninstall":
			service.UninstallService()
			if jsonMode {
				emitJSON(true, "", map[string]interface{}{"action": "uninstall"})
			} else {
				fmt.Println("UTMStackEDR service uninstalled")
			}
			return
		case "scan":
			if len(args) < 2 {
				if jsonMode {
					failJSON("usage: utmstack_edr scan <path>")
				}
				fmt.Println("usage: utmstack_edr scan <path>")
				os.Exit(1)
			}
			if strings.HasPrefix(args[1], "--") {
				os.Exit(runDryRun(args[1:], os.Stdout))
			}
			runScan(args[1])
			return
		case "restore":
			if len(args) < 2 {
				if jsonMode {
					failJSON("usage: utmstack_edr restore <id>")
				}
				fmt.Println("usage: utmstack_edr restore <id>")
				os.Exit(1)
			}
			runRestore(args[1])
			return
		case "config":
			runConfig(args[1:])
			return
		case "network":
			os.Exit(runNetwork(args[1:], os.Stdout))
		case "allow":
			runAllow(args[1:])
			return
		case "quarantine":
			runQuarantine(args[1:])
			return
		case "help", "-h", "--help":
			if jsonMode {
				emitJSON(true, "", map[string]interface{}{"usage": usageText()})
			} else {
				printUsage()
			}
			return
		case "status":
			runStatus()
			return
		}
	}
	service.RunService()
}

func runScan(path string) {
	cfg, err := config.Load()
	if err != nil {
		if jsonMode {
			failJSON("config: " + err.Error())
		}
		fmt.Println("config:", err)
		os.Exit(1)
	}
	c, err := cache.Open(config.DBFile)
	if err != nil {
		if jsonMode {
			failJSON("cache: " + err.Error())
		}
		fmt.Println("cache:", err)
		os.Exit(1)
	}
	defer c.Close()
	sp, err := event.OpenSpool(config.SpoolFile, 8<<20)
	if err != nil {
		if jsonMode {
			failJSON("spool: " + err.Error())
		}
		fmt.Println("spool:", err)
		os.Exit(1)
	}
	defer sp.Close()

	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		if jsonMode {
			failJSON("quarantine: " + err.Error())
		}
		fmt.Println("quarantine:", err)
		os.Exit(1)
	}
	eng := engine.New(cfg)
	if err := eng.EnsureRunning(); err != nil {
		// Rejected loose files do not make an active signed generation unusable.
		if !errors.Is(err, engine.ErrLocalSignatures) || eng.Version() == "" {
			if jsonMode {
				failJSON("signature database unavailable: " + err.Error())
			}
			fmt.Println("UTMStack EDR: signature database unavailable:", err)
			os.Exit(1)
		}
		if jsonMode {
			failJSON("warning: " + err.Error())
		}
		fmt.Println("UTMStack EDR: warning:", err)
	}
	defer eng.Stop()
	s := scanner.New(cfg, c, sp, store, eng)
	verdict, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil {
		if jsonMode {
			failJSON("scan error: " + err.Error())
		}
		fmt.Println("scan error:", err)
		os.Exit(1)
	}
	if jsonMode {
		emitJSON(true, "", map[string]interface{}{"verdict": verdict, "signature": sig})
	} else {
		fmt.Println(verdictLine(verdict, sig))
	}
}

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

func usageText() string {
	var b strings.Builder
	printUsageWriter(&b)
	return b.String()
}

// verdictLine is the on-demand scan's output. Detection names pass the
// branding filter, as in events.
func verdictLine(verdict, sig string) string {
	return fmt.Sprintf("UTMStack EDR verdict: %s %s", verdict, event.SanitizeName(sig))
}

func runRestore(id string) {
	cfg, _ := config.Load()
	c, err := cache.Open(config.DBFile)
	if err != nil {
		if jsonMode {
			failJSON("cache: " + err.Error())
		}
		fmt.Println("cache:", err)
		os.Exit(1)
	}
	defer c.Close()
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		if jsonMode {
			failJSON("quarantine: " + err.Error())
		}
		fmt.Println("quarantine:", err)
		os.Exit(1)
	}
	if err := store.Restore(id); err != nil {
		if jsonMode {
			failJSON("restore error: " + err.Error())
		}
		fmt.Println("restore error:", err)
		os.Exit(1)
	}
	rec, _, _ := c.GetQuarantine(id)
	if sp, spErr := event.OpenSpool(config.SpoolFile, 8<<20); spErr == nil {
		defer sp.Close()
		if js, jErr := event.NewQuarantineRestoreEvent(rec.OriginalPath, rec.SHA256, id).ToJSON(); jErr == nil {
			_ = sp.Append(js)
		}
	}
	if jsonMode {
		emitJSON(true, "", map[string]interface{}{"restored": id})
	} else {
		fmt.Println("UTMStack EDR: restored", id)
	}
}
