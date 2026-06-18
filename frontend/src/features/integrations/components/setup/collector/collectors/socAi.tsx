import { registerCollector } from '../registry'

const ROOT = 'integrations.setup.collector.socAi'

registerCollector({
  getName: () => 'SOC_AI',
  sections: [
    {
      id: 'overview',
      titleKey: `${ROOT}.sections.overview.title`,
      bodyKey: `${ROOT}.sections.overview.body`,
    },
    {
      id: 'choose-provider',
      titleKey: `${ROOT}.sections.chooseProvider.title`,
      bodyKey: `${ROOT}.sections.chooseProvider.body`,
    },
    {
      id: 'configure-credentials',
      titleKey: `${ROOT}.sections.configureCredentials.title`,
      bodyKey: `${ROOT}.sections.configureCredentials.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: () => null,
})
