import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.oracle'

registerCollector({
  getName: () => 'ORACLE',
  sections: [
    {
      id: 'enable-audit',
      titleKey: `${ROOT}.sections.enableAudit.title`,
      bodyKey: `${ROOT}.sections.enableAudit.body`,
    },
    {
      id: 'configure-trail',
      titleKey: `${ROOT}.sections.configureTrail.title`,
      bodyKey: `${ROOT}.sections.configureTrail.body`,
    },
    {
      id: 'export-trail',
      titleKey: `${ROOT}.sections.exportTrail.title`,
      bodyKey: `${ROOT}.sections.exportTrail.body`,
    },
    {
      id: 'enable-collector',
      titleKey: `${ROOT}.sections.enableCollector.title`,
      bodyKey: `${ROOT}.sections.enableCollector.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="514/udp" sourceType="oracle" />,
})
