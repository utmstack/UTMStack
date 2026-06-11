import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.bitdefender'
const IMG = '/integrations/guides/collector/bitdefender'

registerCollector({
  getName: () => 'BITDEFENDER',
  sections: [
    {
      id: 'my-account',
      titleKey: `${ROOT}.sections.myAccount.title`,
      bodyKey: `${ROOT}.sections.myAccount.body`,
      image: `${IMG}/step2.png`,
    },
    {
      id: 'access-url',
      titleKey: `${ROOT}.sections.accessUrl.title`,
      bodyKey: `${ROOT}.sections.accessUrl.body`,
      image: `${IMG}/step3.png`,
    },
    {
      id: 'create-api-key',
      titleKey: `${ROOT}.sections.createApiKey.title`,
      bodyKey: `${ROOT}.sections.createApiKey.body`,
      image: `${IMG}/step4.png`,
    },
    {
      id: 'my-company',
      titleKey: `${ROOT}.sections.myCompany.title`,
      bodyKey: `${ROOT}.sections.myCompany.body`,
      image: `${IMG}/bdgz4.png`,
    },
    {
      id: 'licensing',
      titleKey: `${ROOT}.sections.licensing.title`,
      bodyKey: `${ROOT}.sections.licensing.body`,
      image: `${IMG}/bdgz5.png`,
    },
    {
      id: 'add-instances',
      titleKey: `${ROOT}.sections.addInstances.title`,
      bodyKey: `${ROOT}.sections.addInstances.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/udp" sourceType="bitdefender" />,
})
