// Package aws_iam_user implements the AWS_IAM_USER module kind. The panel
// supplies an IAM access key pair, region, and log group; the collector tails
// CloudWatch logs from the named group.
package aws_iam_user

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "AWS_IAM_USER"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "aws_access_key_id",
			ConfName: "Access Key", ConfDescription: "Configure Aws Iam User access key",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "aws_default_region",
			ConfName: "Default Region", ConfDescription: "Configure Aws Iam User default region",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "aws_secret_access_key",
			ConfName: "Secret Key", ConfDescription: "Configure Aws Iam User secret kew",
			ConfDataType: domain.ConfTypePassword, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "aws_log_group_name",
			ConfName: "Log Group Name", ConfDescription: "Configure Aws Log Group Name",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
	}
}
