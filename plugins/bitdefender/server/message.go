package server

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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
					_ = catcher.Error("error matching pattern", err, map[string]any{})
					continue
				}

				if !match {
					continue
				}

				processor.LogsChan <- &plugins.Log{
					Id:         uuid.New().String(),
					TenantId:   configuration.DefaultTenant,
					DataType:   "antivirus-bitdefender-gz",
					DataSource: cnf.GroupName,
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					Raw:        syslogMessage,
				}
				break
			}
		}
	}
}
