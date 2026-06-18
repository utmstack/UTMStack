import { registerCollector } from '../registry'
import { SophosGuide } from './SophosGuide'

registerCollector({
  getName: () => 'SOPHOS',
  matches: (n) => n.includes('sophos') && !n.includes('xg'),
  sections: [],
  render: (m) => <SophosGuide module={m} />,
})
