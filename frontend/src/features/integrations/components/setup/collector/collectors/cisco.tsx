import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.cisco'
const IMG = '/integrations/guides/collector/cisco'

registerCollector({
  getName: () => 'CISCO',
  sections: [
    {
      id: 'send-logs',
      titleKey: `${ROOT}.sections.sendLogs.title`,
      bodyKey: `${ROOT}.sections.sendLogs.body`,
    },
    {
      id: 'enable-collector',
      titleKey: `${ROOT}.sections.enableCollector.title`,
      bodyKey: `${ROOT}.sections.enableCollector.body`,
      image: `${IMG}/cisco.png`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/udp" sourceType="cisco" />,
})
