import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  KeyRound,
  Loader2,
  Plus,
  RefreshCw,
  Search,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { useBilling } from '@/features/billing'
import { EnterpriseGate } from '@/shared/components/EnterpriseGate'
import { apiKeysHttpService } from '../services/api-keys-http.service'
import type { ApiKey, ApiKeyPageInfo } from '../types/api-key.types'
import { COLS, KeyRow } from '../components/KeyRow'
import { UpsertDialog } from '../components/UpsertDialog'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { RevealModal } from '../components/RevealModal'

const PAGE_SIZE = 50

type DialogState = { mode: 'create' } | { mode: 'edit'; key: ApiKey } | null
type ConfirmState = { kind: 'rotate' | 'delete'; key: ApiKey } | null

export function ApiKeysPage() {
  const { t } = useTranslation()
  const { license } = useBilling()
  const [keys, setKeys] = useState<ApiKey[] | null>(null)
  const [pageInfo, setPageInfo] = useState<ApiKeyPageInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')

  const [dialog, setDialog] = useState<DialogState>(null)
  const [confirm, setConfirm] = useState<ConfirmState>(null)
  const [revealed, setRevealed] = useState<{ name: string; token: string } | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const resp = await apiKeysHttpService.list(page, PAGE_SIZE)
      // Go marshals an empty slice as null, so allowed_ip can come back null —
      // normalize it to an array so the rest of the page can treat it uniformly.
      const normalized = resp.data.map((k) => ({ ...k, allowed_ip: k.allowed_ip ?? [] }))
      setKeys((prev) => (page === 1 ? normalized : [...(prev ?? []), ...normalized]))
      setPageInfo(resp.page_info)
    } catch {
      setError(true)
      setKeys([])
    } finally {
      setLoading(false)
    }
  }, [page])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return (keys ?? []).filter((k) =>
      q ? (k.name + ' ' + k.allowed_ip.join(' ')).toLowerCase().includes(q) : true,
    )
  }, [keys, search])

  // API keys creation/rotation is an Enterprise capability. When gated, the
  // page still renders but shows an upgrade note instead of the manager.
  const isEnterprise = license?.edition === 'enterprise'

  if (!isEnterprise) {
    return (
      <EnterpriseGate
        header={
          <header>
            <h1 className="flex items-center gap-2 text-base font-semibold">
              <KeyRound size={16} strokeWidth={1.75} />
              {t('settings.apiKeys')}
            </h1>
          </header>
        }
        title={t('apiKeys.enterprise.title')}
        body={t('apiKeys.enterprise.body')}
        cta={t('apiKeys.enterprise.upgrade')}
      />
    )
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-base font-semibold">
            <KeyRound size={16} strokeWidth={1.75} />
            {t('settings.apiKeys')}
          </h1>
          <p className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
            <span>
              <span className="font-medium text-foreground">{pageInfo?.total_items ?? 0}</span>{' '}
              {t('apiKeys.title').toLowerCase()}
            </span>
          </p>
        </div>
        <Button size="sm" onClick={() => setDialog({ mode: 'create' })}>
          <Plus size={14} className="mr-1.5" />
          {t('apiKeys.new')}
        </Button>
      </header>

      <section className="mt-6 rounded-xl border border-border bg-card px-6 pb-6 pt-4">
          <div className="flex flex-wrap items-center gap-2">
            <div className="relative min-w-[260px] flex-1">
              <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={t('apiKeys.searchPlaceholder')}
                className="h-9 pl-9"
              />
            </div>
            <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
              <RefreshCw size={14} className={cn('mr-1.5', loading && 'animate-spin')} />
              {t('apiKeys.refresh')}
            </Button>
          </div>

          <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
            <div
              className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
              style={{ gridTemplateColumns: COLS }}
            >
              <div>{t('apiKeys.col.name')}</div>
              <div>{t('apiKeys.col.allowedIps')}</div>
              <div>{t('apiKeys.col.created')}</div>
              <div>{t('apiKeys.col.lastRotated')}</div>
              <div>{t('apiKeys.col.expires')}</div>
              <div>{t('apiKeys.col.status')}</div>
              <div className="text-right">{t('apiKeys.col.actions')}</div>
            </div>

            {loading && (!keys || keys.length === 0) && (
              <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                {t('apiKeys.loading')}
              </div>
            )}

            {!loading && error && (
              <div className="flex flex-col items-center gap-3 px-6 py-12 text-sm">
                <span className="inline-flex items-center gap-2 text-muted-foreground">
                  <AlertTriangle size={16} className="text-amber-500" />
                  {t('apiKeys.loadFailed')}
                </span>
                <Button variant="outline" size="sm" onClick={() => void load()}>
                  {t('apiKeys.retry')}
                </Button>
              </div>
            )}

            {!loading && !error && filtered.length === 0 && (
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                {search ? t('apiKeys.noSearchMatch') : t('apiKeys.empty')}
              </div>
            )}

            {filtered.length > 0 &&
              filtered.map((k) => (
                <KeyRow
                  key={k.id}
                  apiKey={k}
                  onEdit={() => setDialog({ mode: 'edit', key: k })}
                  onRotate={() => setConfirm({ kind: 'rotate', key: k })}
                  onDelete={() => setConfirm({ kind: 'delete', key: k })}
                />
              ))}
          </div>

          {!error && keys && keys.length > 0 && (
            <InfiniteScrollSentinel
              onReach={() => setPage((p) => p + 1)}
              hasMore={keys.length < (pageInfo?.total_items ?? 0)}
              loading={loading}
              endLabel={t('common.allLoaded', { count: pageInfo?.total_items ?? 0 })}
            />
          )}
      </section>

      {dialog && (
        <UpsertDialog
          state={dialog}
          onClose={() => setDialog(null)}
          onSaved={(reveal) => {
            setDialog(null)
            if (reveal) setRevealed(reveal)
            void load()
          }}
        />
      )}

      {confirm && (
        <ConfirmDialog
          state={confirm}
          onClose={() => setConfirm(null)}
          onDone={(reveal) => {
            setConfirm(null)
            if (reveal) setRevealed(reveal)
            void load()
          }}
        />
      )}

      {revealed && <RevealModal name={revealed.name} token={revealed.token} onClose={() => setRevealed(null)} />}
    </div>
  )
}
