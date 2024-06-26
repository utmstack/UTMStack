package agent

import (
	context "context"
	"fmt"
	"runtime"
	"strconv"

	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/models"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UpdateDependencies(conn *grpc.ClientConn, cnf *configuration.Config, versions models.Version) (map[string]*UpdateResponse, error) {
	dependClient := NewUpdatesServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))

	mapResponses := make(map[string]*UpdateResponse)

	resp, err := dependClient.CheckAgentUpdates(ctx, &UpdateAgentRequest{
		DependType:     DependType_DEPEND_TYPE_SERVICE,
		Os:             getOS(),
		CurrentVersion: versions.ServiceVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("error checking agent service: %v", err)
	}
	mapResponses["service"] = resp

	resp, err = dependClient.CheckAgentUpdates(ctx, &UpdateAgentRequest{
		DependType:     DependType_DEPEND_TYPE_DEPEND_ZIP,
		Os:             getOS(),
		CurrentVersion: versions.DependenciesVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("error checking dependencies: %v", err)
	}
	mapResponses["dependencies"] = resp

	return mapResponses, nil
}

func getOS() OSystem {
	if runtime.GOOS == "windows" {
		return OSystem_OS_WINDOWS
	} else {
		return OSystem_OS_LINUX
	}
}
