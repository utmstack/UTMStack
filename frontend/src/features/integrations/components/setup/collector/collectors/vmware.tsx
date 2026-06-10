import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.vmware'

registerCollector({
  getName: () => 'VMWARE',
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
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/tcp" sourceType="vmware" />,
})
