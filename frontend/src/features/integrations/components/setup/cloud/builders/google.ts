import type { CloudConfigField } from './cloudGuideBuilder'

const ROOT = 'integrations.setup.cloud.google'

// The GCP step-by-step guide lives in GcpGuide.tsx (dedicated component). Only the
// tenant config fields are shared here — they drive the CloudTenantForm. The keys
// match what the gcp plugin reads (plugins/gcp/main.go getModuleConfig) and the
// backend verifier (verifier_gcp.go): jsonKey, projectId, subscription. The Pub/Sub
// topic is created in the GCP console but is NOT stored — UTMStack pulls from the
// subscription only.
export const GOOGLE_FIELDS: CloudConfigField[] = [
  {
    key: 'projectId',
    labelKey: `${ROOT}.fields.projectId.label`,
    placeholder: 'my-gcp-project',
  },
  {
    key: 'subscription',
    labelKey: `${ROOT}.fields.subscription.label`,
    placeholder: 'utmstack-subscription',
  },
  {
    key: 'jsonKey',
    labelKey: `${ROOT}.fields.jsonKey.label`,
    type: 'file',
    accept: 'application/json,.json',
  },
]
