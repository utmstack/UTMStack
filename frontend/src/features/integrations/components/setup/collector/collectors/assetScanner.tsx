import { registerCollector } from '../registry'

const ROOT = 'integrations.setup.collector.assetScanner'

registerCollector({
  getName: () => 'ASSET_MANAGEMENT',
  matches: (n) => n === 'asset_scanner' || n === 'asset_management',
  sections: [
    {
      id: 'overview',
      titleKey: `${ROOT}.sections.overview.title`,
      bodyKey: `${ROOT}.sections.overview.body`,
    },
    {
      id: 'agent-requirement',
      titleKey: `${ROOT}.sections.agentRequirement.title`,
      bodyKey: `${ROOT}.sections.agentRequirement.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: () => null,
})
