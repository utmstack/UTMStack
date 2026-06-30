import { useEffect, useRef, useState } from 'react'

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
