import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Check,
  Copy,
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
import { EnterpriseGate } from '@/shared/components/EnterpriseGate'
import { PlatformBroadcastButton, broadcast, BULK_PATHS } from '@/features/platform-broadcast'
import { IdpHttpError, idpHttpService } from '../services/idp-http.service'
import type { GroupMapping, IdentityProvider, IdentityProviderRequest, ProviderType } from '../types/idp.types'
import { EMPTY_SETTINGS, PROVIDER_TYPES, REDIRECTING_PROVIDER_TYPES } from '../types/idp.types'
import { rolesHttpService, type RoleOption } from '../services/roles-http.service'

/* ─── Enterprise gate (shown to Community installs) ────────────────────── */

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
  const [togglingId, setTogglingId] = useState<string | null>(null)

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

  const isEnterprise = license?.edition === 'enterprise'
  if (!isEnterprise) {
    return (
      <EnterpriseGate
        header={<PageHeader />}
        title={t('idp.enterprise.title')}
        body={t('idp.enterprise.body')}
        cta={t('idp.enterprise.upgrade')}
      />
    )
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
    <div className="w-full px-6 pb-6 pt-3">
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
        {SUMMARY_FIELDS[idp.providerType].map((key) => (
          <InfoRow key={key} label={key} value={String(idp.settings?.[key] ?? '')} mono />
        ))}
        {idp.jitProvisioning && <InfoRow label={t('idp.card.jit')} value={t('idp.card.jitOn')} />}
        {idp.syncRolesOnLogin && <InfoRow label={t('idp.card.sync')} value={t('idp.card.syncOn')} />}
        {REDIRECTING_PROVIDER_TYPES.includes(idp.providerType) && (
          <InfoRow label={t('idp.card.loginUrl')} value={loginUrl} mono copyable />
        )}
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


/** What a card shows for each protocol: the fields that identify the provider,
 * not everything it stores. */
const SUMMARY_FIELDS: Record<ProviderType, string[]> = {
  saml: ['metadataUrl', 'spEntityId'],
  oidc: ['issuer', 'clientId'],
  ldap: ['host', 'baseDn'],
}

const REQUIRED_FIELDS: Record<ProviderType, string[]> = {
  saml: ['metadataUrl', 'spEntityId', 'spAcsUrl', 'spCertificatePem'],
  oidc: ['issuer', 'clientId', 'redirectUrl'],
  ldap: ['host', 'bindDn', 'baseDn', 'userFilter'],
}

/** Each protocol carries exactly one secret, and it is write-only. */
const SECRET_FIELD: Record<ProviderType, string> = {
  saml: 'spPrivateKeyPem',
  oidc: 'clientSecret',
  ldap: 'bindPassword',
}

function Select(props: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      {...props}
      className="flex h-10 w-full cursor-pointer rounded-md border border-input bg-background/40 px-3 py-2 text-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
    />
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
  const [providerType, setProviderType] = useState<ProviderType>(existing?.providerType ?? 'saml')
  const [active, setActive] = useState(existing?.active ?? true)

  // The stored settings come back without their secret, so an edit starts from
  // what is known and leaves the secret blank to mean "keep it".
  const [settings, setSettings] = useState<Record<string, unknown>>(
    existing?.settings ?? { ...EMPTY_SETTINGS[existing?.providerType ?? 'saml'] },
  )
  const [secret, setSecret] = useState('')

  const [jit, setJit] = useState(existing?.jitProvisioning ?? false)
  const [defaultRoleId, setDefaultRoleId] = useState(existing?.defaultRoleId ?? '')
  const [groupsAttribute, setGroupsAttribute] = useState(existing?.groupsAttribute ?? '')
  const [syncRoles, setSyncRoles] = useState(existing?.syncRolesOnLogin ?? false)
  const [mappings, setMappings] = useState<GroupMapping[]>([])

  const [roles, setRoles] = useState<RoleOption[]>([])
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    void rolesHttpService.list().then(setRoles).catch(() => setRoles([]))
    if (existing) {
      void idpHttpService.mappings(existing.id).then(setMappings).catch(() => setMappings([]))
    }
  }, [existing])

  // Switching protocol replaces the whole shape: the fields of the previous one
  // mean nothing to the next.
  const changeType = (next: ProviderType) => {
    setProviderType(next)
    setSettings({ ...EMPTY_SETTINGS[next] })
    setSecret('')
  }

  const set = (key: string, value: unknown) => setSettings((s) => ({ ...s, [key]: value }))
  const str = (key: string) => String(settings[key] ?? '')

  const required = REQUIRED_FIELDS[providerType]
  const valid =
    !!name.trim() &&
    required.every((k) => String(settings[k] ?? '').trim() !== '') &&
    (editing || !!secret.trim()) &&
    (providerType !== 'ldap' || str('userFilter').includes('%s'))

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    const payload: IdentityProviderRequest = {
      id: existing?.id,
      name: name.trim(),
      providerType,
      active,
      settings: { ...settings, [SECRET_FIELD[providerType]]: secret.trim() || undefined } as never,
      jitProvisioning: jit,
      defaultRoleId: defaultRoleId || null,
      groupsAttribute: groupsAttribute.trim(),
      syncRolesOnLogin: syncRoles,
      groupMappings: mappings.filter((m) => m.group.trim() && m.roleId),
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
              placeholder="acme-entra"
              disabled={editing}
              className={editing ? 'opacity-70' : ''}
            />
          </Field>
          <Field label={t('idp.form.providerType')} hint={t('idp.form.providerTypeHint')}>
            <Select
              value={providerType}
              onChange={(e) => changeType(e.target.value as ProviderType)}
              disabled={editing}
            >
              {PROVIDER_TYPES.map((p) => (
                <option key={p} value={p}>
                  {t(`idp.providerType.${p}`)}
                </option>
              ))}
            </Select>
          </Field>
        </div>

        {providerType === 'saml' && (
          <>
            <Field label={t('idp.form.metadataUrl')}>
              <Input value={str('metadataUrl')} onChange={(e) => set('metadataUrl', e.target.value)} className="font-mono text-xs" />
            </Field>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('idp.form.spEntityId')}>
                <Input value={str('spEntityId')} onChange={(e) => set('spEntityId', e.target.value)} className="font-mono text-xs" />
              </Field>
              <Field label={t('idp.form.spAcsUrl')}>
                <Input value={str('spAcsUrl')} onChange={(e) => set('spAcsUrl', e.target.value)} className="font-mono text-xs" />
              </Field>
            </div>
            <Field label={t('idp.form.certificate')}>
              <Textarea value={str('spCertificatePem')} onChange={(e) => set('spCertificatePem', e.target.value)} rows={4} />
            </Field>
          </>
        )}

        {providerType === 'oidc' && (
          <>
            <Field label={t('idp.form.issuer')} hint={t('idp.form.issuerHint')}>
              <Input value={str('issuer')} onChange={(e) => set('issuer', e.target.value)} placeholder="https://login.microsoftonline.com/<tenant>/v2.0" className="font-mono text-xs" />
            </Field>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('idp.form.clientId')}>
                <Input value={str('clientId')} onChange={(e) => set('clientId', e.target.value)} className="font-mono text-xs" />
              </Field>
              <Field label={t('idp.form.redirectUrl')} hint={t('idp.form.redirectUrlHint')}>
                <Input value={str('redirectUrl')} onChange={(e) => set('redirectUrl', e.target.value)} className="font-mono text-xs" />
              </Field>
            </div>
          </>
        )}

        {providerType === 'ldap' && (
          <>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <Field label={t('idp.form.host')}>
                <Input value={str('host')} onChange={(e) => set('host', e.target.value)} className="font-mono text-xs" />
              </Field>
              <Field label={t('idp.form.port')}>
                <Input type="number" value={str('port')} onChange={(e) => set('port', Number(e.target.value))} />
              </Field>
              <Field label={t('idp.form.startTls')}>
                <label className="flex h-10 items-center gap-2 text-sm">
                  <input type="checkbox" checked={Boolean(settings.startTls)} onChange={(e) => set('startTls', e.target.checked)} className="h-4 w-4 rounded border-input" />
                  {t('idp.form.startTlsOn')}
                </label>
              </Field>
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('idp.form.bindDn')} hint={t('idp.form.bindDnHint')}>
                <Input value={str('bindDn')} onChange={(e) => set('bindDn', e.target.value)} className="font-mono text-xs" />
              </Field>
              <Field label={t('idp.form.baseDn')}>
                <Input value={str('baseDn')} onChange={(e) => set('baseDn', e.target.value)} className="font-mono text-xs" />
              </Field>
            </div>
            <Field label={t('idp.form.userFilter')} hint={t('idp.form.userFilterHint')}>
              <Input value={str('userFilter')} onChange={(e) => set('userFilter', e.target.value)} className="font-mono text-xs" />
            </Field>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <Field label={t('idp.form.emailAttribute')}>
                <Input value={str('emailAttribute')} onChange={(e) => set('emailAttribute', e.target.value)} className="font-mono text-xs" />
              </Field>
              <Field label={t('idp.form.nameAttribute')}>
                <Input value={str('nameAttribute')} onChange={(e) => set('nameAttribute', e.target.value)} className="font-mono text-xs" />
              </Field>
              <Field label={t('idp.form.groupAttribute')}>
                <Input value={str('groupAttribute')} onChange={(e) => set('groupAttribute', e.target.value)} className="font-mono text-xs" />
              </Field>
            </div>
          </>
        )}

        <Field
          label={t(`idp.form.secret.${providerType}`)}
          hint={editing ? t('idp.form.secretHintEdit') : t('idp.form.secretHintCreate')}
        >
          {providerType === 'saml' ? (
            <Textarea value={secret} onChange={(e) => setSecret(e.target.value)} rows={4} placeholder={editing ? t('idp.form.secretUnchanged') : '-----BEGIN PRIVATE KEY-----'} />
          ) : (
            <Input type="password" value={secret} onChange={(e) => setSecret(e.target.value)} placeholder={editing ? t('idp.form.secretUnchanged') : ''} />
          )}
        </Field>

        <div className="rounded-lg border border-border bg-muted/20 p-4">
          <div className="text-sm font-medium">{t('idp.form.provisioning')}</div>
          <p className="mt-1 text-xs text-muted-foreground">{t('idp.form.provisioningHint')}</p>

          <label className="mt-3 flex items-center gap-2 text-sm">
            <input type="checkbox" checked={jit} onChange={(e) => setJit(e.target.checked)} className="h-4 w-4 rounded border-input" />
            {t('idp.form.jit')}
          </label>

          <div className="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label={t('idp.form.defaultRole')} hint={t('idp.form.defaultRoleHint')}>
              <Select value={defaultRoleId ?? ''} onChange={(e) => setDefaultRoleId(e.target.value)}>
                <option value="">{t('idp.form.noDefaultRole')}</option>
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.display_name || r.name}
                  </option>
                ))}
              </Select>
            </Field>
            <Field label={t('idp.form.groupsAttribute')} hint={t('idp.form.groupsAttributeHint')}>
              <Input value={groupsAttribute} onChange={(e) => setGroupsAttribute(e.target.value)} placeholder={providerType === 'ldap' ? 'memberOf' : 'groups'} className="font-mono text-xs" />
            </Field>
          </div>

          <label className="mt-3 flex items-start gap-2 text-sm">
            <input type="checkbox" checked={syncRoles} onChange={(e) => setSyncRoles(e.target.checked)} className="mt-0.5 h-4 w-4 rounded border-input" />
            <span>
              {t('idp.form.syncRoles')}
              <span className="mt-0.5 block text-xs text-muted-foreground">{t('idp.form.syncRolesHint')}</span>
            </span>
          </label>

          <div className="mt-4">
            <div className="flex items-center justify-between">
              <div className="text-sm font-medium">{t('idp.form.mappings')}</div>
              <Button variant="outline" size="sm" onClick={() => setMappings((m) => [...m, { group: '', roleId: '' }])}>
                <Plus size={13} className="mr-1.5" />
                {t('idp.form.addMapping')}
              </Button>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">{t('idp.form.mappingsHint')}</p>

            {mappings.length === 0 ? (
              <p className="mt-3 text-xs text-muted-foreground">{t('idp.form.noMappings')}</p>
            ) : (
              <div className="mt-3 space-y-2">
                {mappings.map((m, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <Input
                      value={m.group}
                      onChange={(e) =>
                        setMappings((rows) => rows.map((r, j) => (j === i ? { ...r, group: e.target.value } : r)))
                      }
                      placeholder={t('idp.form.groupPlaceholder')}
                      className="font-mono text-xs"
                    />
                    <Select
                      value={m.roleId}
                      onChange={(e) =>
                        setMappings((rows) => rows.map((r, j) => (j === i ? { ...r, roleId: e.target.value } : r)))
                      }
                    >
                      <option value="">{t('idp.form.pickRole')}</option>
                      {roles.map((r) => (
                        <option key={r.id} value={r.id}>
                          {r.display_name || r.name}
                        </option>
                      ))}
                    </Select>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-500 hover:bg-red-500/10"
                      onClick={() => setMappings((rows) => rows.filter((_, j) => j !== i))}
                    >
                      <Trash2 size={13} />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} className="h-4 w-4 rounded border-input" />
          {t('idp.form.active')}
        </label>
      </div>
      <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('idp.form.cancel')}
          </Button>
          <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('idp.form.saving') : editing ? t('idp.form.save') : t('idp.form.create')}
          </Button>
        </div>
        {valid && (
          <PlatformBroadcastButton
            label={t('platformBroadcast.button')}
            title={t(editing ? 'platformBroadcast.action.update' : 'platformBroadcast.action.create', { resource: t('platformBroadcast.resource.idp') })}
            disabled={busy}
            onBroadcast={async (selector) => {
              const payload: IdentityProviderRequest = {
                id: existing?.id,
                name: name.trim(),
                providerType,
                active,
                settings: { ...settings, [SECRET_FIELD[providerType]]: secret.trim() || undefined } as never,
                jitProvisioning: jit,
                defaultRoleId: defaultRoleId || null,
                groupsAttribute: groupsAttribute.trim(),
                syncRolesOnLogin: syncRoles,
                groupMappings: mappings.filter((m) => m.group.trim() && m.roleId),
              }
              return broadcast(editing ? BULK_PATHS.idp.update : BULK_PATHS.idp.create, selector, payload)
            }}
          />
        )}
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
      <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('idp.delete.cancel')}
          </Button>
          <Button size="sm" variant="destructive" disabled={busy} onClick={() => void run()}>
            {busy ? t('idp.delete.deleting') : t('idp.delete.confirm')}
          </Button>
        </div>
        <PlatformBroadcastButton
          label={t('platformBroadcast.button')}
          title={t('platformBroadcast.action.delete', { resource: t('platformBroadcast.resource.idp') })}
          disabled={busy}
          onBroadcast={async (selector) => {
            return broadcast(BULK_PATHS.idp.delete, selector, { id: idp.id })
          }}
        />
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
    // The settings come back without their secret, and sending them that way is
    // what tells the backend to keep the stored one.
    settings: idp.settings as unknown as IdentityProviderRequest['settings'],
    jitProvisioning: idp.jitProvisioning,
    defaultRoleId: idp.defaultRoleId,
    groupsAttribute: idp.groupsAttribute,
    syncRolesOnLogin: idp.syncRolesOnLogin,
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
