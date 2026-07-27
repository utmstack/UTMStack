import type { ReactNode } from 'react'

export function RowMessage({ children }: { children: ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
  )
}
