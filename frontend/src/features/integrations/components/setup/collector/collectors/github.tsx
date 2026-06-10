import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.github'
const IMG = '/integrations/guides/collector/github'

registerCollector({
  getName: () => 'GITHUB',
  sections: [
    {
      id: 'open-settings',
      titleKey: `${ROOT}.sections.openSettings.title`,
      bodyKey: `${ROOT}.sections.openSettings.body`,
      image: `${IMG}/webhook.png`,
    },
    {
      id: 'add-webhook',
      titleKey: `${ROOT}.sections.addWebhook.title`,
      bodyKey: `${ROOT}.sections.addWebhook.body`,
      image: `${IMG}/github-webhook.png`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="github" />,
})
