import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Crown } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

interface EnterpriseGateProps {
  /** The page's own header, so the gate reads as that page and not a dead end. */
  header?: ReactNode
  title: string
  body: string
  cta: string
}

/**
 * Shown in place of a paid feature on a Community install. It is presentation
 * only: the backend refuses these operations regardless of what is rendered.
 */
export function EnterpriseGate({ header, title, body, cta }: EnterpriseGateProps) {
  return (
    <div className="w-full px-6 pb-6 pt-3">
      {header}
      <div className="mt-10 flex flex-col items-center justify-center rounded-xl border border-border bg-card px-6 py-16 text-center">
        <span className="flex h-14 w-14 items-center justify-center rounded-full bg-amber-500/10 text-amber-500 ring-1 ring-inset ring-amber-500/30">
          <Crown size={26} strokeWidth={1.75} />
        </span>
        <h2 className="mt-5 text-lg font-semibold">{title}</h2>
        <p className="mt-2 max-w-md text-sm text-muted-foreground">{body}</p>
        <Button asChild className="mt-6">
          <Link to="/settings/license">
            <Crown size={14} className="mr-1.5" />
            {cta}
          </Link>
        </Button>
      </div>
    </div>
  )
}
