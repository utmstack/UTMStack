import { registerCollector } from '../registry'

const ROOT = 'integrations.setup.collector.vulnerabilities'

registerCollector({
  getName: () => 'VULNERABILITIES',
  sections: [
    {
      id: 'overview',
      titleKey: `${ROOT}.sections.overview.title`,
      bodyKey: `${ROOT}.sections.overview.body`,
    },
    {
      id: 'scope',
      titleKey: `${ROOT}.sections.scope.title`,
      bodyKey: `${ROOT}.sections.scope.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: () => null,
})
