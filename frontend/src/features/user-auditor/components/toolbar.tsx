import type { TFunction } from 'i18next'
import { ChevronDown, LayoutGrid, ListIcon, RefreshCw, Search, Server, Users } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { useUserAuditor } from '../context/user-auditor.context'

export function Toolbar({
  search,
  onSearch,
  tenant,
  tenants,
  onTenant,
  identity,
  onIdentity,
  layout,
  onLayoutChange,
  onRefresh,
  loading,
  t,
}: {
  search: string
  onSearch: (s: string) => void
  tenant: string
  tenants: string[]
  onTenant: (t: string) => void
  identity: string
  onIdentity: (id: string) => void
  layout: 'list' | 'cards'
  onLayoutChange: (l: 'list' | 'cards') => void
  onRefresh: () => void
  loading: boolean
  t: TFunction
}) {
  const { identities } = useUserAuditor()
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={t('userAuditor.searchPlaceholder')}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      {tenants.length > 0 && (
        <div className="relative">
          <Server size={13} className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <select
            value={tenant}
            onChange={(e) => onTenant(e.target.value)}
            className="h-9 appearance-none rounded-md border border-border bg-background pl-8 pr-7 text-xs text-foreground"
          >
            <option value="">{t('userAuditor.allTenants')}</option>
            {tenants.map((tn) => (
              <option key={tn} value={tn}>
                {tn}
              </option>
            ))}
          </select>
          <ChevronDown size={12} className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 opacity-60" />
        </div>
      )}
      <div className="relative">
        <Users size={13} className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <select
          value={identity}
          onChange={(e) => onIdentity(e.target.value)}
          className="h-9 appearance-none rounded-md border border-border bg-background pl-8 pr-7 text-xs text-foreground"
        >
          <option value="">{t('userAuditor.allIdentities')}</option>
          <option value="none">{t('userAuditor.noneIdentity')}</option>
          {identities.map((i) => (
            <option key={i.id} value={i.id}>
              {i.name}
            </option>
          ))}
        </select>
        <ChevronDown size={12} className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 opacity-60" />
      </div>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          onClick={() => onLayoutChange('list')}
          title={t('userAuditor.layout.list')}
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground',
          )}
        >
          <ListIcon size={13} />
        </button>
        <button
          onClick={() => onLayoutChange('cards')}
          title={t('userAuditor.layout.cards')}
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground',
          )}
        >
          <LayoutGrid size={13} />
        </button>
      </div>
      <button
        title={t('userAuditor.refresh')}
        onClick={onRefresh}
        className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </button>
    </div>
  )
}
