import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Loader2, Plus, Trash2, UserCircle2, X } from 'lucide-react'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useUserAuditor } from '../context/user-auditor.context'
import type { IdentityListItem } from '../types/ad-user.types'

/**
 * Dropdown to assign a user to an existing identity group or create a new one.
 * Mirrors IncidentAssigneePicker from the incidents drawer, backed by the
 * shared user-auditor context (/ad-audit/identities).
 */
export function IdentityGroupPicker({
  userId,
  current,
  onChanged,
}: {
  userId: string
  current: string | null
  onChanged: (name: string | null) => void
}) {
  const { t } = useTranslation()
  const { identities, identitiesLoaded, assignIdentity, createIdentity, deleteIdentity } = useUserAuditor()
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [custom, setCustom] = useState('')
  const [busy, setBusy] = useState(false)
  const [deleting, setDeleting] = useState<IdentityListItem | null>(null)
  const [deleteBusy, setDeleteBusy] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const assign = async (id: string) => {
    setBusy(true)
    try {
      await assignIdentity(userId, id)
      const name = id ? identities.find((i) => i.id === id)?.name ?? null : null
      setOpen(false)
      setCustom('')
      onChanged(name)
      toast.success(t('userAuditor.drawer.alias.updated'))
    } catch {
      toast.error(t('userAuditor.drawer.alias.updateFailed'))
    } finally {
      setBusy(false)
    }
  }

  const createAndAssign = async () => {
    const name = custom.trim()
    if (name.length < 2) {
      toast.error(t('userAuditor.drawer.alias.nameTooShort'))
      return
    }
    setBusy(true)
    try {
      const res = await createIdentity(name)
      if (!res) throw new Error('no identity')
      await assignIdentity(userId, res.id)
      setOpen(false)
      setCustom('')
      onChanged(name)
      toast.success(t('userAuditor.drawer.alias.updated'))
    } catch {
      toast.error(t('userAuditor.drawer.alias.updateFailed'))
    } finally {
      setBusy(false)
    }
  }

  const confirmDelete = async () => {
    if (!deleting) return
    setDeleteBusy(true)
    try {
      await deleteIdentity(deleting.id)
      if (deleting.name === current) onChanged(null)
      toast.success(t('userAuditor.drawer.alias.deleted'))
    } catch {
      toast.error(t('userAuditor.drawer.alias.deleteFailed'))
    } finally {
      setDeleteBusy(false)
      setDeleting(null)
    }
  }

  const customTrimmed = custom.trim()
  const canCreateCustom = customTrimmed.length >= 2 && customTrimmed !== current
  const filtered = identities.filter((i) => (q ? i.name.toLowerCase().includes(q.toLowerCase()) : true))

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <UserCircle2 size={13} />
        <span className="max-w-[160px] truncate">{current || t('userAuditor.drawer.alias.add')}</span>
        <ChevronDown size={12} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute  top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              autoFocus
              placeholder={t('userAuditor.drawer.alias.search')}
              className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          {current && (
            <button
              onClick={() => void assign('')}
              className="flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-xs text-muted-foreground hover:bg-muted"
            >
              <X size={12} /> {t('userAuditor.drawer.alias.unassign')}
            </button>
          )}
          <div className="max-h-56 overflow-y-auto">
            {!identitiesLoaded && (
              <div className="flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              </div>
            )}
            {identitiesLoaded && filtered.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('userAuditor.drawer.alias.none')}</div>
            )}
            {filtered.map((i) => (
              <div
                key={i.id}
                className="group flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted"
              >
                <button
                  onClick={() => void assign(i.id)}
                  disabled={busy}
                  className="flex min-w-0 flex-1 items-center justify-between gap-2 text-left disabled:opacity-50"
                >
                  <span className="truncate">
                    {i.name} <span className="font-mono text-[10px] text-muted-foreground">{i.count}</span>
                  </span>
                  {i.name === current && <Check size={14} className="shrink-0 text-primary" />}
                </button>
                <button
                  onClick={() => setDeleting(i)}
                  disabled={busy || deleteBusy || i.count > 0}
                  title={i.count > 0 ? undefined : t('userAuditor.drawer.alias.delete')}
                  className={
                    i.count > 0
                      ? 'shrink-0 cursor-not-allowed text-muted-foreground opacity-0 transition-opacity group-hover:opacity-30'
                      : 'shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-red-500 group-hover:opacity-100'
                  }
                >
                  <Trash2 size={13} />
                </button>
              </div>
            ))}
          </div>
          <div className="mt-1 border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t('userAuditor.drawer.alias.new')}
            </div>
            <div className="flex items-center gap-1.5">
              <input
                value={custom}
                onChange={(e) => setCustom(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && void createAndAssign()}
                placeholder={t('userAuditor.drawer.alias.newPlaceholder')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                onClick={() => void createAndAssign()}
                disabled={!canCreateCustom || busy}
                title={t('userAuditor.drawer.alias.createAndAssign')}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
              >
                <Plus size={14} />
              </button>
            </div>
          </div>
        </div>
      )}
      <ConfirmDialog
        open={deleting !== null}
        title={t('userAuditor.drawer.alias.deleteTitle')}
        body={t('userAuditor.drawer.alias.deleteBody', { name: deleting?.name ?? '', count: deleting?.count ?? 0 })}
        confirmLabel={t('common.confirm.delete')}
        cancelLabel={t('common.cancel')}
        danger
        busy={deleteBusy}
        onClose={() => setDeleting(null)}
        onConfirm={confirmDelete}
      />
    </div>
  )
}
