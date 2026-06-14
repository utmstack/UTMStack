import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Check,
  Copy,
  Crown,
  Globe2,
  KeyRound,
  Loader2,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  Trash2,
  TriangleAlert,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { useBilling } from '@/features/billing'
import { ApiKeyHttpError, apiKeysHttpService } from '../services/api-keys-http.service'
import type { ApiKey, ApiKeyPageInfo } from '../types/api-key.types'

const PAGE_SIZE = 50
const EXPIRING_WINDOW_MS = 7 * 86_400_000

/* ─────────────────────────────────────────────────────────────────────────
 * Page
 * ───────────────────────────────────────────────────────────────────────── */

type DialogState = { mode: 'create' } | { mode: 'edit'; key: ApiKey } | null
type ConfirmState = { kind: 'rotate' | 'delete'; key: ApiKey } | null

export function ApiKeysSection() {
  const { t } = useTranslation()
  const { license } = useBilling()
  const [keys, setKeys] = useState<ApiKey[] | null>(null)
  const [pageInfo, setPageInfo] = useState<ApiKeyPageInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')

  const [dialog, setDialog] = useState<DialogState>(null)
  const [confirm, setConfirm] = useState<ConfirmState>(null)
  const [revealed, setRevealed] = useState<{ name: string; token: string } | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const resp = await apiKeysHttpService.list(page, PAGE_SIZE)
      // Go marshals an empty slice as null, so allowed_ip can come back null —
      // normalize it to an array so the rest of the page can treat it uniformly.
      const normalized = resp.data.map((k) => ({ ...k, allowed_ip: k.allowed_ip ?? [] }))
      setKeys((prev) => (page === 1 ? normalized : [...(prev ?? []), ...normalized]))
      setPageInfo(resp.page_info)
    } catch {
      setError(true)
      setKeys([])
    } finally {
      setLoading(false)
    }
  }, [page])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return (keys ?? []).filter((k) =>
      q ? (k.name + ' ' + k.allowed_ip.join(' ')).toLowerCase().includes(q) : true,
    )
  }, [keys, search])

  // API keys creation/rotation is an Enterprise capability. When gated, the
  // section still renders but shows an upgrade note instead of the manager.
  const isEnterprise = license?.edition === 'enterprise'
  const ENFORCE_ENTERPRISE_GATE = true
  const showManager = !ENFORCE_ENTERPRISE_GATE || isEnterprise

  return (
    <section className="mt-8 rounded-xl border border-border bg-card px-6 pb-6 pt-3">
      <header className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <KeyRound size={14} strokeWidth={1.75} />
          <span>
            <span className="font-medium text-foreground">{pageInfo?.total_items ?? 0}</span> {t('apiKeys.title').toLowerCase()}
          </span>
        </div>
        {showManager && (
          <Button size="sm" onClick={() => setDialog({ mode: 'create' })}>
            <Plus size={14} className="mr-1.5" />
            {t('apiKeys.new')}
          </Button>
        )}
      </header>

      {!showManager ? (
        <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border bg-muted/30 px-6 py-10 text-center">
          <Crown size={24} strokeWidth={1.75} className="text-amber-500" />
          <div className="text-sm font-medium">{t('apiKeys.enterprise.title')}</div>
          <p className="max-w-md text-xs text-muted-foreground">{t('apiKeys.enterprise.body')}</p>
          <Button asChild size="sm" className="mt-1">
            <Link to="/settings/license">
              <Crown size={14} className="mr-1.5" />
              {t('apiKeys.enterprise.upgrade')}
            </Link>
          </Button>
        </div>
      ) : (
        <>
          <div className="flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('apiKeys.searchPlaceholder')}
            className="h-9 pl-9"
          />
        </div>
        <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
          <RefreshCw size={14} className={cn('mr-1.5', loading && 'animate-spin')} />
          {t('apiKeys.refresh')}
        </Button>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: COLS }}
        >
          <div>{t('apiKeys.col.name')}</div>
          <div>{t('apiKeys.col.allowedIps')}</div>
          <div>{t('apiKeys.col.created')}</div>
          <div>{t('apiKeys.col.lastRotated')}</div>
          <div>{t('apiKeys.col.expires')}</div>
          <div>{t('apiKeys.col.status')}</div>
          <div className="text-right">{t('apiKeys.col.actions')}</div>
        </div>

        {loading && (!keys || keys.length === 0) && (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t('apiKeys.loading')}
          </div>
        )}

        {!loading && error && (
          <div className="flex flex-col items-center gap-3 px-6 py-12 text-sm">
            <span className="inline-flex items-center gap-2 text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t('apiKeys.loadFailed')}
            </span>
            <Button variant="outline" size="sm" onClick={() => void load()}>
              {t('apiKeys.retry')}
            </Button>
          </div>
        )}

        {!loading && !error && filtered.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            {search ? t('apiKeys.noSearchMatch') : t('apiKeys.empty')}
          </div>
        )}

        {filtered.length > 0 &&
          filtered.map((k) => (
            <KeyRow
              key={k.id}
              apiKey={k}
              onEdit={() => setDialog({ mode: 'edit', key: k })}
              onRotate={() => setConfirm({ kind: 'rotate', key: k })}
              onDelete={() => setConfirm({ kind: 'delete', key: k })}
            />
          ))}
      </div>

      {!error && keys && keys.length > 0 && (
        <InfiniteScrollSentinel
          onReach={() => setPage((p) => p + 1)}
          hasMore={keys.length < (pageInfo?.total_items ?? 0)}
          loading={loading}
          endLabel={t('common.allLoaded', { count: pageInfo?.total_items ?? 0 })}
        />
      )}
        </>
      )}

      {dialog && (
        <UpsertDialog
          state={dialog}
          onClose={() => setDialog(null)}
          onSaved={(reveal) => {
            setDialog(null)
            if (reveal) setRevealed(reveal)
            void load()
          }}
        />
      )}

      {confirm && (
        <ConfirmDialog
          state={confirm}
          onClose={() => setConfirm(null)}
          onDone={(reveal) => {
            setConfirm(null)
            if (reveal) setRevealed(reveal)
            void load()
          }}
        />
      )}

      {revealed && <RevealModal name={revealed.name} token={revealed.token} onClose={() => setRevealed(null)} />}
    </section>
  )
}

