import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.iis'
const IMG = '/integrations/guides/collector/iis'

registerCollector({
  getName: () => 'IIS',
  sections: [
    {
      id: 'enable-module',
      titleKey: `${ROOT}.sections.enableModule.title`,
      bodyKey: `${ROOT}.sections.enableModule.body`,
    },
    {
      id: 'configure-module',
      titleKey: `${ROOT}.sections.configureModule.title`,
      bodyKey: `${ROOT}.sections.configureModule.body`,
      image: `${IMG}/config-iis-windows.png`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="filebeat/local" sourceType="iis" />,
})
