package agent

import (
	"context"

	"github.com/utmstack/UTMStack/utmstack-collector/config"
	"github.com/utmstack/UTMStack/utmstack-collector/conn"
	"github.com/utmstack/UTMStack/utmstack-collector/models"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
	"google.golang.org/grpc/metadata"
)

func RegisterCollector(cnf *config.Config, UTMKey string) error {
	connection, err := conn.GetAgentManagerConnection(cnf)
	if err != nil {
		return utils.Logger.ErrorF("error connecting to Agent Manager: %v", err)
	}

	collectorClient := NewCollectorServiceClient(connection)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", UTMKey)
	defer cancel()

	ip, err := utils.GetIPAddress()
	if err != nil {
		return utils.Logger.ErrorF("error getting ip address: %v", err)
	}

	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return utils.Logger.ErrorF("error getting os info: %v", err)
	}

	version := models.Version{}
	err = utils.ReadJson(config.VersionPath, &version)
	if err != nil {
		return utils.Logger.ErrorF("error reading version file: %v", err)
	}

	request := &RegisterRequest{
		Ip:        ip,
		Hostname:  osInfo.Hostname,
		Version:   version.Version,
		Collector: CollectorModule_UTMSTACK,
	}

	utils.Logger.Info("Registering UTMStack Collector with Agent Manager...")
	utils.Logger.Info("Collector Details: IP=%s, Hostname=%s, Version=%s, Module=%s",
		ip, osInfo.Hostname, version.Version, CollectorModule_UTMSTACK.String())

	response, err := collectorClient.RegisterCollector(ctx, request)
	if err != nil {
		return utils.Logger.ErrorF("failed to register collector: %v", err)
	}

	cnf.CollectorID = uint(response.Id)
	cnf.CollectorKey = response.Key

	utils.Logger.Info("UTMStack Collector registered successfully")
	utils.Logger.Info("Collector ID: %d", cnf.CollectorID)

	return nil
}
