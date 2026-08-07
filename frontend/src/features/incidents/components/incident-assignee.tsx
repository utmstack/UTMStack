import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Loader2, Plus, UserCircle2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useIncidentAssignment } from '../hooks/use-incident-assignment'
import type { Incident } from '../types/incident.types'

export function IncidentAssignee({ login }: { login?: string }) {
  const { t } = useTranslation()
  if (!login)
    return (
      <span className="inline-flex items-center gap-1 text-muted-foreground/70">
        <UserCircle2 size={12} /> {t('incidents.unassigned')}
      </span>
    )
  return (
    // flex, not inline-flex: an inline box sizes to its content, so a long
    // address ran past the column and over the one beside it instead of
    // truncating inside its own cell.
    <span className="flex min-w-0 items-center gap-1 text-muted-foreground" title={login}>
      <UserCircle2 size={12} className="shrink-0" />
      <span className="truncate">{login}</span>
    </span>
  )
}

export function IncidentAssigneePicker({
  incident,
  onChanged,
}: {
  incident: Incident
  onChanged: () => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [custom, setCustom] = useState('')
  const ref = useRef<HTMLDivElement>(null)
  const current = incident.incidentAssignedTo
  const { users, busy, assign } = useIncidentAssignment(incident.id, open, () => {
    setOpen(false)
    onChanged()
  })

  const customTrimmed = custom.trim()
  const canAssignCustom = customTrimmed.length > 0 && customTrimmed !== current
  const assignCustom = () => {
    if (!canAssignCustom) return
    void assign(customTrimmed)
    setCustom('')
  }

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const filtered = (users ?? []).filter((u) => (q ? u.toLowerCase().includes(q.toLowerCase()) : true))

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <UserCircle2 size={13} /> <span className="max-w-[120px] truncate">{current || t('incidents.assign.assign')}</span>{' '}
        <ChevronDown size={12} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-60 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              autoFocus
              placeholder={t('incidents.assign.search')}
              className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          {current && (
            <button
              onClick={() => void assign('')}
              className="flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-xs text-muted-foreground hover:bg-muted"
            >
              <X size={12} /> {t('incidents.assign.unassign')}
            </button>
          )}
          <div className="max-h-56 overflow-y-auto">
            {users === null && (
              <div className="flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              </div>
            )}
            {users !== null && filtered.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('incidents.assign.noUsers')}</div>
            )}
            {filtered.map((u) => (
              <button
                key={u}
                onClick={() => void assign(u)}
                className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted"
              >
                <span className="truncate">{u}</span>
                {u === current && <Check size={14} className="shrink-0 text-primary" />}
              </button>
            ))}
          </div>
          <div className="mt-1 border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t('incidents.assign.custom')}
            </div>
            <div className="flex items-center gap-1.5">
              <input
                value={custom}
                onChange={(e) => setCustom(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && assignCustom()}
                placeholder={t('incidents.assign.customPlaceholder')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                onClick={assignCustom}
                disabled={!canAssignCustom || busy}
                title={t('incidents.assign.custom')}
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
