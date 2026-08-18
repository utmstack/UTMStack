import type { TFunction } from 'i18next'
import { TenantsHttpError } from '../services/tenants-http.service'

export function tenantError(err: unknown, t: TFunction): string {
  if (err instanceof TenantsHttpError) {
    if (err.status === 409) return t('tenants.toast.domainInUse')
    if (err.status === 403) return t('tenants.toast.noPermission')
    if (err.status === 404) return t('tenants.toast.notFound')
    if (err.status === 400) return err.message || t('tenants.toast.invalidRequest')
  }
  return err instanceof Error ? err.message : t('tenants.toast.operationFailed')
}
