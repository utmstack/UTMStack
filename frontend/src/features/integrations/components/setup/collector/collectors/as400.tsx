import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.as400'

registerCollector({
  getName: () => 'AS_400',
  matches: (n) => n === 'as400' || n === 'as_400',
  sections: [
    {
      id: 'requirements',
      titleKey: `${ROOT}.sections.requirements.title`,
      bodyKey: `${ROOT}.sections.requirements.body`,
    },
    {
      id: 'install-collector',
      titleKey: `${ROOT}.sections.installCollector.title`,
      bodyKey: `${ROOT}.sections.installCollector.body`,
    },
    {
      id: 'configure',
      titleKey: `${ROOT}.sections.configure.title`,
      bodyKey: `${ROOT}.sections.configure.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/tcp" sourceType="as400" />,
})
