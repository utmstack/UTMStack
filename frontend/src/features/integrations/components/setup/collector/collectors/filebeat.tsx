import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.filebeat'
const IMG = '/integrations/guides/collector/filebeat'

registerCollector({
  getName: () => 'LINUX_LOGS',
  matches: (n) => n === 'filebeat' || n === 'linux_logs',
  sections: [
    {
      id: 'install-rsyslog',
      titleKey: `${ROOT}.sections.installRsyslog.title`,
      bodyKey: `${ROOT}.sections.installRsyslog.body`,
    },
    {
      id: 'configure-rsyslog',
      titleKey: `${ROOT}.sections.configureRsyslog.title`,
      bodyKey: `${ROOT}.sections.configureRsyslog.body`,
    },
    {
      id: 'restart-rsyslog',
      titleKey: `${ROOT}.sections.restartRsyslog.title`,
      bodyKey: `${ROOT}.sections.restartRsyslog.body`,
    },
    {
      id: 'enable-collector',
      titleKey: `${ROOT}.sections.enableCollector.title`,
      bodyKey: `${ROOT}.sections.enableCollector.body`,
      image: `${IMG}/log-collector.png`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/udp" sourceType="linux_logs" />,
})
