import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Loader2, Plus, UserCircle2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { adAuditHttpService } from '../services/ad-audit-http.service'
import type { IdentityListItem } from '../types/ad-user.types'

/**
 * Dropdown to assign a user to an existing identity group or create a new one.
 * Mirrors IncidentAssigneePicker from the incidents drawer, backed by /ad-audit/identities.
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
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [custom, setCustom] = useState('')
  const [items, setItems] = useState<IdentityListItem[] | null>(null)
  const [busy, setBusy] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  // Load the alias list on open; search below filters client-side.
  useEffect(() => {
    if (!open) return
    let cancelled = false
    void adAuditHttpService
      .identities()
      .then((res) => {
        if (!cancelled) setItems(res ?? [])
      })
      .catch(() => {
        if (!cancelled) setItems([])
      })
    return () => {
      cancelled = true
    }
  }, [open])

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const assign = (id: string) => {
    setBusy(true)
    void adAuditHttpService
      .assignIdentity(userId, id)
      .then(() => {
        const name = id ? items?.find((i) => i.id === id)?.name ?? null : null
        setOpen(false)
        setCustom('')
        onChanged(name)
        toast.success(t('userAuditor.drawer.alias.updated'))
      })
      .catch(() => toast.error(t('userAuditor.drawer.alias.updateFailed')))
      .finally(() => setBusy(false))
  }

  const createAndAssign = () => {
    const name = custom.trim()
    if (name.length < 2) {
      toast.error(t('userAuditor.drawer.alias.nameTooShort'))
      return
    }
    setBusy(true)
    void adAuditHttpService
      .createIdentity(name)
      .then((res) => {
        if (!res) throw new Error('no identity')
        return adAuditHttpService.assignIdentity(userId, res.id)
      })
      .then(() => {
        setOpen(false)
        setCustom('')
        onChanged(name)
        toast.success(t('userAuditor.drawer.alias.updated'))
      })
      .catch(() => toast.error(t('userAuditor.drawer.alias.updateFailed')))
      .finally(() => setBusy(false))
  }

  const customTrimmed = custom.trim()
  const canCreateCustom = customTrimmed.length >= 2 && customTrimmed !== current
  const filtered = (items ?? []).filter((i) => (q ? i.name.toLowerCase().includes(q.toLowerCase()) : true))

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
        <div className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
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
              onClick={() => assign('')}
              className="flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-xs text-muted-foreground hover:bg-muted"
            >
              <X size={12} /> {t('userAuditor.drawer.alias.unassign')}
            </button>
          )}
          <div className="max-h-56 overflow-y-auto">
            {items === null && (
              <div className="flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              </div>
            )}
            {items !== null && filtered.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('userAuditor.drawer.alias.none')}</div>
            )}
            {filtered.map((i) => (
              <button
                key={i.id}
                onClick={() => assign(i.id)}
                disabled={busy}
                className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted"
              >
                <span className="truncate">
                  {i.name} <span className="font-mono text-[10px] text-muted-foreground">{i.count}</span>
                </span>
                {i.name === current && <Check size={14} className="shrink-0 text-primary" />}
              </button>
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
                onKeyDown={(e) => e.key === 'Enter' && createAndAssign()}
                placeholder={t('userAuditor.drawer.alias.newPlaceholder')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                onClick={createAndAssign}
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
    </div>
  )
}
