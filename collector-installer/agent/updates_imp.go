package agent

import (
	context "context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/models"
	"github.com/utmstack/UTMStack/collector-installer/utils"
	"google.golang.org/grpc/metadata"
)

func GetCollectorType(collectorType config.Collector) CollectorType {
	switch collectorType {
	case config.AS400:
		return CollectorType_AS_400_COLLECTOR
	}
	return CollectorType_AS_400_COLLECTOR
}

func DownloadDependencies(address, authKey string, collectorType config.Collector, h *utils.BeautyLogger) error {
	collect := GetCollectorType(collectorType)

	conn, err := ConnectToServer(h, address, config.AgentManagerPort)
	if err != nil {
		return fmt.Errorf("error connecting to server to download dependencies: %v", err)
	}
	defer conn.Close()

	dependClient := NewUpdatesServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", authKey)

	Version := models.Version{}

	resp, err := dependClient.CheckCollectorUpdates(ctx, &UpdateCollectorRequest{
		CollectorType:  collect,
		DependType:     DependType_DEPEND_TYPE_SERVICE,
		Os:             getOS(),
		CurrentVersion: "0",
	})
	if err != nil {
		return fmt.Errorf("error checking collector service: %v", err)
	}
	newVersion, _, err := processUpdateResponse(resp, config.GetDownloadFilePath("service", config.AS400), true)
	if err != nil {
		return fmt.Errorf("error processing collector service: %v", err)
	}
	Version.ServiceVersion = newVersion

	resp, err = dependClient.CheckCollectorUpdates(ctx, &UpdateCollectorRequest{
		CollectorType:  collect,
		DependType:     DependType_DEPEND_TYPE_DEPEND_ZIP,
		Os:             getOS(),
		CurrentVersion: "0",
	})
	if err != nil {
		return fmt.Errorf("error checking collector dependencies: %v", err)
	}

	downloadDepend := config.CheckIfNecessaryDownloadDependencies(collectorType)
	newVersion, _, err = processUpdateResponse(resp, config.GetDownloadFilePath("dependencies", config.AS400), downloadDepend)
	if err != nil {
		return fmt.Errorf("error processing collector dependencies: %v", err)
	}
	Version.DependenciesVersion = newVersion

	if err := utils.WriteJSON(config.GetVersionPath(), &Version); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	err = utils.Unzip(config.GetDownloadFilePath("dependencies", config.AS400), utils.GetMyPath())
	if err != nil {
		return fmt.Errorf("error unzipping dependencies.zip: %v", err)
	}

	err = os.Remove(config.GetDownloadFilePath("dependencies", config.AS400))
	if err != nil {
		h.WriteError("error deleting dependencies.zip file", err)
	}

	return nil
}

func processUpdateResponse(updateResponse *UpdateResponse, filepath string, download bool) (string, bool, error) {
	switch {
	case updateResponse.Message == "Dependency not found", strings.Contains(updateResponse.Message, "Error getting dependency file"):
		return updateResponse.Version, false, fmt.Errorf("error getting dependency file: %v", updateResponse.Message)
	case updateResponse.Message == "Dependency already up to date":
		return updateResponse.Version, false, nil
	case strings.Contains(updateResponse.Message, "Dependency update available: v"):
		version := strings.TrimPrefix(updateResponse.Message, "Dependency update available: v")
		if download {
			if err := utils.WriteBytesToFile(filepath, updateResponse.File); err != nil {
				return "", false, fmt.Errorf("error writing dependency file %s: %v", filepath, err)
			}
		}
		return version, true, nil
	default:
		return "", false, fmt.Errorf("error processing update response: %v", updateResponse.Message)
	}
}

func getOS() OSystem {
	if runtime.GOOS == "windows" {
		return OSystem_OS_WINDOWS
	} else {
		return OSystem_OS_LINUX
	}
}
