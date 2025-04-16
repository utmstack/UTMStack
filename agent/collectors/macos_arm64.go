//go:build darwin && arm64
// +build darwin,arm64

package collectors

import (
	"bufio"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/utils"
	"os/exec"
	"path/filepath"
)

type Darwin struct{}

func (d Darwin) Install() error {
	return nil
}

func getCollectorsInstances() []Collector {
	var collectors []Collector
	collectors = append(collectors, Darwin{})
	return collectors
}

func (d Darwin) SendLogs() {
	path := utils.GetMyPath()
	collectorPath := filepath.Join(path, "utmstack-collector-mac")

	cmd := exec.Command(collectorPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = utils.Logger.ErrorF("error creating stdout pipe: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = utils.Logger.ErrorF("error creating stderr pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		_ = utils.Logger.ErrorF("error starting macOS collector: %v", err)
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logLine := scanner.Text()

			utils.Logger.LogF(100, "output: %s", logLine)

			validatedLog, _, err := validations.ValidateString(logLine, false)
			if err != nil {
				utils.Logger.ErrorF("error validating log: %s: %v", logLine, err)
				continue
			}

			logservice.LogQueue <- logservice.LogPipe{
				Src:  string(config.DataTypeMacOSAgent),
				Logs: []string{validatedLog},
			}
		}

		if err := scanner.Err(); err != nil {
			_ = utils.Logger.ErrorF("error reading stdout: %v", err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			errLine := scanner.Text()
			_ = utils.Logger.ErrorF("collector error: %s", errLine)
		}

		if err := scanner.Err(); err != nil {
			_ = utils.Logger.ErrorF("error reading stderr: %v", err)
		}
	}()

	if err := cmd.Wait(); err != nil {
		_ = utils.Logger.ErrorF("macOS collector process ended with error: %v", err)
	}
}

func (d Darwin) Uninstall() error {
	return nil
}
