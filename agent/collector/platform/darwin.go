//go:build darwin
// +build darwin

package platform

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

const (
	maxRestartDelay = 5 * time.Minute
	baseRestartDelay = 5 * time.Second
)

type Darwin struct{}

func GetCollectors() []Collector {
	return []Collector{Darwin{}}
}

func (d Darwin) Name() string {
	return "darwin"
}

func (d Darwin) Start(ctx context.Context, queue chan *plugins.Log) {
	path := fs.GetExecutablePath()
	collectorPath := filepath.Join(path, "utmstack-collector-mac")

	// Verify binary exists before attempting to run
	if _, err := os.Stat(collectorPath); os.IsNotExist(err) {
		utils.Logger.ErrorF("macOS collector binary not found at: %s", collectorPath)
		return
	}

	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

	restartDelay := baseRestartDelay

	// Restart loop with exponential backoff
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("macOS collector stopping due to context cancellation")
			return
		default:
		}

		exitCode := d.runCollector(collectorPath, host, queue)

		if exitCode == 0 {
			utils.Logger.Info("macOS collector exited normally")
		} else {
			utils.Logger.ErrorF("macOS collector exited with code %d, restarting in %v", exitCode, restartDelay)
		}

		time.Sleep(restartDelay)

		// Exponential backoff up to max delay
		restartDelay *= 2
		if restartDelay > maxRestartDelay {
			restartDelay = maxRestartDelay
		}
	}
}

func (d Darwin) runCollector(collectorPath, host string, queue chan *plugins.Log) int {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in macOS collector: %v", r)
		}
	}()

	cmd := exec.Command(collectorPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		utils.Logger.ErrorF("error creating stdout pipe: %v", err)
		return -1
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		utils.Logger.ErrorF("error creating stderr pipe: %v", err)
		return -1
	}

	if err := cmd.Start(); err != nil {
		utils.Logger.ErrorF("error starting macOS collector: %v", err)
		return -1
	}

	utils.Logger.Info("macOS collector started successfully")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in stdout reader: %v", r)
			}
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logLine := scanner.Text()
			utils.Logger.LogF(100, "output: %s", logLine)

			validatedLog, _, err := entities.ValidateString(logLine, false)
			if err != nil {
				utils.Logger.ErrorF("error validating log: %s: %v", logLine, err)
				continue
			}

			queue <- &plugins.Log{
				DataType:   string(config.DataTypeMacOs),
				DataSource: host,
				Raw:        validatedLog,
			}
		}

		if err := scanner.Err(); err != nil {
			utils.Logger.ErrorF("error reading stdout: %v", err)
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in stderr reader: %v", r)
			}
		}()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			errLine := scanner.Text()
			utils.Logger.ErrorF("collector error: %s", errLine)
		}

		if err := scanner.Err(); err != nil {
			utils.Logger.ErrorF("error reading stderr: %v", err)
		}
	}()

	if err := cmd.Wait(); err != nil {
		utils.Logger.ErrorF("macOS collector process ended with error: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return -1
	}

	return 0
}

func (d Darwin) Stop() {}
