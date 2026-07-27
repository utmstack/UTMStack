import type { ReactNode } from 'react'

export function Row({ k, children }: { k: string; children: ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="min-w-0">{children}</dd>
    </>
  )
}
