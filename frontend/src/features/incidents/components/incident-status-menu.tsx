import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { ChevronDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { PILL_BASE } from '@/features/alerts/lib/alert-meta'
import { useIncidentStatus } from '../hooks/use-incident-status'
import { ST_META, STATUSES, statusKey } from '../lib/incident-meta'
import type { Incident } from '../types/incident.types'

/**
 * Status pill that opens the list of statuses, like the one on an alert row.
 * The list is portaled with fixed coordinates: a table with one or two rows
 * sits in a scroll container shorter than the list, which would clip it.
 */
export function IncidentStatusMenu({ incident, onChanged }: { incident: Incident; onChanged: (id: string) => void }) {
  const { t } = useTranslation()
  const { busy, changeStatus } = useIncidentStatus(incident, onChanged)
  const [pos, setPos] = useState<{ left: number; top: number } | null>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const status = incident.incidentStatus
  const open = pos !== null

  useLayoutEffect(() => {
    if (!open || !triggerRef.current || !menuRef.current) return
    const trigger = triggerRef.current.getBoundingClientRect()
    const height = menuRef.current.offsetHeight
    const below = trigger.bottom + 4
    const top = below + height > window.innerHeight ? Math.max(4, trigger.top - 4 - height) : below
    menuRef.current.style.top = `${top}px`
  }, [open])

  useEffect(() => {
    if (!open) return
    const close = () => setPos(null)
    const onDown = (e: MouseEvent) => {
      const target = e.target as Node
      if (!triggerRef.current?.contains(target) && !menuRef.current?.contains(target)) close()
    }
    document.addEventListener('mousedown', onDown)
    window.addEventListener('scroll', close, true)
    window.addEventListener('resize', close)
    return () => {
      document.removeEventListener('mousedown', onDown)
      window.removeEventListener('scroll', close, true)
      window.removeEventListener('resize', close)
    }
  }, [open])

  const toggle = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (open) return setPos(null)
    const rect = triggerRef.current?.getBoundingClientRect()
    if (rect) setPos({ left: rect.left, top: rect.bottom + 4 })
  }

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        disabled={busy}
        onClick={toggle}
        className={cn(PILL_BASE, ST_META[status].pill, busy && 'opacity-60')}
      >
        <span>{t(`incidents.status.${statusKey(status)}`)}</span>
        <ChevronDown size={10} />
      </button>
      {pos &&
        createPortal(
          <div
            ref={menuRef}
            onClick={(e) => e.stopPropagation()}
            style={{ position: 'fixed', left: pos.left, top: pos.top }}
            className="z-[70] w-max min-w-[10rem] rounded-md border border-border bg-popover py-1 shadow-lg"
          >
            {STATUSES.map((s) => (
              <button
                key={s}
                disabled={s === status}
                onClick={() => {
                  setPos(null)
                  void changeStatus(s)
                }}
                className="flex w-full items-center px-3 py-1.5 text-left hover:bg-muted disabled:opacity-50 disabled:hover:bg-transparent"
              >
                <span className={cn(PILL_BASE, ST_META[s].pill)}>{t(`incidents.status.${statusKey(s)}`)}</span>
              </button>
            ))}
          </div>,
          document.body,
        )}
    </>
  )
}
