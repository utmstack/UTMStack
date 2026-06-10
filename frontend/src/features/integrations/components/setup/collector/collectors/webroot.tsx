import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.webroot'
const IMG = '/integrations/guides/collector/webroot'

registerCollector({
  getName: () => 'WEBROOT',
  sections: [
    {
      id: 'login',
      titleKey: `${ROOT}.sections.login.title`,
      bodyKey: `${ROOT}.sections.login.body`,
      image: `${IMG}/sites-tab.png`,
    },
    {
      id: 'open-settings',
      titleKey: `${ROOT}.sections.openSettings.title`,
      bodyKey: `${ROOT}.sections.openSettings.body`,
      image: `${IMG}/click-settings.png`,
    },
    {
      id: 'api-access',
      titleKey: `${ROOT}.sections.apiAccess.title`,
      bodyKey: `${ROOT}.sections.apiAccess.body`,
      image: `${IMG}/api-access-tab.png`,
    },
    {
      id: 'create-credential',
      titleKey: `${ROOT}.sections.createCredential.title`,
      bodyKey: `${ROOT}.sections.createCredential.body`,
      image: `${IMG}/create-credential.png`,
    },
    {
      id: 'note-client-secret',
      titleKey: `${ROOT}.sections.noteClientSecret.title`,
      bodyKey: `${ROOT}.sections.noteClientSecret.body`,
      image: `${IMG}/credential-record.png`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="webroot" />,
})
