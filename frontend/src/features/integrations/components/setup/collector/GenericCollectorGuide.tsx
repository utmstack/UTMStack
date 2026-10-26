import { ChannelGuide } from '@/features/integrations/components/setup/ChannelGuide'
import type { Integration } from '@/features/integrations/types'

interface GenericCollectorGuideProps {
  integration: Integration
}

export function GenericCollectorGuide({ integration: i }: GenericCollectorGuideProps) {
  return <ChannelGuide integration={{ ...i, guideChannel: i.guideChannel ?? 'syslog' }} />
}
