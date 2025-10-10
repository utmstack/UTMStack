package collector

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/utmstack-collector/config"
	"github.com/utmstack/UTMStack/utmstack-collector/logservice"
	"github.com/utmstack/UTMStack/utmstack-collector/models"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
)

type DockerCollector struct {
	client      *client.Client
	filter      *models.ContainerFilter
	ctx         context.Context
	cancel      context.CancelFunc
	containerMu sync.RWMutex
	containers  map[string]models.Container
	eventChan   chan models.ContainerEvent
	stopOnce    sync.Once
}

type Config struct {
	DockerHost   string              `yaml:"docker_host"`
	APIVersion   string              `yaml:"api_version"`
	FilterRules  []models.FilterRule `yaml:"filter_rules"`
	MaxLogLength int                 `yaml:"max_log_length"`
	BufferSize   int                 `yaml:"buffer_size"`
}

func DefaultConfig() *Config {
	return &Config{
		DockerHost:   config.DockerHost,
		APIVersion:   config.DockerAPIVersion,
		MaxLogLength: config.MaxLogLength,
		BufferSize:   config.BufferSize,
		FilterRules:  config.FilterRules,
	}
}

func NewDockerCollector(config *Config) (*DockerCollector, error) {
	var cli *client.Client
	var err error

	if config.DockerHost != "" {
		cli, err = client.NewClientWithOpts(
			client.WithHost(config.DockerHost),
			client.WithAPIVersionNegotiation(),
		)
	} else {
		cli, err = client.NewClientWithOpts(client.FromEnv)
	}

	if err != nil {
		return nil, utils.Logger.ErrorF("failed to create Docker client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, utils.Logger.ErrorF("failed to connect to Docker: %v", err)
	}

	utils.Logger.Info("Connected to Docker daemon successfully")

	ctx, cancel = context.WithCancel(context.Background())

	return &DockerCollector{
		client:     cli,
		filter:     models.NewContainerFilter(config.FilterRules),
		ctx:        ctx,
		cancel:     cancel,
		containers: make(map[string]models.Container),
		eventChan:  make(chan models.ContainerEvent, config.BufferSize),
	}, nil
}

func (d *DockerCollector) Start() error {
	utils.Logger.Info("Starting Docker log collector")

	if err := d.discoverContainers(); err != nil {
		return utils.Logger.ErrorF("failed to discover containers: %v", err)
	}

	go d.monitorEvents()

	go d.periodicRediscovery()

	d.containerMu.RLock()
	for _, container := range d.containers {
		if d.filter.ShouldCollect(container) {
			go d.streamContainerLogs(container)
		}
	}
	d.containerMu.RUnlock()

	utils.Logger.Info("Docker log collector started successfully")
	return nil
}

func (d *DockerCollector) Stop() {
	d.stopOnce.Do(func() {
		utils.Logger.Info("Stopping Docker log collector")
		d.cancel()

		close(d.eventChan)

		if d.client != nil {
			d.client.Close()
		}
	})
}

func (d *DockerCollector) discoverContainers() error {
	containers, err := d.client.ContainerList(d.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}

	d.containerMu.Lock()
	defer d.containerMu.Unlock()

	for _, c := range containers {
		container := d.convertContainer(c)
		d.containers[container.ID] = container
	}

	utils.Logger.Info("Discovered %d containers", len(containers))
	return nil
}

func (d *DockerCollector) monitorEvents() {
	utils.Logger.Info("Starting Docker events monitoring")

	eventChan, errChan := d.client.Events(d.ctx, types.EventsOptions{})

	for {
		select {
		case <-d.ctx.Done():
			return

		case event := <-eventChan:
			d.handleDockerEvent(event)

		case err := <-errChan:
			if err != nil && err != io.EOF {
				utils.Logger.ErrorF("Error monitoring Docker events: %v", err)
			}
		}
	}
}

func (d *DockerCollector) handleDockerEvent(event events.Message) {
	if event.Type != events.ContainerEventType {
		return
	}

	containerEvent := models.ContainerEvent{
		ID:          uuid.New().String(),
		ContainerID: event.Actor.ID,
		Action:      event.Action,
		Timestamp:   time.Unix(event.Time, 0),
		Attributes:  event.Actor.Attributes,
	}

	select {
	case d.eventChan <- containerEvent:
	case <-d.ctx.Done():
		return
	default:
		utils.Logger.Info("Event channel is full, dropping event")
	}

	switch event.Action {
	case "start":
		d.handleContainerStart(event.Actor.ID)
	case "stop", "die", "kill":
		d.handleContainerStop(event.Actor.ID)
	case "destroy":
		d.handleContainerDestroy(event.Actor.ID)
	}
}

func (d *DockerCollector) handleContainerStart(containerID string) {
	containerJSON, err := d.client.ContainerInspect(d.ctx, containerID)
	if err != nil {
		utils.Logger.ErrorF("Failed to inspect container %s: %v", containerID, err)
		return
	}

	container := d.convertContainerJSON(containerJSON)

	d.containerMu.Lock()
	d.containers[containerID] = container
	d.containerMu.Unlock()

	if d.filter.ShouldCollect(container) {
		go d.streamContainerLogs(container)
	}
}

func (d *DockerCollector) handleContainerStop(containerID string) {
	d.containerMu.Lock()
	if container, exists := d.containers[containerID]; exists {
		utils.Logger.Info("Container %s stopped, removing from cache", container.Name)
		delete(d.containers, containerID)
	}
	d.containerMu.Unlock()
}

func (d *DockerCollector) handleContainerDestroy(containerID string) {
	d.containerMu.Lock()
	delete(d.containers, containerID)
	d.containerMu.Unlock()
}

func (d *DockerCollector) streamContainerLogs(container models.Container) {
	utils.Logger.Info("Starting log stream for container: %s", container.Name)

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "0",
		Since:      time.Now().Format(time.RFC3339),
		Timestamps: false,
	}

	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		reader, err := d.client.ContainerLogs(d.ctx, container.ID, options)
		if err != nil {
			utils.Logger.ErrorF("Failed to get logs for container %s (attempt %d/%d): %v",
				container.Name, retry+1, maxRetries, err)
			if retry < maxRetries-1 {
				time.Sleep(time.Duration(retry+1) * time.Second)
				continue
			}
			return
		}

		d.processLogs(reader, container)
		reader.Close()
		return
	}
}

