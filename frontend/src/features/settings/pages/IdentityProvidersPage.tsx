import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Check,
  Copy,
  Crown,
  Loader2,
  Pencil,
  Plus,
  ShieldCheck,
  Trash2,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useBilling } from '@/features/billing'
import { IdpHttpError, idpHttpService } from '../services/idp-http.service'
import type { IdentityProvider, IdentityProviderRequest } from '../types/idp.types'

/* ─── Enterprise gate (shown to Community installs) ────────────────────── */

function EnterpriseGate() {
  const { t } = useTranslation()
  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pb-6 pt-3">
      <PageHeader />
      <div className="mt-10 flex flex-col items-center justify-center rounded-xl border border-border bg-card px-6 py-16 text-center">
        <span className="flex h-14 w-14 items-center justify-center rounded-full bg-amber-500/10 text-amber-500 ring-1 ring-inset ring-amber-500/30">
          <Crown size={26} strokeWidth={1.75} />
        </span>
        <h2 className="mt-5 text-lg font-semibold">{t('idp.enterprise.title')}</h2>
        <p className="mt-2 max-w-md text-sm text-muted-foreground">{t('idp.enterprise.body')}</p>
        <Button asChild className="mt-6">
          <Link to="/settings/license">
            <Crown size={14} className="mr-1.5" />
            {t('idp.enterprise.upgrade')}
          </Link>
        </Button>
      </div>
    </div>
  )
}

