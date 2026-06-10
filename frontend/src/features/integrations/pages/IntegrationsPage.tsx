import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Filter, Loader2, Search } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import { IntegrationsHeader } from '@/features/integrations/components/IntegrationsHeader'
import { IntegrationsTabs } from '@/features/integrations/components/IntegrationsTabs'
import { IntegrationCard } from '@/features/integrations/components/IntegrationCard'
import { IntegrationDrawer } from '@/features/integrations/components/IntegrationDrawer'
import { AddCustomIntegrationCard } from '@/features/integrations/components/AddCustomIntegrationCard'
import { KIND_META } from '@/features/integrations/constants'
import type { Integration, Tab, DeployKind } from '@/features/integrations/types'

const LOGO = (slug: string) => {
  if(slug.startsWith('http')){
    return slug
  }
  return `/integrations/${slug}`
}

const TABS: { id: Tab; predicate: (kind: DeployKind) => boolean }[] = [
  { id: 'all', predicate: () => true },
  { id: 'agents', predicate: (k) => KIND_META[k].group === 'agents' },
  { id: 'collectors', predicate: (k) => KIND_META[k].group === 'collectors' },
  { id: 'cloud', predicate: (k) => KIND_META[k].group === 'cloud' },
  { id: 'custom', predicate: (k) => KIND_META[k].group === 'custom'}
]

function mapModuleToIntegration(module: any): Integration {
  const validKinds: DeployKind[] = ['agents & syslog', 'device', 'antivirus', 'other', 'custom', 'cloud', 'utmstack modules']
  const categoryLower = module.moduleCategory?.toLowerCase() || ''
  const dataTypeLower = module.dataType?.toLowerCase() || ''

  let kind: DeployKind = 'custom'
  if (validKinds.includes(categoryLower as DeployKind)) {
    kind = categoryLower as DeployKind
  } else if (validKinds.includes(dataTypeLower as DeployKind)) {
    kind = dataTypeLower as DeployKind
  }

  return {
    id: module.id.toString(),
    name: module.prettyName || module.moduleName,
    kind,
    status: module.moduleActive ? 'configured' : 'available',
    description: module.moduleDescription || '',
    category: module.moduleCategory || '',
    logo: LOGO(module.moduleIcon) || LOGO('placeholder.svg'),
    darkInvert: false,
    events24h: undefined,
    rate: undefined,
  }
}

export function IntegrationsPage() {
  const { t } = useTranslation()
  const integrations = useIntegrations()
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [showOnlyConfigured, setShowOnlyConfigured] = useState(false)
  const [open, setOpen] = useState<Integration | null>(null)

  const modules = integrations.modules.data || []
  const displayList = useMemo(() => modules.map(mapModuleToIntegration), [modules])

  const filtered = useMemo(() => {
    const t_filter = TABS.find((x) => x.id === tab) ?? TABS[0]
    return displayList
      .filter(({kind})=>t_filter.predicate(kind) as any)
      .filter((i) => (showOnlyConfigured ? i.status === 'configured' : true))
      .filter((i) =>
        search
          ? (i.name + i.description + i.category).toLowerCase().includes(search.toLowerCase())
          : true
      )
  }, [tab, search, showOnlyConfigured, displayList])

  const counts: Record<Tab, number> = useMemo(() => {
    return {
      all: displayList.length,
      agents: displayList.filter((i) => KIND_META[i.kind].group === 'agents').length,
      collectors: displayList.filter((i) => KIND_META[i.kind].group === 'collectors').length,
      cloud: displayList.filter((i) => KIND_META[i.kind].group === 'cloud').length,
      custom: displayList.filter((i) => KIND_META[i.kind].group === 'custom').length,
    }
  }, [displayList])

  const configuredCount = displayList.filter((i) => i.status === 'configured').length

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      {/* Loading overlay */}
      {integrations.isLoading && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-card/60 backdrop-blur-sm">
          <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-3">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span className="text-sm text-muted-foreground">{t('integrations.loading')}</span>
          </div>
        </div>
      )}

      <IntegrationsHeader configured={configuredCount} total={displayList.length} />

      <div className="mt-5 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[300px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('integrations.searchPlaceholder')}
            className="h-10 pl-9"
          />
        </div>
        <button
          onClick={() => setShowOnlyConfigured((v) => !v)}
          className={cn(
            'flex h-10 items-center gap-2 rounded-md border px-3 text-sm transition-colors',
            showOnlyConfigured
              ? 'border-primary/30 bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          <Filter size={14} />
          {t('integrations.configuredOnly')}
        </button>
      </div>

      <div className="mt-5">
        <IntegrationsTabs current={tab} onChange={setTab} counts={counts} />
      </div>

      <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
        <AddCustomIntegrationCard onClick={() => {}} />
        {filtered.map((i) => (
          <IntegrationCard key={i.id} integration={i} onOpen={() => setOpen(i)} />
        ))}
        {filtered.length === 0 && (
          <div className="col-span-full rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
            {t('integrations.noMatch')}
          </div>
        )}
      </div>

      {open && <IntegrationDrawer integration={open} onClose={() => setOpen(null)} />}
    </div>
  )
}
