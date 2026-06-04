// Package crowdstrike implements the CROWDSTRIKE module kind. The optional
// "app_name" field is the OAuth application name CrowdStrike requires for
// the audit-log scope; it's the only non-required key.
package crowdstrike

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "CROWDSTRIKE"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "crowdstrike_client_id",
			ConfName: "Client ID", ConfDescription: "CrowdStrike Client ID",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "crowdstrike_client_secret",
			ConfName: "Client Secret", ConfDescription: "CrowdStrike Client Secret",
			ConfDataType: domain.ConfTypePassword, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "crowdstrike_cloud_region_url",
			ConfName: "Cloud Region URL", ConfDescription: "CrowdStrike Cloud Region URL",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "crowdstrike_app_name",
			ConfName: "App Name", ConfDescription: "App Name for CrowdStrike integration",
			ConfDataType: domain.ConfTypeText, ConfRequired: false,
		},
	}
}
