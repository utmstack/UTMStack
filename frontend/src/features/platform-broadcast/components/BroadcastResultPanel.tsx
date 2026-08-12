import { useTranslation } from 'react-i18next'
import { CheckCircle2, XCircle } from 'lucide-react'
import type { BulkResult } from '../services/broadcast-http.service'
import type { Tenant } from '@/features/tenants/types/tenant.types'

interface BroadcastResultPanelProps {
  result: BulkResult
  tenants: Tenant[]
}

function nameFor(tenants: Tenant[], id: string): string {
  return tenants.find((t) => t.id === id)?.name ?? id
}

export function BroadcastResultPanel({ result, tenants }: BroadcastResultPanelProps) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-3">
      <div>
        <div className="mb-1 flex items-center gap-2 text-sm font-medium text-foreground">
          <CheckCircle2 size={14} className="text-emerald-500" />
          {t('platformBroadcast.result.succeeded', 'Succeeded')} ({result.succeeded.length})
        </div>
        {result.succeeded.length === 0 ? (
          <p className="text-xs text-muted-foreground">
            {t('platformBroadcast.result.noneSucceeded', 'No tenants succeeded.')}
          </p>
        ) : (
          <ul className="max-h-32 overflow-y-auto rounded border border-border bg-muted/40 p-2 text-xs">
            {result.succeeded.map((id) => (
              <li key={id} className="truncate">
                {nameFor(tenants, id)}
              </li>
            ))}
          </ul>
        )}
      </div>

      <div>
        <div className="mb-1 flex items-center gap-2 text-sm font-medium text-foreground">
          <XCircle size={14} className="text-destructive" />
          {t('platformBroadcast.result.failed', 'Failed')} ({result.failed.length})
        </div>
        {result.failed.length === 0 ? (
          <p className="text-xs text-muted-foreground">
            {t('platformBroadcast.result.noneFailed', 'No failures.')}
          </p>
        ) : (
          <ul className="max-h-40 overflow-y-auto rounded border border-border bg-muted/40 p-2 text-xs">
            {result.failed.map((f) => (
              <li key={f.tenantId} className="mb-1 last:mb-0">
                <span className="font-medium">{nameFor(tenants, f.tenantId)}</span>
                <span className="text-muted-foreground"> — {f.error}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
