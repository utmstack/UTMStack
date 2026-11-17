import { useTranslation  } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide  } from '../ForwarderGuide'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.fortigate'
const PORT = '7005'

function FortigateGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-fortigate-traffic">
    <></>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'FORTIGATE',
  matches: (n) => n === 'fortigate' || n === 'firewall_fortigate_traffic',
  sections: [],
  render: (m) => <FortigateGuide module={m} />,
})
