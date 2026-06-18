import type { CloudConfigField } from './cloudGuideBuilder'

const ROOT = 'integrations.setup.cloud.azure'

// The Azure step-by-step guide lives in AzureGuide.tsx (dedicated component). Only
// the tenant config fields are shared here — they drive the CloudTenantForm.
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
