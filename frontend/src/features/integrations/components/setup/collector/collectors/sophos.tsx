import { registerCollector } from '../registry'
import { SophosGuide } from './SophosGuide'

registerCollector({
  getName: () => 'SOPHOS',
  matches: (n) => n === 'sophos' || n === 'sophos_central',
  sections: [],
  render: (m) => <SophosGuide module={m} />,
})