function PageHeader({ onAdd }: { onAdd?: () => void }) {
  const { t } = useTranslation()
  return (
    <header className="flex items-end justify-between gap-3">
      <div>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <ShieldCheck size={16} strokeWidth={1.75} />
          {t('idp.title')}
        </h1>
      </div>
      {onAdd && (
        <Button size="sm" onClick={onAdd}>
          <Plus size={14} className="mr-1.5" />
          {t('idp.add')}
        </Button>
      )}
    </header>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Page
 * ───────────────────────────────────────────────────────────────────────── */

type DialogState = { mode: 'create' } | { mode: 'edit'; idp: IdentityProvider } | null

export function IdentityProvidersPage() {
  const { t } = useTranslation()
  const { license } = useBilling()
  const [items, setItems] = useState<IdentityProvider[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [dialog, setDialog] = useState<DialogState>(null)
  const [confirmDelete, setConfirmDelete] = useState<IdentityProvider | null>(null)
  const [togglingId, setTogglingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      setItems(await idpHttpService.list())
    } catch {
      setError(true)
      setItems([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const sorted = useMemo(
    () => [...(items ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
    [items],
  )

  // Enterprise-only feature gate. SAML SSO is a paid capability, so Community
  // installs should see an upgrade screen instead of the manager.
  //
  // TEMPORARILY DISABLED so the page can be worked without a license. Flip
  // ENFORCE_ENTERPRISE_GATE to `true` once licensing is wired:
  //   Community  → <EnterpriseGate/> ;  Enterprise → the manager below.
  const isEnterprise = license?.edition === 'enterprise'
  const ENFORCE_ENTERPRISE_GATE = true
  if (ENFORCE_ENTERPRISE_GATE && !isEnterprise) {
    return <EnterpriseGate />
  }

  const toggleActive = async (idp: IdentityProvider) => {
    setTogglingId(idp.id)
    try {
      await idpHttpService.update(toRequest(idp, { active: !idp.active }))
      toast.success(idp.active ? t('idp.toast.disabled') : t('idp.toast.enabled'))
      await load()
    } catch (err) {
      toast.error(idpError(err, t))
    } finally {
      setTogglingId(null)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pb-6 pt-3">
      <PageHeader onAdd={() => setDialog({ mode: 'create' })} />

      <div className="mt-5">
        {loading && (
          <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t('idp.loading')}
          </div>
        )}

        {!loading && error && (
          <div className="flex flex-col items-center gap-3 rounded-xl border border-border bg-card px-6 py-12 text-sm">
            <span className="inline-flex items-center gap-2 text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t('idp.loadFailed')}
            </span>
            <Button variant="outline" size="sm" onClick={() => void load()}>
              {t('idp.retry')}
            </Button>
          </div>
        )}

        {!loading && !error && sorted.length === 0 && (
          <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center">
            <ShieldCheck size={28} strokeWidth={1.5} className="text-muted-foreground/60" />
            <div className="text-sm font-medium">{t('idp.empty.title')}</div>
            <p className="max-w-sm text-xs text-muted-foreground">{t('idp.empty.body')}</p>
            <Button size="sm" className="mt-1" onClick={() => setDialog({ mode: 'create' })}>
              <Plus size={14} className="mr-1.5" />
              {t('idp.add')}
            </Button>
          </div>
        )}

        {!loading && !error && sorted.length > 0 && (
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            {sorted.map((idp) => (
              <ProviderCard
                key={idp.id}
                idp={idp}
                toggling={togglingId === idp.id}
                onToggle={() => void toggleActive(idp)}
                onEdit={() => setDialog({ mode: 'edit', idp })}
                onDelete={() => setConfirmDelete(idp)}
              />
            ))}
          </div>
        )}
      </div>

      {dialog && (
        <UpsertDialog
          state={dialog}
          onClose={() => setDialog(null)}
          onSaved={() => {
            setDialog(null)
            void load()
          }}
        />
      )}

      {confirmDelete && (
        <ConfirmDeleteDialog
          idp={confirmDelete}
          onClose={() => setConfirmDelete(null)}
          onDeleted={() => {
            setConfirmDelete(null)
            void load()
          }}
        />
      )}
    </div>
  )
}

/* ─── Provider card ────────────────────────────────────────────────────── */

function ProviderCard({
  idp,
  toggling,
  onToggle,
  onEdit,
  onDelete,
}: {
  idp: IdentityProvider
  toggling: boolean
  onToggle: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  const { t } = useTranslation()
  const loginUrl = `/api/v1/sso/saml/${idp.name}/login`
  return (
    <div className={cn('rounded-xl border border-border bg-card p-5', !idp.active && 'opacity-75')}>
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <ShieldCheck size={16} />
          </span>
          <div className="min-w-0">
            <div className="truncate text-sm font-semibold">{idp.name}</div>
            <div className="text-[11px] uppercase tracking-wide text-muted-foreground">
              {idp.providerType}
            </div>
          </div>
        </div>
        <button
          onClick={onToggle}
          disabled={toggling}
          title={idp.active ? t('idp.card.disable') : t('idp.card.enable')}
          className="disabled:opacity-50"
        >
          {idp.active ? (
            <span className="inline-flex items-center gap-1.5 rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600 ring-1 ring-inset ring-emerald-500/30 dark:text-emerald-300">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" /> {t('idp.card.active')}
            </span>
          ) : (
            <span className="inline-flex items-center gap-1.5 rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground ring-1 ring-inset ring-border">
              <span className="h-1.5 w-1.5 rounded-full bg-zinc-400" /> {t('idp.card.disabled')}
            </span>
          )}
        </button>
      </div>

      <dl className="mt-4 space-y-2 text-xs">
        <InfoRow label={t('idp.card.metadataUrl')} value={idp.metadataUrl} mono />
        <InfoRow label={t('idp.card.spEntityId')} value={idp.spEntityId} mono />
        <InfoRow label={t('idp.card.spAcsUrl')} value={idp.spAcsUrl} mono />
        <InfoRow label={t('idp.card.loginUrl')} value={loginUrl} mono copyable />
      </dl>

      <div className="mt-4 flex items-center justify-end gap-2 border-t border-border pt-3">
        <Button variant="outline" size="sm" onClick={onEdit}>
          <Pencil size={13} className="mr-1.5" />
          {t('idp.card.edit')}
        </Button>
        <Button variant="outline" size="sm" className="text-red-500 hover:bg-red-500/10" onClick={onDelete}>
          <Trash2 size={13} className="mr-1.5" />
          {t('idp.card.delete')}
        </Button>
      </div>
    </div>
  )
}

function InfoRow({
  label,
  value,
  mono,
  copyable,
}: {
  label: string
  value: string
  mono?: boolean
  copyable?: boolean
}) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  return (
    <div className="grid grid-cols-[110px_1fr] items-start gap-2">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="flex min-w-0 items-center gap-1.5">
        <span className={cn('min-w-0 truncate', mono && 'font-mono text-[11px]')} title={value}>
          {value || '—'}
        </span>
        {copyable && value && (
          <button
            onClick={() => {
              navigator.clipboard.writeText(value).then(() => {
                setCopied(true)
                setTimeout(() => setCopied(false), 1200)
              })
            }}
            className="shrink-0 rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
            title={t('idp.card.copy')}
          >
            {copied ? <Check size={12} className="text-emerald-500" /> : <Copy size={12} />}
          </button>
        )}
      </dd>
    </div>
  )
}

/* ─── Create / edit dialog ─────────────────────────────────────────────── */

function UpsertDialog({
  state,
  onClose,
  onSaved,
}: {
  state: { mode: 'create' } | { mode: 'edit'; idp: IdentityProvider }
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const editing = state.mode === 'edit'
  const existing = editing ? state.idp : null

  const [name, setName] = useState(existing?.name ?? '')
  const [providerType, setProviderType] = useState(existing?.providerType ?? '')
  const [metadataUrl, setMetadataUrl] = useState(existing?.metadataUrl ?? '')
  const [spEntityId, setSpEntityId] = useState(existing?.spEntityId ?? '')
  const [spAcsUrl, setSpAcsUrl] = useState(existing?.spAcsUrl ?? '')
  const [cert, setCert] = useState(existing?.spCertificatePem ?? '')
  const [privateKey, setPrivateKey] = useState('')
  const [active, setActive] = useState(existing?.active ?? true)
  const [busy, setBusy] = useState(false)

  const valid =
    !!name.trim() &&
    !!providerType.trim() &&
    !!metadataUrl.trim() &&
    !!spEntityId.trim() &&
    !!spAcsUrl.trim() &&
    !!cert.trim() &&
    (editing || !!privateKey.trim()) // private key required on create

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    const payload: IdentityProviderRequest = {
      id: existing?.id,
      name: name.trim(),
      providerType: providerType.trim(),
      metadataUrl: metadataUrl.trim(),
      spEntityId: spEntityId.trim(),
      spAcsUrl: spAcsUrl.trim(),
      spCertificatePem: cert.trim(),
      spPrivateKeyPem: privateKey.trim() || undefined,
      active,
    }
    try {
      if (editing) {
        await idpHttpService.update(payload)
        toast.success(t('idp.toast.updated'))
      } else {
        await idpHttpService.create(payload)
        toast.success(t('idp.toast.created'))
      }
      onSaved()
    } catch (err) {
      toast.error(idpError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal onClose={onClose} title={editing ? t('idp.form.editTitle') : t('idp.form.addTitle')} icon={ShieldCheck}>
      <div className="max-h-[70vh] space-y-4 overflow-y-auto px-6 py-5">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t('idp.form.name')} hint={t('idp.form.nameHint')}>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="acme-azure"
              disabled={editing}
              className={editing ? 'opacity-70' : ''}
            />
          </Field>
          <Field label={t('idp.form.providerType')} hint={t('idp.form.providerTypeHint')}>
            <Input value={providerType} onChange={(e) => setProviderType(e.target.value)} placeholder="azure" />
          </Field>
        </div>
        <Field label={t('idp.form.metadataUrl')}>
          <Input
            value={metadataUrl}
            onChange={(e) => setMetadataUrl(e.target.value)}
            placeholder="https://login.microsoftonline.com/…/federationmetadata.xml"
            className="font-mono text-xs"
          />
        </Field>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t('idp.form.spEntityId')}>
            <Input value={spEntityId} onChange={(e) => setSpEntityId(e.target.value)} placeholder="https://utmstack.example.com/saml" className="font-mono text-xs" />
          </Field>
          <Field label={t('idp.form.spAcsUrl')}>
            <Input value={spAcsUrl} onChange={(e) => setSpAcsUrl(e.target.value)} placeholder="https://…/api/v1/sso/saml/acme-azure/acs" className="font-mono text-xs" />
          </Field>
        </div>
        <Field label={t('idp.form.certificate')}>
          <Textarea value={cert} onChange={(e) => setCert(e.target.value)} placeholder="-----BEGIN CERTIFICATE-----" rows={4} />
        </Field>
        <Field
          label={t('idp.form.privateKey')}
          hint={editing ? t('idp.form.privateKeyHintEdit') : t('idp.form.privateKeyHintCreate')}
        >
          <Textarea
            value={privateKey}
            onChange={(e) => setPrivateKey(e.target.value)}
            placeholder={editing ? t('idp.form.privateKeyUnchanged') : '-----BEGIN PRIVATE KEY-----'}
            rows={4}
          />
        </Field>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} className="h-4 w-4 rounded border-input" />
          {t('idp.form.active')}
        </label>
      </div>
      <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
        <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
          {t('idp.form.cancel')}
        </Button>
        <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
          {busy ? t('idp.form.saving') : editing ? t('idp.form.save') : t('idp.form.create')}
        </Button>
      </footer>
    </Modal>
  )
}

/* ─── Delete confirm ───────────────────────────────────────────────────── */

function ConfirmDeleteDialog({
  idp,
  onClose,
  onDeleted,
}: {
  idp: IdentityProvider
  onClose: () => void
  onDeleted: () => void
}) {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)
  const run = async () => {
    setBusy(true)
    try {
      await idpHttpService.remove(idp.id)
      toast.success(t('idp.toast.deleted'))
      onDeleted()
    } catch (err) {
      toast.error(idpError(err, t))
    } finally {
      setBusy(false)
    }
  }
  return (
    <Modal onClose={onClose} title={t('idp.delete.title')} icon={Trash2}>
      <div className="px-6 py-5 text-sm text-muted-foreground">
        {t('idp.delete.body', { name: idp.name })}
      </div>
      <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
        <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
          {t('idp.delete.cancel')}
        </Button>
        <Button size="sm" variant="destructive" disabled={busy} onClick={() => void run()}>
          {busy ? t('idp.delete.deleting') : t('idp.delete.confirm')}
        </Button>
      </footer>
    </Modal>
  )
}

/* ─── Shared bits ──────────────────────────────────────────────────────── */

function Modal({
  title,
  icon: Icon,
  onClose,
  children,
}: {
  title: string
  icon: typeof ShieldCheck
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Icon size={17} strokeWidth={1.75} />
            {title}
          </h2>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>
        {children}
      </div>
    </div>
  )
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function Textarea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  const { className, ...rest } = props
  return (
    <textarea
      {...rest}
      className={cn(
        'flex w-full rounded-md border border-input bg-background/40 px-3 py-2 font-mono text-[11px] leading-relaxed',
        'placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
        className,
      )}
    />
  )
}

/* ─── helpers ──────────────────────────────────────────────────────────── */

function toRequest(idp: IdentityProvider, patch: Partial<IdentityProviderRequest>): IdentityProviderRequest {
  return {
    id: idp.id,
    name: idp.name,
    providerType: idp.providerType,
    metadataUrl: idp.metadataUrl,
    spEntityId: idp.spEntityId,
    spAcsUrl: idp.spAcsUrl,
    spCertificatePem: idp.spCertificatePem,
    // private key omitted → backend keeps the stored one
    active: idp.active,
    ...patch,
  }
}

function idpError(err: unknown, t: TFunction): string {
  if (err instanceof IdpHttpError) {
    if (err.status === 402) return t('idp.toast.paidPlan')
    if (err.status === 403) return t('idp.toast.enterpriseRequired')
    if (err.status === 404) return t('idp.toast.notFound')
    if (err.status === 400) return err.message || t('idp.toast.invalid')
  }
  return err instanceof Error ? err.message : t('idp.toast.failed')
}
