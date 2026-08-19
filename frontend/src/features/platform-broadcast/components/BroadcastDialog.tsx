import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Search } from 'lucide-react'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import type { BulkResult, BulkSelector } from '../services/broadcast-http.service'
import {
  DEFAULT_TENANT_ID,
  filterForAllTenants,
  useTenantsForBroadcast,
} from '../hooks/useTenantsForBroadcast'
import { BroadcastResultPanel } from './BroadcastResultPanel'

export interface BroadcastDialogProps {
  open: boolean
  title: string
  /** Extra context above the tenant picker (e.g. "Broadcast SMTP config"). */
  intro?: ReactNode
  /**
   * SMTP + branding endpoints exclude the platform-plane tenant even under
   * `allTenants=true`. Set this so the picker's count matches what will run.
   */
  excludeDefaultTenant?: boolean
  onClose: () => void
  /**
   * Fire the bulk endpoint with the selector the user picked. The dialog
   * handles busy state and result display.
   */
  onConfirm: (selector: BulkSelector) => Promise<BulkResult>
}

export function BroadcastDialog({
  open,
  title,
  intro,
  excludeDefaultTenant = false,
  onClose,
  onConfirm,
}: BroadcastDialogProps) {
  const { t } = useTranslation()
  const { tenants, loading, error } = useTenantsForBroadcast(open)
  const [allTenants, setAllTenants] = useState(false)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [search, setSearch] = useState('')
  const [result, setResult] = useState<BulkResult | null>(null)
  const [busy, setBusy] = useState(false)

  // Reset local UI state every time the modal opens fresh.
  useEffect(() => {
    if (!open) return
    setAllTenants(false)
    setSelected(new Set())
    setSearch('')
    setResult(null)
    setBusy(false)
  }, [open])

  const targetTenants = useMemo(() => {
    if (allTenants) return filterForAllTenants(tenants, excludeDefaultTenant)
    return tenants.filter((t) => selected.has(t.id))
  }, [tenants, allTenants, selected, excludeDefaultTenant])

  const filteredList = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return tenants
    return tenants.filter(
      (tenant) =>
        tenant.name.toLowerCase().includes(q) || tenant.domain.toLowerCase().includes(q),
    )
  }, [tenants, search])

  const toggle = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const canConfirm = !busy && !result && targetTenants.length > 0

  const run = async () => {
    if (!canConfirm) return
    setBusy(true)
    try {
      const selector: BulkSelector = allTenants
        ? { tenantIds: [], allTenants: true }
        : { tenantIds: targetTenants.map((tenant) => tenant.id), allTenants: false }
      const r = await onConfirm(selector)
      setResult(r)
    } finally {
      setBusy(false)
    }
  }

  const body: ReactNode = result ? (
    <BroadcastResultPanel result={result} tenants={tenants} />
  ) : (
    <div className="flex flex-col gap-3">
      <div className="flex items-start gap-2 rounded border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive-foreground">
        <AlertTriangle size={16} className="mt-0.5 shrink-0 text-destructive" />
        <div>
          <p className="font-medium text-foreground">
            {t(
              'platformBroadcast.warning.title',
              'This action will run on every selected tenant.',
            )}
          </p>
          <p className="text-xs text-muted-foreground">
            {t(
              'platformBroadcast.warning.body',
              'Per-tenant failures are reported after the run and are not rolled back.',
            )}
          </p>
        </div>
      </div>

      {intro && <div className="text-sm text-foreground">{intro}</div>}

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={allTenants}
          onChange={(e) => setAllTenants(e.target.checked)}
          disabled={busy}
        />
        <span>{t('platformBroadcast.allTenants', 'Apply to all tenants')}</span>
      </label>

      {allTenants && excludeDefaultTenant && (
        <p className="text-xs text-muted-foreground">
          {t(
            'platformBroadcast.excludesDefault',
            'The platform-plane tenant is excluded from this action.',
          )}
        </p>
      )}

      {!allTenants && (
        <div className="flex flex-col gap-2">
          <div className="relative">
            <Search
              size={14}
              className="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground"
            />
            <input
              type="search"
              placeholder={t('platformBroadcast.searchPlaceholder', 'Search tenants…')}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              disabled={loading || busy}
              className="h-8 w-full rounded border border-border bg-background pl-7 pr-2 text-sm"
            />
          </div>
          <div className="max-h-52 overflow-y-auto rounded border border-border">
            {loading ? (
              <p className="p-3 text-xs text-muted-foreground">
                {t('platformBroadcast.loading', 'Loading tenants…')}
              </p>
            ) : error ? (
              <p className="p-3 text-xs text-destructive">{error}</p>
            ) : filteredList.length === 0 ? (
              <p className="p-3 text-xs text-muted-foreground">
                {t('platformBroadcast.empty', 'No tenants match.')}
              </p>
            ) : (
              <ul>
                {filteredList.map((tenant) => (
                  <li key={tenant.id}>
                    <label className="flex cursor-pointer items-center gap-2 border-b border-border/60 px-2 py-1.5 text-sm last:border-0 hover:bg-muted/50">
                      <input
                        type="checkbox"
                        checked={selected.has(tenant.id)}
                        onChange={() => toggle(tenant.id)}
                        disabled={busy}
                      />
                      <span className="flex-1 truncate">{tenant.name}</span>
                      <span className="truncate text-xs text-muted-foreground">
                        {tenant.domain}
                      </span>
                      {tenant.id === DEFAULT_TENANT_ID && (
                        <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                          {t('platformBroadcast.platformTenant', 'platform')}
                        </span>
                      )}
                    </label>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      )}

      <p className="text-xs text-muted-foreground">
        {t(
          'platformBroadcast.targetCount',
          'Will target {{count}} tenant(s).',
          { count: targetTenants.length },
        )}
      </p>
    </div>
  )

  const confirmLabel = result
    ? t('common.actions.close', 'Close')
    : t('platformBroadcast.confirm', 'Apply to {{count}} tenant(s)', {
        count: targetTenants.length,
      })

  return (
    <ConfirmDialog
      open={open}
      title={title}
      body={body}
      danger={!result}
      icon={AlertTriangle}
      busy={busy}
      confirmLabel={confirmLabel}
      cancelLabel={t('common.actions.cancel', 'Cancel')}
      hideCancel={!!result}
      onClose={onClose}
      onConfirm={result ? onClose : run}
    />
  )
}
