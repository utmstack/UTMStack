package main

import (
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			if err := runVersion(os.Args[2:], os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "engine-check":
			runEngineCheck()
			return
		case "install":
			service.InstallService()
			fmt.Println("UTMStackEDR service installed")
			return
		case "uninstall":
			service.UninstallService()
			fmt.Println("UTMStackEDR service uninstalled")
			return
		case "scan":
			if len(os.Args) < 3 {
				fmt.Println("usage: utmstack_edr scan <path>")
				os.Exit(1)
			}
			if strings.HasPrefix(os.Args[2], "--") {
				os.Exit(runDryRun(os.Args[2:], os.Stdout))
			}
			runScan(os.Args[2])
			return
		case "restore":
			if len(os.Args) < 3 {
				fmt.Println("usage: utmstack_edr restore <id>")
				os.Exit(1)
			}
			runRestore(os.Args[2])
			return
		case "config":
			runConfig(os.Args[2:])
			return
		case "network":
			os.Exit(runNetwork(os.Args[2:], os.Stdout))
		case "allow":
			runAllow(os.Args[2:])
			return
		case "quarantine":
			runQuarantine(os.Args[2:])
			return
		case "help", "-h", "--help":
			printUsage()
			return
		case "status":
			if fs.Exists(config.StatusFile) {
				lines, _ := fs.ReadLines(config.StatusFile)
				for _, l := range lines {
					fmt.Println(l)
				}
			} else {
				fmt.Println("UTMStack EDR: not running")
			}
			return
		}
	}
	service.RunService()
}

func runScan(path string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config:", err)
		os.Exit(1)
	}
	c, err := cache.Open(config.DBFile)
	if err != nil {
		fmt.Println("cache:", err)
		os.Exit(1)
	}
	defer c.Close()
	sp, err := event.OpenSpool(config.SpoolFile, 8<<20)
	if err != nil {
		fmt.Println("spool:", err)
		os.Exit(1)
	}
	defer sp.Close()

	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		fmt.Println("quarantine:", err)
		os.Exit(1)
	}
	eng := engine.New(cfg)
	if err := eng.EnsureRunning(); err != nil {
		// Rejected loose files do not make an active signed generation unusable.
		if !errors.Is(err, engine.ErrLocalSignatures) || eng.Version() == "" {
			fmt.Println("UTMStack EDR: signature database unavailable:", err)
			os.Exit(1)
		}
		fmt.Println("UTMStack EDR: warning:", err)
	}
	defer eng.Stop()
	s := scanner.New(cfg, c, sp, store, eng)
	verdict, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil {
		fmt.Println("scan error:", err)
		os.Exit(1)
	}
	fmt.Println(verdictLine(verdict, sig))
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
		fmt.Println("cache:", err)
		os.Exit(1)
	}
	defer c.Close()
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		fmt.Println("quarantine:", err)
		os.Exit(1)
	}
	if err := store.Restore(id); err != nil {
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
	fmt.Println("UTMStack EDR: restored", id)
}
