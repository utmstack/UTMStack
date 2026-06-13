import { AlertTriangle, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      {children}
    </div>
  )
}

export function DescRow({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="min-w-0">{children}</dd>
    </>
  )
}

export function Center({ children }: { children: React.ReactNode }) {
  return (
    <div className="mt-4 flex flex-1 items-center justify-center gap-2 rounded-xl border border-border bg-card text-sm text-muted-foreground">
      {children}
    </div>
  )
}

export function TabLoader() {
  return (
    <div className="flex h-32 items-center justify-center">
      <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
    </div>
  )
}

export function TabEmpty({ children }: { children: React.ReactNode }) {
  return <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">{children}</div>
}

export function TabError({ onRetry }: { onRetry: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex h-32 items-center justify-center gap-2 text-sm text-muted-foreground">
      <AlertTriangle size={15} className="text-amber-500" /> {t('incidents.loadError')}
      <button onClick={onRetry} className="text-primary hover:underline">
        {t('incidents.retry')}
      </button>
    </div>
  )
}
