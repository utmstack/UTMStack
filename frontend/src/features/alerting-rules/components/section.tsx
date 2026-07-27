import type { ReactNode } from 'react'

export function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      {children}
    </div>
  )
}