const COLS = '1.4fr 1.2fr 110px 110px 110px 100px 110px'

/* ─── Row ──────────────────────────────────────────────────────────────── */

function KeyRow({
  apiKey: k,
  onEdit,
  onRotate,
  onDelete,
}: {
  apiKey: ApiKey
  onEdit: () => void
  onRotate: () => void
  onDelete: () => void
}) {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-3 border-b border-border px-4 py-3 text-xs last:border-b-0 hover:bg-muted/30"
      style={{ gridTemplateColumns: COLS }}
    >
      <div className="flex min-w-0 items-center gap-2">
        <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
          <KeyRound size={13} />
        </span>
        <span className="truncate font-medium">{k.name}</span>
      </div>
      <div className="min-w-0 truncate text-[11px] text-muted-foreground">
        {k.allowed_ip.length === 0 ? (
          <span className="inline-flex items-center gap-1">
            <Globe2 size={11} /> {t('apiKeys.anyIp')}
          </span>
        ) : (
          <span className="font-mono" title={k.allowed_ip.join(', ')}>
            {k.allowed_ip.join(', ')}
          </span>
        )}
      </div>
      <div className="text-[11px] text-muted-foreground">{formatDate(k.created_at)}</div>
      <div className="text-[11px] text-muted-foreground">{formatDate(k.generated_at)}</div>
      <div className="text-[11px] text-muted-foreground">
        {k.expires_at ? formatDate(k.expires_at) : <span className="text-muted-foreground/60">{t('apiKeys.never')}</span>}
      </div>
      <div>
        <StatusBadge expiresAt={k.expires_at} />
      </div>
      <div className="flex items-center justify-end gap-1">
        <IconAction label={t('apiKeys.action.rotate')} onClick={onRotate}>
          <RefreshCw size={13} />
        </IconAction>
        <IconAction label={t('apiKeys.action.edit')} onClick={onEdit}>
          <Pencil size={13} />
        </IconAction>
        <IconAction label={t('apiKeys.action.delete')} danger onClick={onDelete}>
          <Trash2 size={13} />
        </IconAction>
      </div>
    </div>
  )
}

function IconAction({
  label,
  onClick,
  danger,
  children,
}: {
  label: string
  onClick: () => void
  danger?: boolean
  children: React.ReactNode
}) {
  return (
    <button
      title={label}
      aria-label={label}
      onClick={onClick}
      className={cn(
        'flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground',
        danger && 'hover:bg-red-500/10 hover:text-red-500',
      )}
    >
      {children}
    </button>
  )
}

