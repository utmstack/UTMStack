import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.salesforce'
const IMG = '/integrations/guides/collector/salesforce'

registerCollector({
  getName: () => 'SALESFORCE',
  sections: [
    {
      id: 'open-setup',
      titleKey: `${ROOT}.sections.openSetup.title`,
      bodyKey: `${ROOT}.sections.openSetup.body`,
      image: `${IMG}/picture-11.png`,
    },
    {
      id: 'app-manager',
      titleKey: `${ROOT}.sections.appManager.title`,
      bodyKey: `${ROOT}.sections.appManager.body`,
      image: `${IMG}/picture-12.png`,
    },
    {
      id: 'new-connected-app',
      titleKey: `${ROOT}.sections.newConnectedApp.title`,
      bodyKey: `${ROOT}.sections.newConnectedApp.body`,
      image: `${IMG}/picture-13.png`,
    },
    {
      id: 'oauth-settings',
      titleKey: `${ROOT}.sections.oauthSettings.title`,
      bodyKey: `${ROOT}.sections.oauthSettings.body`,
      image: `${IMG}/picture-14.png`,
    },
    {
      id: 'oauth-scopes',
      titleKey: `${ROOT}.sections.oauthScopes.title`,
      bodyKey: `${ROOT}.sections.oauthScopes.body`,
      image: `${IMG}/picture-15.png`,
    },
    {
      id: 'save',
      titleKey: `${ROOT}.sections.save.title`,
      bodyKey: `${ROOT}.sections.save.body`,
      image: `${IMG}/picture-17.png`,
    },
    {
      id: 'manage-consumer',
      titleKey: `${ROOT}.sections.manageConsumer.title`,
      bodyKey: `${ROOT}.sections.manageConsumer.body`,
      image: `${IMG}/picture-19.png`,
    },
    {
      id: 'copy-credentials',
      titleKey: `${ROOT}.sections.copyCredentials.title`,
      bodyKey: `${ROOT}.sections.copyCredentials.body`,
      image: `${IMG}/picture-21.png`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="salesforce" />,
})
