import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.crowdstrike'

registerCollector({
  getName: () => 'CROWDSTRIKE',
  sections: [
    {
      id: 'create-api-client',
      titleKey: `${ROOT}.sections.createApiClient.title`,
      bodyKey: `${ROOT}.sections.createApiClient.body`,
    },
    {
      id: 'set-scopes',
      titleKey: `${ROOT}.sections.setScopes.title`,
      bodyKey: `${ROOT}.sections.setScopes.body`,
    },
    {
      id: 'copy-credentials',
      titleKey: `${ROOT}.sections.copyCredentials.title`,
      bodyKey: `${ROOT}.sections.copyCredentials.body`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="crowdstrike" />,
})
