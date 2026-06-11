import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.sentinelOne'

registerCollector({
  getName: () => 'SENTINEL_ONE',
  matches: (n) => n.includes('sentinel'),
  sections: [
    {
      id: 'create-api-token',
      titleKey: `${ROOT}.sections.createApiToken.title`,
      bodyKey: `${ROOT}.sections.createApiToken.body`,
    },
    {
      id: 'console-url',
      titleKey: `${ROOT}.sections.consoleUrl.title`,
      bodyKey: `${ROOT}.sections.consoleUrl.body`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="sentinel_one" />,
})
