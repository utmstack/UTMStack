import { useTranslation } from 'react-i18next'
import { RefreshCw, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { ConfirmDialog as SharedConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { apiKeysHttpService } from '../services/api-keys-http.service'
import type { ApiKey } from '../types/api-key.types'
import { apiKeyError } from './api-key-error'

export function ConfirmDialog({
  state,
  onClose,
  onDone,
}: {
  state: { kind: 'rotate' | 'delete'; key: ApiKey }
  onClose: () => void
  onDone: (reveal?: { name: string; token: string }) => void
}) {
  const { t } = useTranslation()
  const isDelete = state.kind === 'delete'

  const run = async () => {
    try {
      if (isDelete) {
        await apiKeysHttpService.remove(state.key.id)
        toast.success(t('apiKeys.toast.deleted'))
        onDone()
      } else {
        const { api_key } = await apiKeysHttpService.generate(state.key.id)
        onDone({ name: state.key.name, token: api_key })
      }
    } catch (err) {
      toast.error(apiKeyError(err, t))
    }
  }

  return (
    <SharedConfirmDialog
      open
      icon={isDelete ? Trash2 : RefreshCw}
      title={isDelete ? t('apiKeys.confirm.deleteTitle') : t('apiKeys.confirm.rotateTitle')}
      body={
        isDelete
          ? t('apiKeys.confirm.deleteBody', { name: state.key.name })
          : t('apiKeys.confirm.rotateBody', { name: state.key.name })
      }
      confirmLabel={isDelete ? t('apiKeys.confirm.delete') : t('apiKeys.confirm.rotate')}
      cancelLabel={t('apiKeys.confirm.cancel')}
      danger={isDelete}
      onClose={onClose}
      onConfirm={run}
    />
  )
}
