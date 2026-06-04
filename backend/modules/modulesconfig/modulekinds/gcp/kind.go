// Package gcp implements the GCP module kind. GCP Pub/Sub log delivery — the
// panel uploads a service-account JSON key (file type, encrypted at rest) and
// supplies project/subscription/topic ids for the collector.
package gcp

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "GCP"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "jsonKey",
			ConfName: "Json Key", ConfDescription: "Configure your GCP json key",
			ConfDataType: domain.ConfTypeFile, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "projectId",
			ConfName: "Project ID", ConfDescription: "Configure your GCP project ID",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "subscription",
			ConfName: "Subscription ID", ConfDescription: "Configure your GCP subscription ID",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "topic",
			ConfName: "Topic ID", ConfDescription: "Configure your GCP topic ID",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
	}
}
