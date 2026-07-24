import type { TFunction } from 'i18next'
import { ApiKeyHttpError } from '../services/api-keys-http.service'

export function apiKeyError(err: unknown, t: TFunction): string {
  if (err instanceof ApiKeyHttpError) {
    if (err.status === 409) return t('apiKeys.error.nameTaken')
    if (err.status === 403) return t('apiKeys.error.enterpriseRequired')
    if (err.status === 404) return t('apiKeys.error.notFound')
    if (err.status === 400) return err.message || t('apiKeys.error.invalidRequest')
  }
  return err instanceof Error ? err.message : t('apiKeys.error.failed')
}