func (d *DockerCollector) processLogs(reader io.ReadCloser, container models.Container) {
	scanner := bufio.NewScanner(reader)
	logCount := 0

	utils.Logger.Info("Processing logs for container: %s", container.Name)

	for scanner.Scan() {
		select {
		case <-d.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if err := d.processAndSendLogLine(line, container); err == nil {
			logCount++
		}
	}

	utils.Logger.Info("Log stream ended for container %s. Processed %d logs", container.Name, logCount)

	if err := scanner.Err(); err != nil && err != io.EOF {
		utils.Logger.ErrorF("Log stream error for container %s: %v", container.Name, err)
		go d.retryLogStream(container)
	}
}

func (d *DockerCollector) retryLogStream(container models.Container) {
	select {
	case <-d.ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}

	containerJSON, err := d.client.ContainerInspect(d.ctx, container.ID)
	if err == nil && containerJSON.State.Running {
		utils.Logger.Info("Retrying log stream for container: %s", container.Name)
		go d.streamContainerLogs(container)
		return
	}

	utils.Logger.Info("Container %s with ID %s not found, searching by name for Swarm replacement", container.Name, container.ID[:12])

	containers, err := d.client.ContainerList(d.ctx, types.ContainerListOptions{All: false})
	if err != nil {
		utils.Logger.ErrorF("Failed to list containers while searching for %s: %v", container.Name, err)
		return
	}

	for _, c := range containers {
		containerName := ""
		if len(c.Names) > 0 {
			containerName = strings.TrimPrefix(c.Names[0], "/")
		}

		if containerName == container.Name && c.State == "running" {
			utils.Logger.Info("Found replacement container %s with new ID %s", container.Name, c.ID[:12])
			newContainer := d.convertContainer(c)

			d.containerMu.Lock()
			delete(d.containers, container.ID)           // Remove old id
			d.containers[newContainer.ID] = newContainer // Add new Id
			d.containerMu.Unlock()

			if d.filter.ShouldCollect(newContainer) {
				go d.streamContainerLogs(newContainer)
			}
			return
		}
	}

	utils.Logger.Info("Container %s no longer running and no replacement found, stopping retry", container.Name)
}

func (d *DockerCollector) processAndSendLogLine(line string, container models.Container) error {
	if len(line) > 8 {
		line = line[8:]
	}

	cleaned, err := models.ParseLogLine(line, config.MaxLogLength)
	if err != nil {
		return err
	}

	enrichedLog := models.EnrichLogWithContainer(cleaned, container.Name)

	if logservice.LogQueue != nil {
		osInfo, err := utils.GetOsInfo()
		if err != nil {
			utils.Logger.ErrorF("Failed to get OS info: %v", err)
			return err
		}

		utmLog := &plugins.Log{
			Raw:        enrichedLog,
			DataType:   config.DataType,
			DataSource: osInfo.Hostname,
		}

		d.sendToUTMStack(utmLog)
	}

	return nil
}

func (d *DockerCollector) sendToUTMStack(utmLog *plugins.Log) {
	select {
	case logservice.LogQueue <- utmLog:
	case <-d.ctx.Done():
		return
	default:
		select {
		case <-time.After(100 * time.Millisecond):
			select {
			case logservice.LogQueue <- utmLog:
			default:
				// Drop log if queue is still full
			}
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *DockerCollector) convertContainer(c types.Container) models.Container {
	name := ""
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}

	return models.Container{
		ID:      c.ID,
		Name:    name,
		Image:   c.Image,
		Status:  c.Status,
		State:   c.State,
		Created: time.Unix(c.Created, 0),
		Labels:  c.Labels,
	}
}

func (d *DockerCollector) convertContainerJSON(c types.ContainerJSON) models.Container {
	var created time.Time
	if c.Created != "" {
		if parsedTime, err := time.Parse(time.RFC3339Nano, c.Created); err == nil {
			created = parsedTime
		}
	}

	return models.Container{
		ID:      c.ID,
		Name:    strings.TrimPrefix(c.Name, "/"),
		Image:   c.Config.Image,
		Status:  c.State.Status,
		State:   c.State.Status,
		Created: created,
		Labels:  c.Config.Labels,
	}
}

func (d *DockerCollector) periodicRediscovery() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.rediscoverContainers(); err != nil {
				utils.Logger.ErrorF("Error during rediscovery: %v", err)
			}
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *DockerCollector) rediscoverContainers() error {
	containers, err := d.client.ContainerList(d.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return utils.Logger.ErrorF("failed to list containers: %v", err)
	}

	d.containerMu.Lock()
	defer d.containerMu.Unlock()

	newContainers := 0
	for _, c := range containers {
		if _, exists := d.containers[c.ID]; !exists {
			container := d.convertContainer(c)
			d.containers[container.ID] = container
			newContainers++

			if d.filter.ShouldCollect(container) {
				go d.streamContainerLogs(container)
			}
		}
	}

	if newContainers > 0 {
		utils.Logger.Info("Rediscovered %d new containers", newContainers)
	}

	return nil
}
