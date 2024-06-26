package agent

import (
	"context"
	"fmt"

	"github.com/utmstack/UTMStack/agent-manager/updates"
)

func (s *Grpc) CheckAgentUpdates(ctx context.Context, in *UpdateAgentRequest) (*UpdateResponse, error) {
	depenType := in.GetDependType()
	oSys := in.GetOs()
	currentVersion := in.GetCurrentVersion()

	agentUpdater := updates.NewAgentUpdater()

	latestVersion := agentUpdater.GetLatestVersion(depenType.String())
	if latestVersion == "" {
		return &UpdateResponse{
			Message: "Dependency not found",
			Version: "",
			File:    []byte{},
		}, nil
	} else if latestVersion == currentVersion {
		return &UpdateResponse{
			Message: "Dependency already up to date",
			Version: latestVersion,
			File:    []byte{},
		}, nil
	} else {
		file, err := agentUpdater.GetFile(oSys.String(), depenType.String())
		if err != nil {
			return &UpdateResponse{
				Message: fmt.Sprintf("Error getting dependency file: %v", err),
				Version: latestVersion,
				File:    []byte{},
			}, nil
		}
		return &UpdateResponse{
			Message: "Dependency update available",
			Version: latestVersion,
			File:    file,
		}, nil
	}
}

func (s *Grpc) CheckCollectorUpdates(ctx context.Context, in *UpdateCollectorRequest) (*UpdateResponse, error) {
	collectorType := in.GetCollectorType()
	depenType := in.GetDependType()
	oSys := in.GetOs()
	currentVersion := in.GetCurrentVersion()

	var updater updates.UpdaterInterf
	switch collectorType {
	case CollectorType_AS_400_COLLECTOR:
		updater = updates.NewAs400Updater()
	default:
		return &UpdateResponse{
			Message: "Invalid collector type",
			Version: "",
			File:    []byte{},
		}, fmt.Errorf("invalid collector type")
	}

	latestVersion := updater.GetLatestVersion(depenType.String())
	if latestVersion == "" {
		return &UpdateResponse{
			Message: "Dependency not found",
			Version: "",
			File:    []byte{},
		}, nil
	} else if latestVersion == currentVersion {
		return &UpdateResponse{
			Message: "Dependency already up to date",
			Version: latestVersion,
			File:    []byte{},
		}, nil
	} else {
		file, err := updater.GetFile(oSys.String(), depenType.String())
		if err != nil {
			return &UpdateResponse{
				Message: fmt.Sprintf("Error getting dependency file: %v", err),
				Version: latestVersion,
				File:    []byte{},
			}, nil
		}
		return &UpdateResponse{
			Message: "Dependency update available",
			Version: latestVersion,
			File:    file,
		}, nil
	}
}
