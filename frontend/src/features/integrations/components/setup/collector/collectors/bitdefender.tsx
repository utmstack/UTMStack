import { registerCollector } from '../registry'
import { BitdefenderGuide } from './BitdefenderGuide'

registerCollector({
  getName: () => 'BITDEFENDER',
  sections: [],
  render: (m) => <BitdefenderGuide module={m} />,
})
