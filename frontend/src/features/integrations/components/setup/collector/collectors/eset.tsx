import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.eset'
const IMG = '/integrations/guides/collector/eset'

registerCollector({
  getName: () => 'ESET',
  sections: [
    {
      id: 'open-menu',
      titleKey: `${ROOT}.sections.openMenu.title`,
      bodyKey: `${ROOT}.sections.openMenu.body`,
      image: `${IMG}/main_page.png`,
    },
    {
      id: 'server-config',
      titleKey: `${ROOT}.sections.serverConfig.title`,
      bodyKey: `${ROOT}.sections.serverConfig.body`,
      image: `${IMG}/more_settings.png`,
    },
    {
      id: 'advanced-settings',
      titleKey: `${ROOT}.sections.advancedSettings.title`,
      bodyKey: `${ROOT}.sections.advancedSettings.body`,
      image: `${IMG}/server_config.png`,
    },
    {
      id: 'syslog-server',
      titleKey: `${ROOT}.sections.syslogServer.title`,
      bodyKey: `${ROOT}.sections.syslogServer.body`,
      image: `${IMG}/syslog_server.png`,
    },
    {
      id: 'enable-collector',
      titleKey: `${ROOT}.sections.enableCollector.title`,
      bodyKey: `${ROOT}.sections.enableCollector.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/udp" sourceType="eset" />,
})
