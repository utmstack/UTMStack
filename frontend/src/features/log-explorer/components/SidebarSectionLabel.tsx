import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/utils'

export function SidebarSectionLabel({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn('px-2 pb-1 pt-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground/55', className)}>
      {children}
    </div>
  )
}
