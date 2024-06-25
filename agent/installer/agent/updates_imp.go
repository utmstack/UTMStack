package agent

import (
	context "context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/installer/configuration"
	"github.com/utmstack/UTMStack/agent/installer/models"
	"github.com/utmstack/UTMStack/agent/installer/utils"
	"google.golang.org/grpc/metadata"
)

func DownloadDependencies(address, authKey, skip string, h *logger.Logger) error {
	conn, err := ConnectToServer(h, address, configuration.AgentManagerPort, skip)
	if err != nil {
		return fmt.Errorf("error connecting to server to download dependencies: %v", err)
	}
	defer conn.Close()

	dependClient := NewUpdatesServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", authKey)

	Version := models.Version{}

	resp, err := dependClient.CheckAgentUpdates(ctx, &UpdateAgentRequest{
		DependType:     DependType_DEPEND_TYPE_SERVICE,
		Os:             getOS(),
		CurrentVersion: "0",
	})
	if err != nil {
		return fmt.Errorf("error checking agent service: %v", err)
	}
	v, err := processUpdateResponse(resp, configuration.GetDownloadFilePath("service"))
	if err != nil {
		return fmt.Errorf("error processing agent service: %v", err)
	}
	Version.ServiceVersion = v

	resp, err = dependClient.CheckAgentUpdates(ctx, &UpdateAgentRequest{
		DependType:     DependType_DEPEND_TYPE_DEPEND_ZIP,
		Os:             getOS(),
		CurrentVersion: "0",
	})
	if err != nil {
		return fmt.Errorf("error checking agent dependencies: %v", err)
	}
	v, err = processUpdateResponse(resp, configuration.GetDownloadFilePath("dependencies"))
	if err != nil {
		return fmt.Errorf("error processing agent dependencies: %v", err)
	}
	Version.DependenciesVersion = v

	if err := utils.WriteJSON(configuration.GetVersionPath(), &Version); err != nil {
		return fmt.Errorf("error writing version file: %v", err)
	}

	err = utils.Unzip(configuration.GetDownloadFilePath("dependencies"), utils.GetMyPath())
	if err != nil {
		return fmt.Errorf("error unzipping dependencies.zip: %v", err)
	}

	if runtime.GOOS == "linux" {
		if err = utils.Execute("chmod", utils.GetMyPath(), "-R", "777", "utmstack_updater_self"); err != nil {
			return fmt.Errorf("error executing chmod: %v", err)
		}
	}

	err = os.Remove(configuration.GetDownloadFilePath("dependencies"))
	if err != nil {
		utils.GetBeautyLogger().WriteError("error deleting dependencies.zip file", err)
	}

	return nil

}

func processUpdateResponse(updateResponse *UpdateResponse, filepath string) (string, error) {
	switch {
	case updateResponse.Message == "Dependency not found", strings.Contains(updateResponse.Message, "Error getting dependency file"), updateResponse.Message == "Dependency already up to date":
		return "", fmt.Errorf("error getting dependency file: %v", updateResponse.Message)
	case strings.Contains(updateResponse.Message, "Dependency update available: v"):
		version := strings.TrimPrefix(updateResponse.Message, "Dependency update available: v")
		if err := utils.WriteBytesToFile(filepath, updateResponse.File); err != nil {
			return "", fmt.Errorf("error writing dependency file %s: %v", filepath, err)
		}
		return version, nil
	default:
		return "", fmt.Errorf("error processing update response: %v", updateResponse.Message)
	}
}

func getOS() OSystem {
	if runtime.GOOS == "windows" {
		return OSystem_OS_WINDOWS
	} else {
		return OSystem_OS_LINUX
	}
}
