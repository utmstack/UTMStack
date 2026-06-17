import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { Check, ChevronDown, Server, Settings2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { setCurrentInstanceId, useCurrentInstanceId } from '@/shared/lib/current-instance'
import { useInstances } from '../hooks/use-instances'

/**
 * Topbar instance switcher (federation mode). Switching drops the previous
 * instance's React Query cache so no stale data bleeds across instances.
 */
export function InstanceSelector() {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const navigate = useNavigate()
  const qc = useQueryClient()
  const currentId = useCurrentInstanceId()
  const { data: instances } = useInstances()
  const current = instances?.find((i) => i.id === currentId)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('mousedown', handler)
    return () => window.removeEventListener('mousedown', handler)
  }, [])

  const switchTo = (id: number) => {
    setOpen(false)
    if (id === currentId) return
    setCurrentInstanceId(id)
    qc.clear() // the previous instance's cached queries are no longer valid
    navigate('/home')
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'flex h-8 items-center gap-2 rounded-md border border-border px-2.5 text-sm transition-colors',
          open ? 'bg-muted text-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground',
        )}
      >
        <Server size={14} className="text-primary" />
        <span className="max-w-[160px] truncate">{current?.name ?? 'Select instance'}</span>
        <ChevronDown size={12} className={cn('transition-transform', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-50 mt-1 w-64 overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-lg">
          <div className="max-h-72 overflow-y-auto py-1">
            {instances?.map((i) => (
              <button
                key={i.id}
                onClick={() => switchTo(i.id)}
                className="flex w-full items-center justify-between gap-2 px-3 py-2 text-sm transition-colors hover:bg-muted"
              >
                <span className="min-w-0">
                  <span className="block truncate">{i.name}</span>
                  <span className="block truncate font-mono text-[11px] text-muted-foreground">{i.baseUrl}</span>
                </span>
                {i.id === currentId && <Check size={14} className="shrink-0 text-primary" />}
              </button>
            ))}
          </div>
          <button
            onClick={() => {
              setOpen(false)
              navigate('/instances')
            }}
            className="flex w-full items-center gap-2 border-t border-border px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <Settings2 size={14} /> Manage instances
          </button>
        </div>
      )}
    </div>
  )
}
