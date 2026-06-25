package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
	"google.golang.org/grpc/metadata"
)

type versionFile struct {
	Version string `json:"version"`
}

func getForwarderVersion() (string, error) {
	versionPath := "version.json"
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return "1.0.0", nil
	}
	var v versionFile
	if err := json.Unmarshal(data, &v); err != nil {
		return "1.0.0", nil
	}
	return v.Version, nil
}

func RegisterCollector(cnf *config.Config, UTMKey string) error {
	connection, err := GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	client := NewCollectorServiceClient(connection)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", UTMKey)
	defer cancel()

	ip, err := utils.GetIPAddress()
	if err != nil {
		return fmt.Errorf("error getting ip address: %v", err)
	}

	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return fmt.Errorf("error getting os info: %v", err)
	}

	version, err := getForwarderVersion()
	if err != nil {
		return fmt.Errorf("error getting version: %v", err)
	}

	request := &RegisterRequest{
		Ip:        ip,
		Hostname:  osInfo.Hostname,
		Version:   version,
		Collector: CollectorModule_FORWARDER,
	}

	response, err := client.RegisterCollector(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to register collector: %v", err)
	}

	cnf.CollectorID = uint(response.Id)
	cnf.CollectorKey = response.Key

	utils.Logger.LogF(100, "Collector registered with ID: %v", cnf.CollectorID)

	return nil
}
