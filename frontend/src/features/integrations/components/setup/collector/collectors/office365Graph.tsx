import { registerCollector } from '../registry'
import { M365Guide, O365_GRAPH_FIELDS } from './M365Guide'

registerCollector({
  getName: () => 'O365_GRAPH',
  matches: (n) => n === 'o365_graph' || n === 'o365graph' || n.includes('microsoft graph') || n.includes('graph api'),
  sections: [],
  render: (m) => <M365Guide module={m} fields={O365_GRAPH_FIELDS} variant="graph" />,
})
