import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Loader2, UserCircle2, X } from 'lucide-react'
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
    <span className="inline-flex min-w-0 items-center gap-1 text-muted-foreground">
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
  const ref = useRef<HTMLDivElement>(null)
  const current = incident.incidentAssignedTo
  const { users, busy, assign } = useIncidentAssignment(incident.id, open, () => {
    setOpen(false)
    onChanged()
  })

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
              onClick={() => void assign(null)}
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
        </div>
      )}
    </div>
  )
}
