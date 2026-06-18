import { useEffect, useRef, useState } from 'react'
import { Tag as TagIcon } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { AlertTag } from '../types/alert.types'

/* A tag chip painted with the tag's own colour (from the tag catalog). The alert
 * only stores tag names, so we look up the colour by name. */
export function TagChip({ name, catalog, size = 'sm' }: { name: string; catalog: AlertTag[]; size?: 'sm' | 'xs' }) {
  const color = catalog.find((t) => t.tagName === name)?.tagColor
  const style = color ? { backgroundColor: `${color}26`, color, borderColor: `${color}59` } : undefined
  return (
    <span
      className={cn(
        'inline-flex max-w-[140px] shrink-0 items-center gap-1 whitespace-nowrap rounded-md border font-medium leading-none',
        size === 'xs' ? 'px-1.5 py-0.5 text-[10px]' : 'px-2 py-0.5 text-[11px]',
        !color && 'border-border bg-muted text-muted-foreground'
      )}
      style={style}
      title={name}
    >
      <TagIcon size={size === 'xs' ? 9 : 10} className="shrink-0" />
      <span className="truncate">{name}</span>
    </span>
  )
}

export function Menu({ trigger, children }: { trigger: React.ReactNode; children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])
  return (
    <div className="relative" ref={ref}>
      <button onClick={() => setOpen((v) => !v)} className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 hover:bg-muted">
        {trigger}
      </button>
      {open && (
        <div onClick={() => setOpen(false)} className="absolute left-0 top-full z-30 mt-1 max-h-64 w-48 overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-lg">
          {children}
        </div>
      )}
    </div>
  )
}

export function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

export function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

export function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
