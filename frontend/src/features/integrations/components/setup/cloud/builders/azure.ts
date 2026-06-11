import type { CloudGuideSection, CloudConfigField } from './cloudGuideBuilder'

const ROOT = 'integrations.setup.cloud.azure'
const IMG = '/integrations/guides/azure'

export const AZURE_SECTIONS: CloudGuideSection[] = [
  {
    id: 'event-hub',
    titleKey: `${ROOT}.sections.eventHub.title`,
    bodyKey: `${ROOT}.sections.eventHub.body`,
    image: `${IMG}/event-hub-selection.png`,
  },
  {
    id: 'shared-access-policy',
    titleKey: `${ROOT}.sections.sharedAccessPolicy.title`,
    bodyKey: `${ROOT}.sections.sharedAccessPolicy.body`,
    image: `${IMG}/event-hub-access-policy.png`,
  },
  {
    id: 'connection-string',
    titleKey: `${ROOT}.sections.connectionString.title`,
    bodyKey: `${ROOT}.sections.connectionString.body`,
    image: `${IMG}/access-shared-policy.png`,
  },
  {
    id: 'consumer-group',
    titleKey: `${ROOT}.sections.consumerGroup.title`,
    bodyKey: `${ROOT}.sections.consumerGroup.body`,
    image: `${IMG}/consumer-group.png`,
  },
  {
    id: 'storage-account',
    titleKey: `${ROOT}.sections.storageAccount.title`,
    bodyKey: `${ROOT}.sections.storageAccount.body`,
    image: `${IMG}/account-storage.png`,
  },
  {
    id: 'storage-container',
    titleKey: `${ROOT}.sections.storageContainer.title`,
    bodyKey: `${ROOT}.sections.storageContainer.body`,
    image: `${IMG}/container-name.png`,
  },
  {
    id: 'event-grid',
    titleKey: `${ROOT}.sections.eventGrid.title`,
    bodyKey: `${ROOT}.sections.eventGrid.body`,
    image: `${IMG}/event-grid.png`,
  },
  {
    id: 'event-subscription',
    titleKey: `${ROOT}.sections.eventSubscription.title`,
    bodyKey: `${ROOT}.sections.eventSubscription.body`,
    image: `${IMG}/event-subs.png`,
  },
]

export const AZURE_FIELDS: CloudConfigField[] = [
  {
    key: 'eventHubConnection',
    labelKey: `${ROOT}.fields.eventHubConnection.label`,
    placeholder: 'Endpoint=sb://...;SharedAccessKeyName=...;SharedAccessKey=...;EntityPath=...',
    secret: true,
  },
  {
    key: 'consumerGroup',
    labelKey: `${ROOT}.fields.consumerGroup.label`,
    placeholder: 'utmstack-logstash',
  },
  {
    key: 'storageContainer',
    labelKey: `${ROOT}.fields.storageContainer.label`,
    placeholder: 'utmstack-events',
  },
  {
    key: 'storageConnection',
    labelKey: `${ROOT}.fields.storageConnection.label`,
    placeholder: 'DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...',
    secret: true,
  },
]
