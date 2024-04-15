package server

import (
	"regexp"
	"strings"
	"time"

	syslog "github.com/RackSec/srslog"
	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/bdgz/constants"
	"github.com/utmstack/config-client-go/types"
)

type EpsSyslogHelper struct {
	clientSyslog *syslog.Writer
}

// Initialing connection with syslogServer
func (g *EpsSyslogHelper) Init() {
	for {
		clientSyslog, err := syslog.Dial(constants.GetSyslogProto(), constants.GetSyslogHost()+":"+constants.GetSyslogPort(),
			syslog.LOG_SYSLOG, "bitdefender")
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		g.clientSyslog = clientSyslog
		break
	}
}

// SentToSyslog send event by event to syslog server
func (g *EpsSyslogHelper) SentToSyslog(config *types.ConfigurationSection, events []string, h *logger.Logger) {
	for _, syslogMessage := range events {
		for _, cnf := range config.ConfigurationGroups {
			companiesIDs := strings.Split(cnf.Configurations[3].ConfValue, ",")
			for _, compID := range companiesIDs {
				pattern := "BitdefenderGZCompanyId=" + compID
				match, err := regexp.MatchString(pattern, syslogMessage)
				if err != nil {
					h.ErrorF("error matching pattern: %v", err)
					continue
				}
				if match {
					syslogMessage += " UTM_TENANT=" + cnf.GroupName
					g.clientSyslog.Warning(syslogMessage)
					h.Info("message recived: %s", syslogMessage)
					break
				} else {
					h.Info("Event received that is not within the configured CompanyId: %s", syslogMessage)
				}
			}
		}
	}
}
