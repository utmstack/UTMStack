package server

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/bitdefender/configuration"
	"github.com/utmstack/UTMStack/plugins/bitdefender/processor"
	"github.com/utmstack/config-client-go/types"
)

func CreateMessage(config *types.ConfigurationSection, events []string) {
	for _, syslogMessage := range events {
		for _, cnf := range config.ConfigurationGroups {
			companiesIDs := strings.Split(cnf.Configurations[3].ConfValue, ",")
			for _, compID := range companiesIDs {
				pattern := "BitdefenderGZCompanyId=" + compID
				match, err := regexp.MatchString(pattern, syslogMessage)
				if err != nil {
					go_sdk.Logger().ErrorF("error matching pattern: %v", err)
					continue
				}
				if match {
					processor.LogsChan <- &go_sdk.Log{
						Id:         uuid.New().String(),
						TenantId:   configuration.DefaultTenant,
						DataType:   "antivirus-bitdefender-gz",
						DataSource: cnf.GroupName,
						Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
						Raw:        syslogMessage,
					}
					go_sdk.Logger().Info("message recived: %s", syslogMessage)
					break
				} else {
					go_sdk.Logger().Info("Event received that is not within the configured CompanyId: %s", syslogMessage)
				}
			}
		}
	}
}
