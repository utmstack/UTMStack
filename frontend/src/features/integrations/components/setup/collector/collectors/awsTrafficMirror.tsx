import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsTrafficMirror'

registerCollector({
  getName: () => 'AWS_TRAFFIC_MIRROR',
  sections: [
    {
      id: 'create-target',
      titleKey: `${ROOT}.sections.createTarget.title`,
      bodyKey: `${ROOT}.sections.createTarget.body`,
    },
    {
      id: 'create-filter',
      titleKey: `${ROOT}.sections.createFilter.title`,
      bodyKey: `${ROOT}.sections.createFilter.body`,
    },
    {
      id: 'create-session',
      titleKey: `${ROOT}.sections.createSession.title`,
      bodyKey: `${ROOT}.sections.createSession.body`,
    },
    {
      id: 'iam-prerequisite',
      titleKey: `${ROOT}.sections.iamPrerequisite.title`,
      bodyKey: `${ROOT}.sections.iamPrerequisite.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="vxlan (4789/udp)" sourceType="aws_traffic_mirror" />,
})
