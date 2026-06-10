import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.utmstack'

registerCollector({
  getName: () => 'UTMSTACK',
  sections: [
    {
      id: 'requirements',
      titleKey: `${ROOT}.sections.requirements.title`,
      bodyKey: `${ROOT}.sections.requirements.body`,
    },
    {
      id: 'install-agent',
      titleKey: `${ROOT}.sections.installAgent.title`,
      bodyKey: `${ROOT}.sections.installAgent.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="agent (local)" sourceType="utmstack" />,
})
