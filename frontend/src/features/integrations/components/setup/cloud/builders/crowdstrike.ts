import type { CloudConfigField } from './cloudGuideBuilder'

const ROOT = 'integrations.setup.cloud.crowdstrike'

// The CrowdStrike step-by-step guide lives in CrowdStrikeGuide.tsx. Only the
// tenant config fields are shared here — they drive the CloudTenantForm. The
// keys MUST match the ConfKeys the crowdstrike plugin reads (see plugins/
// crowdstrike/main.go buildProcessor) and the backend verifier.
export const CROWDSTRIKE_FIELDS: CloudConfigField[] = [
  {
    key: 'crowdstrike_app_name',
    labelKey: `${ROOT}.fields.appName.label`,
    placeholder: 'utmstack-falcon',
  },
  {
    key: 'crowdstrike_client_id',
    labelKey: `${ROOT}.fields.clientId.label`,
    placeholder: 'a1b2c3d4e5f6g7h8i9j0...',
  },
  {
    key: 'crowdstrike_client_secret',
    labelKey: `${ROOT}.fields.clientSecret.label`,
    placeholder: '••••••••••••••••',
    secret: true,
  },
  {
    key: 'crowdstrike_cloud_region_url',
    labelKey: `${ROOT}.fields.cloudRegionUrl.label`,
    placeholder: 'https://api.crowdstrike.com',
  },
]
