import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { License, VersionInfo } from '../types/billing.types'

const api = createApiClient()

export { ApiError as BillingHttpError }

export const billingHttpService = {
  /** Deployment version + instance info. */
  getVersion: () => api.get<VersionInfo>('/billing/version'),
  /** The evaluated license / edition. */
  getLicense: () => api.get<License>('/billing/license'),
  /**
   * Upload/replace the signed LICENSE envelope (admin only). Returns the newly
   * evaluated license on success; 400 if the envelope is invalid/expired,
   * 403 if the caller is not an admin.
   */
  uploadLicense: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post<License>('/billing/license', fd)
  },
}
