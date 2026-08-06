package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/playground/config"
	"github.com/utmstack/UTMStack/plugins/playground/internal/api"
	runner_pkg "github.com/utmstack/UTMStack/plugins/playground/internal/runner"
)

const playgroundModeEnv = "MODE=playground"

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.playground").Env.Mode
	if mode != "manager" {
		return
	}

	if _, err := os.Stat(config.PlaygroundBinary); err != nil {
		_ = catcher.Error("EP playground binary missing; plugin will not start", err, map[string]any{
			"path": config.PlaygroundBinary,
		})
		return
	}

	if err := runPlaygroundInit(); err != nil {
		_ = catcher.Error("playground -init failed; plugin will not start", err, map[string]any{"process": config.ProcessName})
		return
	}

	go supervisePlaygroundWatch()

	r := runner_pkg.New(config.WorkDir, config.EventTimeout, config.AlertGrace, config.PollInterval)

	if err := api.StartHTTPServer(config.HTTPPort, r, config.InternalKey); err != nil {
		os.Exit(1)
	}
}

func runPlaygroundInit() error {
	cmd := exec.Command(
		config.PlaygroundBinary,
		"-init",
		"-production-workdir", config.ProductionWorkDir,
	)
	cmd.Env = append(os.Environ(), "WORK_DIR="+config.WorkDir, playgroundModeEnv)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func supervisePlaygroundWatch() {
	for {
		cmd := exec.Command(config.PlaygroundBinary, "-watch")
		cmd.Env = append(os.Environ(), "WORK_DIR="+config.WorkDir, playgroundModeEnv)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		_ = catcher.Error("playground -watch exited, restarting", err, map[string]any{
			"process": config.ProcessName,
			"workDir": config.WorkDir,
		})
		time.Sleep(config.WatchRestartDelay)
	}
}
