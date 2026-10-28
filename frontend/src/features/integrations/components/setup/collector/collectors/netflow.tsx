import { registerCollector } from '../registry'
import { ChannelGuide } from '@/features/integrations/components/setup/ChannelGuide'
import type { Integration } from '@/features/integrations/types'

function NetflowGuide({ module }: { module: Integration }) {
  return <ChannelGuide integration={{ ...module, guideChannel: 'netflow' }} />
}

registerCollector({
  getName: () => 'NETFLOW',
  sections: [],
  render: (m) => <NetflowGuide module={m} />,
})
