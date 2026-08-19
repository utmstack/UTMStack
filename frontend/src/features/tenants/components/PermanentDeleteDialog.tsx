import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { tenantsHttpService } from '../services/tenants-http.service'
import type { Tenant } from '../types/tenant.types'
import { Modal } from './Modal'
import { Field } from './Field'
import { tenantError } from './tenant-error'

export function PermanentDeleteDialog({
  tenant,
  onClose,
  onDone,
}: {
  tenant: Tenant
  onClose: () => void
  onDone: () => void
}) {
  const { t } = useTranslation()
  const [confirmation, setConfirmation] = useState('')
  const [busy, setBusy] = useState(false)

  const valid = confirmation.trim() === tenant.name

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      await tenantsHttpService.permanentlyDelete(tenant.id)
      toast.success(t('tenants.toast.permanentlyDeleted'))
      onDone()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.deletePermanent.title')}
      icon={Trash2}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button variant="destructive" size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.deletePermanent.submit')}
          </Button>
        </>
      }
    >
      <p className="text-sm text-muted-foreground">
        {t('tenants.deletePermanent.body', { name: tenant.name })}
      </p>
      <Field label={t('tenants.deletePermanent.confirmLabel', { name: tenant.name })}>
        <Input
          value={confirmation}
          onChange={(e) => setConfirmation(e.target.value)}
          placeholder={tenant.name}
        />
      </Field>
    </Modal>
  )
}
