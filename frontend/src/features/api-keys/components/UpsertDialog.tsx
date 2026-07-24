import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { KeyRound, TriangleAlert } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { apiKeysHttpService } from '../services/api-keys-http.service'
import type { ApiKey } from '../types/api-key.types'
import { Modal } from './Modal'
import { Field } from './Field'
import { apiKeyError } from './api-key-error'

function parseIps(s: string): string[] {
  return s
    .split(/[\s,]+/)
    .map((x) => x.trim())
    .filter(Boolean)
}

function toDateInput(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toISOString().slice(0, 10)
}

export function UpsertDialog({
  state,
  onClose,
  onSaved,
}: {
  state: { mode: 'create' } | { mode: 'edit'; key: ApiKey }
  onClose: () => void
  onSaved: (reveal?: { name: string; token: string }) => void
}) {
  const { t } = useTranslation()
  const editing = state.mode === 'edit'
  const existing = editing ? state.key : null

  const [name, setName] = useState(existing?.name ?? '')
  const [ips, setIps] = useState((existing?.allowed_ip ?? []).join(', '))
  const [expiresAt, setExpiresAt] = useState(existing?.expires_at ? toDateInput(existing.expires_at) : '')
  const [busy, setBusy] = useState(false)

  const valid = name.trim().length > 0

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    const payload = {
      name: name.trim(),
      allowed_ip: parseIps(ips),
      expires_at: expiresAt ? new Date(`${expiresAt}T23:59:59`).toISOString() : null,
    }
    try {
      if (editing && existing) {
        await apiKeysHttpService.update(existing.id, payload)
        toast.success(t('apiKeys.toast.updated'))
        onSaved()
      } else {
        const created = await apiKeysHttpService.create(payload)
        // The secret is not returned by create — reveal it via a rotate.
        try {
          const { api_key } = await apiKeysHttpService.generate(created.id)
          onSaved({ name: created.name, token: api_key })
        } catch {
          toast.warning(t('apiKeys.toast.secretIssueFailed'))
          onSaved()
        }
      }
    } catch (err) {
      toast.error(apiKeyError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal onClose={onClose} title={editing ? t('apiKeys.dialog.editTitle') : t('apiKeys.dialog.newTitle')} icon={KeyRound}>
      <div className="space-y-4 px-6 py-5">
        <Field label={t('apiKeys.dialog.name')}>
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t('apiKeys.dialog.namePlaceholder')} />
        </Field>
        <Field label={t('apiKeys.dialog.allowedIps')} hint={t('apiKeys.dialog.allowedIpsHint')}>
          <Input
            value={ips}
            onChange={(e) => setIps(e.target.value)}
            placeholder="203.0.113.4, 198.51.100.0/24"
            className="font-mono text-xs"
          />
        </Field>
        <Field label={t('apiKeys.dialog.expires')} hint={t('apiKeys.dialog.expiresHint')}>
          <Input type="date" value={expiresAt} onChange={(e) => setExpiresAt(e.target.value)} className="max-w-[200px]" />
        </Field>
        {!editing && (
          <p className="flex items-start gap-2 rounded-md border border-border bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
            <TriangleAlert size={13} className="mt-0.5 shrink-0 text-amber-500" />
            {t('apiKeys.dialog.secretNote')}
          </p>
        )}
      </div>
      <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
        <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
          {t('apiKeys.dialog.cancel')}
        </Button>
        <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
          {busy ? t('apiKeys.dialog.saving') : editing ? t('apiKeys.dialog.saveChanges') : t('apiKeys.dialog.createKey')}
        </Button>
      </footer>
    </Modal>
  )
}
