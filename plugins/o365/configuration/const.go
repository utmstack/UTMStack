package configuration

import (
	"fmt"
)

const (
	loginUrl                  = "https://login.microsoftonline.com/"
	GRANTTYPE                 = "client_credentials"
	SCOPE                     = "https://manage.office.com/.default"
	endPointLogin             = "/oauth2/v2.0/token"
	endPointStartSubscription = "/activity/feed/subscriptions/start"
	endPointContent           = "/activity/feed/subscriptions/content"
	BASEURL                   = "https://manage.office.com/api/v1.0/"
	DefaultTenant             = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

func GetMicrosoftLoginLink(tenant string) string {
	return fmt.Sprintf("%s%s%s", loginUrl, tenant, endPointLogin)
}

func GetStartSubscriptionLink(tenant string) string {
	return fmt.Sprintf("%s%s%s", BASEURL, tenant, endPointStartSubscription)
}

func GetContentLink(tenant string) string {
	return fmt.Sprintf("%s%s%s", BASEURL, tenant, endPointContent)
}

func GetTenantId() string {
	return DefaultTenant
}
