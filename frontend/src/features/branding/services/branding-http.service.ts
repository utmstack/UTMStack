import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { Branding, BrandingAssetSlot, BrandingPublic, BrandingRequest } from '../types/branding.types'

const api = createApiClient()

export { ApiError as BrandingHttpError }

export const brandingHttpService = {
  get: () => api.get<Branding>('/branding'),
  public: () => api.get<BrandingPublic>('/branding/public'),
  update: (input: BrandingRequest) => api.put<Branding>('/branding', input),
  // Uploads an image for a slot; the backend persists it immediately and returns
  // the updated branding with the new URL pointed in.
  uploadAsset: (slot: BrandingAssetSlot, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.post<Branding>(`/branding/assets/${slot}`, form)
  },
}
