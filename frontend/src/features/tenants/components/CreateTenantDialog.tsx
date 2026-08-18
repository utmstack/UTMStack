import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Building2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { tenantsHttpService } from '../services/tenants-http.service'
import type { CreateTenantRequest } from '../types/tenant.types'
import { Modal } from './Modal'
import { Field } from './Field'
import { tenantError } from './tenant-error'

export function CreateTenantDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [domain, setDomain] = useState('')
  const [adminEmail, setAdminEmail] = useState('')
  const [busy, setBusy] = useState(false)

  const valid =
    name.trim().length >= 2 && domain.trim().length >= 3 && /.+@.+\..+/.test(adminEmail.trim())

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      const body: CreateTenantRequest = {
        name: name.trim(),
        domain: domain.trim().toLowerCase(),
        adminEmail: adminEmail.trim(),
      }
      await tenantsHttpService.create(body)
      toast.success(t('tenants.toast.created'))
      onCreated()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.create.title')}
      subtitle={t('tenants.create.subtitle')}
      icon={Building2}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.create.submit')}
          </Button>
        </>
      }
    >
      <Field label={t('tenants.fields.name')}>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Acme Corp" />
      </Field>
      <Field label={t('tenants.fields.domain')} hint={t('tenants.fields.domainHint')}>
        <Input
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
          className="font-mono"
          placeholder="acme.utmstack.com"
        />
      </Field>
      <Field label={t('tenants.fields.adminEmail')} hint={t('tenants.fields.adminEmailHint')}>
        <Input
          type="email"
          value={adminEmail}
          onChange={(e) => setAdminEmail(e.target.value)}
          placeholder="admin@acme.com"
        />
      </Field>
    </Modal>
  )
}
