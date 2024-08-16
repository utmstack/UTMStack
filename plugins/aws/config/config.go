package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/utmstack/UTMStack/plugins/aws/schema"
	"github.com/utmstack/config-client-go/types"
)

const DefaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

func NewAWSSession(accessKeyID, secretAccessKey, region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
}

func CompareConfigs(savedConfs map[string]schema.AWSConfig, newConfs []types.ModuleGroup) bool {
	for _, newConf := range newConfs {
		cnf, ok := savedConfs[newConf.GroupName]
		if !ok {
			return true
		}

		if cnf.AccessKeyID != newConf.Configurations[0].ConfValue ||
			cnf.SecretAccessKey != newConf.Configurations[1].ConfValue ||
			cnf.Region != newConf.Configurations[2].ConfValue ||
			cnf.LogGroupName != newConf.Configurations[3].ConfValue ||
			cnf.LogStreamName != newConf.Configurations[4].ConfValue {
			return true
		}
	}
	return false
}
