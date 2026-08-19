import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { tenantsHttpService } from '../services/tenants-http.service'
import type { Tenant } from '../types/tenant.types'
import { Modal } from './Modal'
import { Field } from './Field'
import { tenantError } from './tenant-error'

export function EditTenantDialog({
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
