import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { tenantsHttpService } from '../services/tenants-http.service'
import type { Tenant, TenantStatus } from '../types/tenant.types'
import { Modal } from './Modal'
import { Field } from './Field'
import { tenantError } from './tenant-error'
import { ReactivateTenantDialog } from './ReactivateTenantDialog'

const STATUSES: TenantStatus[] = ['ACTIVE', 'SUSPENDED']

export function EditTenantDialog({
  tenant,
  onClose,
  onSaved,
}: {
  tenant: Tenant
  onClose: () => void
  onSaved: () => void
}) {
  if (tenant.status === 'TERMINATED') {
    return <ReactivateTenantDialog tenant={tenant} onClose={onClose} onSaved={onSaved} />
  }

  return <EditTenantForm tenant={tenant} onClose={onClose} onSaved={onSaved} />
}

function EditTenantForm({
  tenant,
  onClose,
  onSaved,
}: {
  tenant: Tenant
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState(tenant.name)
  const [domain, setDomain] = useState(tenant.domain)
  const [status, setStatus] = useState<TenantStatus>(tenant.status)
  const [capped, setCapped] = useState(tenant.limits.maxAIRequests != null)
  const [maxAI, setMaxAI] = useState(String(tenant.limits.maxAIRequests ?? ''))
  const [busy, setBusy] = useState(false)

  const parsedMax = Number(maxAI)
  const limitValid = !capped || (maxAI.trim() !== '' && Number.isInteger(parsedMax) && parsedMax >= 0)
  const valid = name.trim().length >= 2 && domain.trim().length >= 3 && limitValid

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      await tenantsHttpService.update(tenant.id, {
        name: name.trim(),
        domain: domain.trim().toLowerCase(),
        status,
        // Clearing the cap sends an explicit null: omitting the field would
        // mean "leave it as it is", which is the opposite of the intent.
        maxAIRequests: capped ? parsedMax : null,
      })
      toast.success(t('tenants.toast.updated'))
      onSaved()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.edit.title')}
      subtitle={t('tenants.edit.subtitle')}
      icon={Pencil}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.edit.submit')}
          </Button>
        </>
      }
    >
      <Field label={t('tenants.fields.name')}>
        <Input value={name} onChange={(e) => setName(e.target.value)} />
      </Field>
      <Field label={t('tenants.fields.domain')} hint={t('tenants.fields.domainHint')}>
        <Input value={domain} onChange={(e) => setDomain(e.target.value)} className="font-mono" />
      </Field>

      <Field label={t('tenants.fields.status')}>
        <div className="flex gap-2">
          {STATUSES.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => setStatus(s)}
              className={cn(
                'flex-1 rounded-md border px-3 py-2 text-xs font-medium transition-colors',
                status === s ? 'border-primary/40 bg-primary/5 text-foreground' : 'border-border text-muted-foreground hover:bg-muted/40'
              )}
            >
              {t(`tenants.status.${s}`, { defaultValue: s })}
            </button>
          ))}
        </div>
        {status === 'SUSPENDED' && (
          <p className="text-[11px] text-amber-600 dark:text-amber-300">
            {t('tenants.fields.suspendedWarning')}
          </p>
        )}
      </Field>

      <Field label={t('tenants.fields.aiLimit')} hint={t('tenants.fields.aiLimitHint')}>
        <button
          type="button"
          onClick={() => setCapped((c) => !c)}
          className="flex w-full items-center gap-2 rounded-md border border-border px-3 py-2 text-left text-xs hover:bg-muted/40"
        >
          <span
            className={cn(
              'flex h-4 w-4 shrink-0 items-center justify-center rounded border',
              capped ? 'border-primary bg-primary text-primary-foreground' : 'border-input'
            )}
          >
            {capped && <Check size={11} strokeWidth={3} />}
          </span>
          {t('tenants.fields.aiLimitToggle')}
        </button>
        {capped && (
          <Input
            type="number"
            min={0}
            value={maxAI}
            onChange={(e) => setMaxAI(e.target.value)}
            placeholder="1000"
            className="mt-2"
          />
        )}
      </Field>
    </Modal>
  )
}
