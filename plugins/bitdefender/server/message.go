package server

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
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
					helpers.Logger().ErrorF("error matching pattern: %v", err)
					continue
				}
				if match {
					processor.LogsChan <- &plugins.Log{
						Id:         uuid.New().String(),
						TenantId:   configuration.DefaultTenant,
						DataType:   "antivirus-bitdefender-gz",
						DataSource: cnf.GroupName,
						Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
						Raw:        syslogMessage,
					}
					helpers.Logger().Info("message recived: %s", syslogMessage)
					break
				} else {
					helpers.Logger().Info("Event received that is not within the configured CompanyId: %s", syslogMessage)
				}
			}
		}
	}
}
