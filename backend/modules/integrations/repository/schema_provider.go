package repository

import (
	"strings"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
)

var moduleSchemas = map[string]map[string]string{
	"AWS_IAM_USER": {
		"aws_access_key_id":     "text",
		"aws_secret_access_key": "password",
		"aws_default_region":    "text",
	},
	"AZURE": {
		"eventHubConnection": "password", // connection string contains SharedAccessKey
		"consumerGroup":      "text",
		"storageContainer":   "text",
		"storageConnection":  "password", // connection string contains AccountKey
	},
	"GCP": {
		"jsonKey":      "file",
		"projectId":    "text",
		"subscription": "text",
	},
	"O365": {
		"office365_client_id":         "text",
		"office365_client_secret":     "password",
		"office365_tenant_id":         "text",
		"office365_cloud_environment": "text",
	},
	"BITDEFENDER": {
		"connectionKey": "password",
		"accessUrl":     "text",
		"utmPublicIP":   "text",
		"companyIds":    "text",
	},
	"CROWDSTRIKE": {
		"crowdstrike_client_id":        "text",
		"crowdstrike_client_secret":    "password",
		"crowdstrike_cloud_region_url": "text",
		"crowdstrike_app_name":         "text",
	},
	"SOPHOS": {
		"sophos_client_id": "text",
		"sophos_x_api_key": "password",
	},
}

type CodeSchemaProvider struct{}

func NewCodeSchemaProvider() *CodeSchemaProvider { return &CodeSchemaProvider{} }

var _ connectors.SchemaProvider = (*CodeSchemaProvider)(nil)

func (CodeSchemaProvider) Schema(module string) (map[string]string, error) {
	s, ok := moduleSchemas[strings.ToUpper(module)]
	if !ok {
		return nil, domain.ErrNotConfigurable
	}
	return s, nil
}
