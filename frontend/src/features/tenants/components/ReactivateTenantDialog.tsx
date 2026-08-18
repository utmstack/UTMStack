import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Power } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { tenantsHttpService } from '../services/tenants-http.service'
import type { Tenant } from '../types/tenant.types'
import { Modal } from './Modal'
import { tenantError } from './tenant-error'

export function ReactivateTenantDialog({
  tenant,
  onClose,
  onSaved,
}: {
  tenant: Tenant
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)

  const submit = async () => {
    if (busy) return
    setBusy(true)
    try {
      await tenantsHttpService.reactivate(tenant.id)
      toast.success(t('tenants.toast.reactivated'))
      onSaved()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.reactivate.title')}
      icon={Power}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button size="sm" disabled={busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.reactivate.submit')}
          </Button>
        </>
      }
    >
      <p className="text-sm text-muted-foreground">
        {t('tenants.reactivate.body', { name: tenant.name })}
      </p>
    </Modal>
  )
}
