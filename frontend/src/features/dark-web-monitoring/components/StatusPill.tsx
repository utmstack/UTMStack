import { useRef, useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { FloatingPanel } from '@/shared/components/ui/floating-panel'
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
  const [open, setOpen] = useState(false)
  const triggerRef = useRef<HTMLButtonElement>(null)

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={(event) => {
          event.stopPropagation()
          setOpen((current) => !current)
        }}
        className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-1 text-[10px] font-medium ring-1 ring-inset ${STATUS_STYLES[status]}`}
      >
        <span className="h-1.5 w-1.5 rounded-full bg-current" />
        {label}
        <ChevronDown size={10} />
      </button>
      <FloatingPanel open={open} onClose={() => setOpen(false)} triggerRef={triggerRef}>
        <div role="listbox" aria-label={label}>
          {BREACH_STATUSES.map((option) => (
            <button
              key={option}
              type="button"
              role="option"
              aria-selected={option === status}
              disabled={option === status}
              onClick={() => {
                onChange(option)
                setOpen(false)
              }}
              className="flex w-full items-center px-3 py-1.5 text-left hover:bg-muted disabled:opacity-50 disabled:hover:bg-transparent"
            >
              <span className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-1 text-[10px] font-medium ring-1 ring-inset ${STATUS_STYLES[option]}`}>
                <span className="h-1.5 w-1.5 rounded-full bg-current" />
                {t(`darkWebMonitoring.status.${STATUS_KEYS[option]}`)}
              </span>
            </button>
          ))}
        </div>
      </FloatingPanel>
    </>
  )
}
