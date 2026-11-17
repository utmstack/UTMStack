import { useTranslation } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.fortiweb'
const PORT = '7018'

function FortiwebGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-fortiweb">
      <></>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'FORTIWEB',
  matches: (n) => n === 'fortiweb' || n === 'firewall_fortiweb',
  sections: [],
  render: (m) => <FortiwebGuide module={m} />,
})
