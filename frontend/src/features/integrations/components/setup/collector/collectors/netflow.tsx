import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.netflow'
const IMG = '/integrations/guides/collector/netflow'

registerCollector({
  getName: () => 'NETFLOW',
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
      image: `${IMG}/netflow.png`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="2055/udp" sourceType="netflow" />,
})
