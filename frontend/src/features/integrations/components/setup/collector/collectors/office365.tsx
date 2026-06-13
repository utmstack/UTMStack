import { registerCollector } from '../registry'
import { M365Guide } from './M365Guide'

registerCollector({
  getName: () => 'O365',
  matches: (n) => n === 'office365' || n === 'o365' || n.includes('microsoft 365') || n.includes('m365'),
  sections: [],
  render: (m) => <M365Guide module={m} />,
})
