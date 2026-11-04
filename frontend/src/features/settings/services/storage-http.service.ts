import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  DatasetUsage,
  ColdTargetInput,
  Retention,
  StoreHealth,
  Tiering,
} from '../types/storage.types'

const api = createApiClient()

export { ApiError as StorageHttpError }

/**
 * Retention belongs to the tables, which every tenant shares, so these are
 * platform-only: a tenant admin gets 403.
 */
export const storageHttpService = {
  retention: () => api.get<Retention[]>('/storage/retention'),
  setRetention: (r: Pick<Retention, 'dataset' | 'keepDays' | 'coldDays'>) =>
    api.put<Retention>('/storage/retention', r),

  usage: () => api.get<DatasetUsage[]>('/storage/usage'),
  health: () => api.get<StoreHealth>('/storage/health'),

  tiering: () => api.get<Tiering>('/storage/tiering'),
  enableTiering: (input: ColdTargetInput) => api.put<Tiering>('/storage/tiering', input),
}
