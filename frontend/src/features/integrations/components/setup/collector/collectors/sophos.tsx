import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.sophos'

registerCollector({
  getName: () => 'SOPHOS',
  sections: [
    {
      id: 'create-api-credentials',
      titleKey: `${ROOT}.sections.createApiCredentials.title`,
      bodyKey: `${ROOT}.sections.createApiCredentials.body`,
    },
    {
      id: 'tenant-id',
      titleKey: `${ROOT}.sections.tenantId.title`,
      bodyKey: `${ROOT}.sections.tenantId.body`,
    },
    {
      id: 'fill-form',
      titleKey: `${ROOT}.sections.fillForm.title`,
      bodyKey: `${ROOT}.sections.fillForm.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="sophos" />,
})
