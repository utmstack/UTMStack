import type { Integration } from '@/features/integrations/types'
import { getCollector } from './registry'
import { CollectorGuideSection } from './CollectorGuideSection'
import { GenericCollectorGuide } from './GenericCollectorGuide'
import { ChannelGuide } from '@/features/integrations/components/setup/ChannelGuide'
import './collectors'

interface CollectorSetupProps {
  integration: Integration
}

export function CollectorSetup({ integration }: CollectorSetupProps) {
  const collector = getCollector(integration.moduleName ?? integration.name)

  if (
    integration.guideChannel &&
    (integration.guideOnly || !collector || integration.guideChannel === 'windows-agent' || integration.guideChannel === 'linux-agent' ||
      integration.guideChannel === 'netflow' || integration.moduleName === 'SYSLOG')
  ) {
    return <ChannelGuide integration={integration} />
  }

  if (!collector) {
    return <GenericCollectorGuide integration={integration} />
  }

  return (
    <div className="space-y-4">
      {collector.sections.map((section) => (
        <CollectorGuideSection key={section.id} section={section} />
      ))}
      {collector.render(integration)}
    </div>
  )
}