function StatusBadge({ expiresAt }: { expiresAt: string | null }) {
  const { t } = useTranslation()
  const status = keyStatus(expiresAt)
  const meta = {
    active: { label: t('apiKeys.status.active'), cls: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
    expiring: { label: t('apiKeys.status.expiring'), cls: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
    expired: { label: t('apiKeys.status.expired'), cls: 'bg-red-500/15 text-red-500 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500' },
  }[status]
  return (
    <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', meta.cls)}>
      <span className={cn('h-1.5 w-1.5 rounded-full', meta.dot)} />
      {meta.label}
    </span>
  )
}

/* ─── Create / edit dialog ─────────────────────────────────────────────── */

function UpsertDialog({
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

/* ─── Confirm (rotate / delete) ────────────────────────────────────────── */

function ConfirmDialog({
  state,
  onClose,
  onDone,
}: {
  state: { kind: 'rotate' | 'delete'; key: ApiKey }
  onClose: () => void
  onDone: (reveal?: { name: string; token: string }) => void
}) {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)
  const isDelete = state.kind === 'delete'

  const run = async () => {
    setBusy(true)
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
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      onClose={onClose}
      title={isDelete ? t('apiKeys.confirm.deleteTitle') : t('apiKeys.confirm.rotateTitle')}
      icon={isDelete ? Trash2 : RefreshCw}
    >
      <div className="px-6 py-5 text-sm text-muted-foreground">
        {isDelete
          ? t('apiKeys.confirm.deleteBody', { name: state.key.name })
          : t('apiKeys.confirm.rotateBody', { name: state.key.name })}
      </div>
      <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
        <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
          {t('apiKeys.confirm.cancel')}
        </Button>
        <Button
          size="sm"
          variant={isDelete ? 'destructive' : 'default'}
          disabled={busy}
          onClick={() => void run()}
        >
          {busy ? t('apiKeys.confirm.working') : isDelete ? t('apiKeys.confirm.delete') : t('apiKeys.confirm.rotate')}
        </Button>
      </footer>
    </Modal>
  )
}

/* ─── Reveal token (shown once) ────────────────────────────────────────── */

function RevealModal({ name, token, onClose }: { name: string; token: string; onClose: () => void }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const copy = () => {
    navigator.clipboard.writeText(token).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }
  return (
    <Modal onClose={onClose} title={t('apiKeys.reveal.title')} icon={KeyRound}>
      <div className="space-y-4 px-6 py-5">
        <p className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
          <TriangleAlert size={14} className="mt-0.5 shrink-0" />
          {t('apiKeys.reveal.note', { name })}
        </p>
        <div className="flex items-stretch gap-2">
          <code className="min-w-0 flex-1 truncate rounded-md border border-border bg-muted/40 px-3 py-2 font-mono text-xs">
            {token}
          </code>
          <Button variant="outline" size="sm" onClick={copy}>
            {copied ? <Check size={14} className="text-emerald-500" /> : <Copy size={14} />}
            <span className="ml-1.5">{copied ? t('apiKeys.reveal.copied') : t('apiKeys.reveal.copy')}</span>
          </Button>
        </div>
      </div>
      <footer className="flex items-center justify-end border-t border-border px-6 py-3">
        <Button size="sm" onClick={onClose}>
          {t('apiKeys.reveal.done')}
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
  icon: typeof KeyRound
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
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

/* ─── helpers ──────────────────────────────────────────────────────────── */

function keyStatus(expiresAt: string | null): 'active' | 'expiring' | 'expired' {
  if (!expiresAt) return 'active'
  const exp = new Date(expiresAt).getTime()
  const now = Date.now()
  if (exp <= now) return 'expired'
  if (exp - now <= EXPIRING_WINDOW_MS) return 'expiring'
  return 'active'
}

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

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: '2-digit' })
}

function apiKeyError(err: unknown, t: TFunction): string {
  if (err instanceof ApiKeyHttpError) {
    if (err.status === 409) return t('apiKeys.error.nameTaken')
    if (err.status === 403) return t('apiKeys.error.enterpriseRequired')
    if (err.status === 404) return t('apiKeys.error.notFound')
    if (err.status === 400) return err.message || t('apiKeys.error.invalidRequest')
  }
  return err instanceof Error ? err.message : t('apiKeys.error.failed')
}
