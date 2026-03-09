package collector

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/as400/config"
	"github.com/utmstack/UTMStack/as400/logservice"
	"github.com/utmstack/UTMStack/as400/utils"
)

type AS400Collector struct {
	configStreamManager *ConfigStreamManager
	collectorJarPath    string
	configFilePath      string
	currentProcess      *exec.Cmd
	processMutex        sync.Mutex
	isRunning           bool
	ctx                 context.Context
	cancel              context.CancelFunc
	hostname            string
}

func NewAS400Collector() *AS400Collector {
	hostname, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		hostname = "unknown"
	}

	collector := &AS400Collector{
		collectorJarPath: config.CollectorJarPath,
		configFilePath:   config.ConfigFilePath,
		hostname:         hostname,
	}

	collector.configStreamManager = NewConfigStreamManager(collector.handleConfigurationChange)

	return collector
}

func (c *AS400Collector) Start(ctx context.Context, cnf *config.Config) error {
	utils.Logger.Info("Starting AS400 Collector...")

	c.ctx, c.cancel = context.WithCancel(ctx)

	if !utils.CheckIfPathExist(c.collectorJarPath) {
		return utils.Logger.ErrorF("AS400 collector JAR not found at: %s", c.collectorJarPath)
	}

	if utils.CheckIfPathExist(c.configFilePath) {
		utils.Logger.Info("Found existing configuration file, starting JAR...")
		if err := c.startCollectorProcess(); err != nil {
			utils.Logger.ErrorF("Error starting JAR with existing config: %v", err)
		}
	}

	go c.configStreamManager.Start(cnf, c.ctx)

	utils.Logger.Info("AS400 started, waiting for configuration...")

	return nil
}

func (c *AS400Collector) Stop() error {
	utils.Logger.Info("Stopping AS400 Collector...")

	if c.cancel != nil {
		c.cancel()
	}

	if err := c.stopCollectorProcess(); err != nil {
		utils.Logger.ErrorF("Error stopping collector process: %v", err)
	}

	utils.Logger.Info("AS400 Collector stopped")
	return nil
}

func (c *AS400Collector) handleConfigurationChange(newConfig *AS400CollectorConfig) {
	if err := c.saveConfig(newConfig); err != nil {
		utils.Logger.ErrorF("Error saving configuration: %v", err)
		return
	}

	if !c.isRunning {
		if err := c.startCollectorProcess(); err != nil {
			utils.Logger.ErrorF("Error starting JAR: %v", err)
		}
	}
}

func (c *AS400Collector) saveConfig(config *AS400CollectorConfig) error {
	if err := EncryptPasswords(config); err != nil {
		return utils.Logger.ErrorF("error encrypting passwords: %v", err)
	}

	if err := utils.WriteJSON(c.configFilePath, config); err != nil {
		return utils.Logger.ErrorF("error writing config file: %v", err)
	}

	utils.Logger.Info("Configuration saved: %d servers", len(config.Servers))
	return nil
}

func (c *AS400Collector) startCollectorProcess() error {
	c.processMutex.Lock()
	defer c.processMutex.Unlock()

	if c.isRunning {
		return nil
	}

	utils.Logger.Info("Starting AS400 collector JAR...")

	cmd := exec.CommandContext(c.ctx, "java", "-jar", c.collectorJarPath, "RUN")
	cmd.Dir = utils.GetMyPath()
	cmd.Env = append(os.Environ(), "AS400_SECRET="+config.REPLACE_KEY)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return utils.Logger.ErrorF("error creating stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return utils.Logger.ErrorF("error creating stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return utils.Logger.ErrorF("error starting JAR: %v", err)
	}

	c.currentProcess = cmd
	c.isRunning = true

	utils.Logger.Info("JAR started (PID: %d)", cmd.Process.Pid)

	go c.processCollectorLogs(stdout)
	go c.processCollectorErrors(stderr)
	go c.monitorProcess(cmd)

	return nil
}

func (c *AS400Collector) processCollectorLogs(stdout io.ReadCloser) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		logLine := scanner.Text()

		validatedLog, _, err := entities.ValidateString(logLine, false)
		if err != nil {
			utils.Logger.ErrorF("invalid log: %v", err)
			continue
		}

		logservice.LogQueue <- &plugins.Log{
			DataType:   string(config.DataType),
			DataSource: c.hostname,
			Raw:        validatedLog,
		}
	}

	if err := scanner.Err(); err != nil {
		utils.Logger.ErrorF("error reading stdout: %v", err)
	}
}

func (c *AS400Collector) processCollectorErrors(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		utils.Logger.ErrorF("JAR error: %s", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		utils.Logger.ErrorF("error reading stderr: %v", err)
	}
}

func (c *AS400Collector) stopCollectorProcess() error {
	c.processMutex.Lock()
	defer c.processMutex.Unlock()

	if !c.isRunning || c.currentProcess == nil {
		return nil
	}

	utils.Logger.Info("Stopping JAR (PID: %d)...", c.currentProcess.Process.Pid)

	if err := c.currentProcess.Process.Signal(syscall.SIGTERM); err != nil {
		utils.Logger.ErrorF("error sending SIGTERM: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- c.currentProcess.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		utils.Logger.Info("Forcing SIGKILL...")
		c.currentProcess.Process.Kill()
		<-done
	case <-done:
	}

	c.isRunning = false
	c.currentProcess = nil

	utils.Logger.Info("JAR stopped")
	return nil
}

func (c *AS400Collector) monitorProcess(cmd *exec.Cmd) {
	err := cmd.Wait()

	c.processMutex.Lock()
	c.isRunning = false
	c.currentProcess = nil
	c.processMutex.Unlock()

	if err != nil {
		utils.Logger.ErrorF("JAR exited with error: %v", err)
	} else {
		utils.Logger.Info("JAR exited")
	}
}
