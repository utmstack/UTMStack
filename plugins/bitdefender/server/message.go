package server

import (
	"regexp"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/plugins/bitdefender/config"
)

func CreateMessage(cnf *config.ConfigurationSection, events []string) {
	for _, syslogMessage := range events {
		for _, cnf := range cnf.ModuleGroups {
			moduleConfig := config.GetBDGZModuleConfig(cnf)

			for _, compID := range moduleConfig.CompaniesIDs {
				pattern := "BitdefenderGZCompanyId=" + compID
				match, err := regexp.MatchString(pattern, syslogMessage)
				if err != nil {
					_ = catcher.Error("error matching pattern", err, nil)
					continue
				}

				if !match {
					continue
				}

				tenantId := cnf.TenantId
				if tenantId == "" {
					tenantId = config.DefaultTenant
				}

				_ = plugins.EnqueueLog(&plugins.Log{
					Id:         uuid.New().String(),
					TenantId:   tenantId,
					DataType:   "antivirus-bitdefender-gz",
					DataSource: cnf.GroupName,
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					Raw:        syslogMessage,
				}, "com.utmstack.bitdefender")

				break
			}
		}
	}
}
