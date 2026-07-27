import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, RefreshCw } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'

const OPTIONS: { id: string; ms: number | null }[] = [
  { id: 'off', ms: null },
  { id: '10s', ms: 10_000 },
  { id: '30s', ms: 30_000 },
  { id: '1m', ms: 60_000 },
  { id: '5m', ms: 300_000 },
  { id: '15m', ms: 900_000 },
]

export function DashboardRefreshInterval({
  value,
  onChange,
}: {
  value: number | null
  onChange: (next: number | null) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const current = OPTIONS.find((o) => o.ms === value) ?? OPTIONS[0]

  useEffect(() => {
    const h = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('mousedown', h)
    return () => window.removeEventListener('mousedown', h)
  }, [])

  return (
    <div ref={ref} className="relative">
      <Button
        variant="outline"
        size="sm"
        onClick={() => setOpen((v) => !v)}
        className="gap-1.5"
        title={t('dashboards.refresh.title') ?? undefined}
      >
        <RefreshCw size={13} />
        <span>{t(`dashboards.refresh.options.${current.id}`)}</span>
        <ChevronDown size={12} className="opacity-60" />
      </Button>

      {open && (
        <div className="absolute right-0 z-50 mt-1 w-32 rounded-lg border border-border bg-popover p-1 text-popover-foreground shadow-lg">
          {OPTIONS.map((o) => (
            <button
              key={o.id}
              type="button"
              onClick={() => {
                onChange(o.ms)
                setOpen(false)
              }}
              className={cn(
                'block w-full rounded-md px-2 py-1 text-left text-xs transition-colors hover:bg-muted',
                o.id === current.id && 'bg-primary/5 text-foreground',
              )}
            >
              {t(`dashboards.refresh.options.${o.id}`)}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
