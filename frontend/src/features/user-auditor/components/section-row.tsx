import type { ReactNode } from 'react'

export function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

export function Row({ k, children, className }: { k: string; children: ReactNode; className?: string }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className={`break-words ${className ?? ''}`}>{children}</dd>
    </>
  )
}
