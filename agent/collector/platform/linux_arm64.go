//go:build linux && arm64
// +build linux,arm64

package platform

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

const (
	journaldRestartDelayArm64    = 5 * time.Second
	journaldMaxRestartDelayArm64 = 5 * time.Minute
)

type LinuxSystemArm64 struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	mu     sync.Mutex
}

func (l *LinuxSystemArm64) Name() string {
	return "linux-system"
}

func (l *LinuxSystemArm64) Start(ctx context.Context, queue chan *plugins.Log) {
	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

	restartDelay := journaldRestartDelayArm64

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("Linux system collector stopping due to context cancellation")
			return
		default:
		}

		exitCode := l.runJournalctl(ctx, host, queue)

		if exitCode == 0 {
			utils.Logger.Info("journalctl exited normally")
		} else {
			utils.Logger.ErrorF("journalctl exited with code %d, restarting in %v", exitCode, restartDelay)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(restartDelay):
		}

		// Exponential backoff
		restartDelay *= 2
		if restartDelay > journaldMaxRestartDelayArm64 {
			restartDelay = journaldMaxRestartDelayArm64
		}
	}
}

func (l *LinuxSystemArm64) runJournalctl(ctx context.Context, host string, queue chan *plugins.Log) int {
	l.mu.Lock()
	cmdCtx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	// journalctl -f: follow, -o json: JSON output, --no-pager: don't use pager
	l.cmd = exec.CommandContext(cmdCtx, "journalctl", "-f", "-o", "json", "--no-pager")

	stdout, err := l.cmd.StdoutPipe()
	if err != nil {
		l.mu.Unlock()
		utils.Logger.ErrorF("error creating journalctl stdout pipe: %v", err)
		return -1
	}

	stderr, err := l.cmd.StderrPipe()
	if err != nil {
		l.mu.Unlock()
		utils.Logger.ErrorF("error creating journalctl stderr pipe: %v", err)
		return -1
	}

	if err := l.cmd.Start(); err != nil {
		l.mu.Unlock()
		utils.Logger.ErrorF("error starting journalctl: %v", err)
		return -1
	}
	l.mu.Unlock()

	utils.Logger.Info("Linux system log collector started (journald)")

	// Read stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			utils.Logger.ErrorF("journalctl error: %s", scanner.Text())
		}
	}()

	// Read stdout (log entries)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		validatedLog, _, err := entities.ValidateString(line, false)
		if err != nil {
			utils.Logger.ErrorF("error validating log: %v", err)
			continue
		}

		queue <- &plugins.Log{
			DataType:   string(config.DataTypeLinuxAgent),
			DataSource: host,
			Raw:        validatedLog,
		}
	}

	if err := scanner.Err(); err != nil {
		utils.Logger.ErrorF("error reading journalctl output: %v", err)
	}

	if err := l.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return -1
	}

	return 0
}

func (l *LinuxSystemArm64) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cancel != nil {
		l.cancel()
	}
	if l.cmd != nil && l.cmd.Process != nil {
		l.cmd.Process.Kill()
	}
}

func GetCollectors() []Collector {
	return []Collector{
		&LinuxSystemArm64{},
	}
}
