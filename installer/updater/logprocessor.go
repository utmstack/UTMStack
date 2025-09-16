package updater

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/utmstack/UTMStack/installer/config"
)

const (
	zipPrefix  = "utmstack-logs"
	timeLayout = "20060102-150405"
)

func SyncSystemLogs() {
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-rootCtx.Done():
			return
		default:
		}

		active, err := isLogSenderEnabled()
		if err != nil {
			if !IsBackendMaintenanceError(err) {
				config.Logger().ErrorF("Error getting log sender config: %v", err)
			}
		}

		if !config.Updating && active {
			err := CollectAndShipSwarmLogs()
			if err != nil {
				config.Logger().ErrorF("Error collecting and shipping logs: %v", err)
			} else {
				config.Logger().Info("Swarm logs collected and shipped successfully")
			}
		}

		time.Sleep(config.SyncSystemLogsEvery)
	}
}

func disableLogSender() error {
	backConf, err := getConfigFromBackend(9)
	if err != nil {
		if IsBackendMaintenanceError(err) {
			return nil
		}
		return err
	}
	for i, c := range backConf {
		if c.ConfParamShort == "utmstack.intance.send.logs" {
			backConf[i].ConfParamValue = "false"
		}
	}

	err = updateConfigInBackend(backConf, 9)
	if err != nil {
		if IsBackendMaintenanceError(err) {
			return nil
		}
		return fmt.Errorf("error disabling log sender in backend: %v", err)
	}

	return nil
}

func isLogSenderEnabled() (bool, error) {
	backConf, err := getConfigFromBackend(9)
	if err != nil {
		if IsBackendMaintenanceError(err) {
			return false, nil
		}
		return false, err
	}

	return backConf[0].ConfParamValue == "true", nil
}

func CollectAndShipSwarmLogs() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("unable to create Docker client: %v", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("unable to list containers: %v", err)
	}

	if len(containers) == 0 {
		return fmt.Errorf("no containers found")
	}

	archiveName := fmt.Sprintf("%s-%s.zip", zipPrefix, time.Now().Format(timeLayout))
	if err := createZip(ctx, cli, containers, archiveName); err != nil {
		return fmt.Errorf("error creating zip: %v", err)
	}

	if config.ConnectedToInternet {
		if err := GetUpdaterClient().UploadLogs(ctx, archiveName); err != nil {
			return fmt.Errorf("error uploading logs: %v", err)
		}

		_ = os.Remove(archiveName)
	}

	err = disableLogSender()
	if err != nil {
		return fmt.Errorf("error disabling log sender: %v", err)
	}

	return nil
}

func createZip(
	ctx context.Context,
	cli *client.Client,
	containers []container.Summary,
	zipPath string,
) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	for _, c := range containers {
		rc, err := cli.ContainerLogs(ctx, c.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: true,
			Details:    true,
		})
		if err != nil {
			config.Logger().ErrorF("Error getting logs for container %s: %v", c.ID, err)
			continue
		}

		entry, err := zipWriter.Create(fmt.Sprintf("%s.log", sanitize(c.Names[0])))
		if err != nil {
			rc.Close()
			return err
		}
		if _, err = io.Copy(entry, rc); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
	}
	if info, err := os.Stat(config.EventProcessorLogsPath); err == nil && info.IsDir() {
		if err := addDirToZip(zipWriter, config.EventProcessorLogsPath, "events-engine"); err != nil {
			return fmt.Errorf("error adding events-engine logs to zip: %v", err)
		}
	}

	return nil
}

func addDirToZip(zipWriter *zip.Writer, dirPath, baseInZip string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		zipEntry, err := zipWriter.Create(filepath.Join(baseInZip, relPath))
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(zipEntry, file)
		return err
	})
}

func sanitize(name string) string { return strings.TrimPrefix(name, "/") }
