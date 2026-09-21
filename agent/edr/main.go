package main

import (
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
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
	s := scanner.New(cfg, c, sp, store)
	verdict, sig, err := s.ScanFile(path, event.SourceEngine)
	if err != nil {
		fmt.Println("scan error:", err)
		os.Exit(1)
	}
	fmt.Printf("UTMStack EDR verdict: %s %s\n", verdict, sig)
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
	fmt.Println("UTMStack EDR: restored", id)
}
