import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.kaspersky'
const IMG = '/integrations/guides/collector/kaspersky'

registerCollector({
  getName: () => 'KASPERSKY',
  sections: [
    {
      id: 'console-settings',
      titleKey: `${ROOT}.sections.consoleSettings.title`,
      bodyKey: `${ROOT}.sections.consoleSettings.body`,
      image: `${IMG}/main_page.png`,
    },
    {
      id: 'siem',
      titleKey: `${ROOT}.sections.siem.title`,
      bodyKey: `${ROOT}.sections.siem.body`,
      image: `${IMG}/integration.png`,
    },
    {
      id: 'configure',
      titleKey: `${ROOT}.sections.configure.title`,
      bodyKey: `${ROOT}.sections.configure.body`,
      image: `${IMG}/configuration.png`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="514/udp" sourceType="kaspersky" />,
})
