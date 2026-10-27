import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { ChevronDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { BreachStatus } from '../types'

const BREACH_STATUSES: BreachStatus[] = ['OPEN', 'IN_MITIGATION', 'COMPLETED']
const STATUS_KEYS: Record<BreachStatus, string> = {
  OPEN: 'open',
  IN_MITIGATION: 'mitigation',
  COMPLETED: 'completed',
}
const STATUS_STYLES: Record<BreachStatus, string> = {
  OPEN: 'bg-red-500/10 text-red-400 ring-red-500/20',
  IN_MITIGATION: 'bg-amber-500/10 text-amber-400 ring-amber-500/20',
  COMPLETED: 'bg-blue-500/10 text-blue-400 ring-blue-500/20',
}

export function StatusPill({
  status,
  label,
  onChange,
}: {
  status: BreachStatus
  label: string
  onChange: (status: BreachStatus) => void
}) {
  const { t } = useTranslation()
  const [position, setPosition] = useState<{ left: number; top: number } | null>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!position) return
    const close = () => setPosition(null)
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') close()
    }
    const onMouseDown = (event: MouseEvent) => {
      const target = event.target as Node
      if (!triggerRef.current?.contains(target) && !menuRef.current?.contains(target)) close()
    }
    document.addEventListener('mousedown', onMouseDown)
    document.addEventListener('keydown', onKeyDown)
    window.addEventListener('scroll', close, true)
    window.addEventListener('resize', close)
    return () => {
      document.removeEventListener('mousedown', onMouseDown)
      document.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('scroll', close, true)
      window.removeEventListener('resize', close)
    }
  }, [position])

  const toggle = () => {
    if (position) {
      setPosition(null)
      return
    }
    const rect = triggerRef.current?.getBoundingClientRect()
    if (rect) setPosition({ left: rect.left, top: rect.bottom + 4 })
  }

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="listbox"
        aria-expanded={position !== null}
        onClick={(event) => {
          event.stopPropagation()
          toggle()
        }}
        className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-1 text-[10px] font-medium ring-1 ring-inset ${STATUS_STYLES[status]}`}
      >
        <span className="h-1.5 w-1.5 rounded-full bg-current" />
        {label}
        <ChevronDown size={10} />
      </button>
      {position && createPortal(
        <div
          ref={menuRef}
          role="listbox"
          aria-label={label}
          onClick={(event) => event.stopPropagation()}
          style={{ position: 'fixed', left: position.left, top: position.top }}
          className="z-[95] w-max min-w-40 rounded-md border border-border bg-popover py-1 shadow-lg"
        >
          {BREACH_STATUSES.map((option) => (
            <button
              key={option}
              type="button"
              role="option"
              aria-selected={option === status}
              disabled={option === status}
              onClick={() => {
                onChange(option)
                setPosition(null)
              }}
              className="flex w-full items-center px-3 py-1.5 text-left hover:bg-muted disabled:opacity-50 disabled:hover:bg-transparent"
            >
              <span className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-1 text-[10px] font-medium ring-1 ring-inset ${STATUS_STYLES[option]}`}>
                <span className="h-1.5 w-1.5 rounded-full bg-current" />
                {t(`darkWebMonitoring.status.${STATUS_KEYS[option]}`)}
              </span>
            </button>
          ))}
        </div>,
        document.body,
      )}
    </>
  )
}
