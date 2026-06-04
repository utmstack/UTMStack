// Package bitdefender implements the BITDEFENDER module kind. The
// "companyIds" field uses the legacy "list" data type — a comma-separated
// CSV — so we keep that string as-is rather than promoting it to a real
// list type; the panel UI and the collector both already speak it.
package bitdefender

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "BITDEFENDER"

// confDataTypeList is the legacy CSV-list data type Bitdefender alone uses.
const confDataTypeList = "list"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "connectionKey",
			ConfName: "Connection key", ConfDescription: "Bitdefender connection key",
			ConfDataType: domain.ConfTypePassword, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "accessUrl",
			ConfName: "Access URL", ConfDescription: "Bitdefender access URL",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "utmPublicIP",
			ConfName: "Master public IP or DNS", ConfDescription: "Master instance public IP or DNS",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "companyIds",
			ConfName:        "Companies IDs",
			ConfDescription: "Separate the company IDs to be associated with this tenant by commas, for example: BDGZ1234,BDGZ5678,BDGZ9012",
			ConfDataType:    confDataTypeList, ConfRequired: true,
		},
	}
}
