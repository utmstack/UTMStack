import { useState } from 'react'
import { ChevronDown, Clock, ListFilter, RefreshCw, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useTiConfigStatus } from '../hooks/use-ti-config-status'
import { NotConfiguredState } from '../components/NotConfiguredState'
import { ThreatIntelHeader } from '../components/ThreatIntelHeader'
import { MatchOverviewCard } from '../components/MatchOverviewCard'
import { IntelInsightsCard } from '../components/IntelInsightsCard'
import { LookupBar } from '../components/LookupBar'
import { TabsRow, type TabKey } from '../components/TabsRow'
import { IocTable } from '../components/IocTable'
import { IocDrawer } from '../components/IocDrawer'
import { ActorsList } from '../components/ActorsList'
import { ActorDrawer } from '../components/ActorDrawer'
import { FeedsList } from '../components/FeedsList'
import type { EntitySearchResponse, EntitySummary } from '../domain/threat-intel.types'

export function ThreatIntelPage() {
  const { isConfigured, isLoading } = useTiConfigStatus()
  const [results, setResults] = useState<EntitySearchResponse | null>(null)
  const [tab, setTab] = useState<TabKey>('iocs')
  const [openIoc, setOpenIoc] = useState<string | null>(null)
  const [openActor, setOpenActor] = useState<string | null>(null)
  const [search, setSearch] = useState('')

  if (isLoading) return null
  if (isConfigured === false) return <NotConfiguredState />

  const iocs: EntitySummary[] = results?.results ?? []
  // ponytail: actors derived by CM entity type until a dedicated actors endpoint exists.
  const actors: EntitySummary[] = iocs.filter((e) => e.type === 'threat')

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <ThreatIntelHeader matchedCount={results?.total} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <MatchOverviewCard total={results?.total} />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <IntelInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <LookupBar onResults={setResults} />
      </div>

      <div className="mt-5">
        <TabsRow active={tab} onChange={setTab} counts={{ iocs: iocs.length, actors: actors.length }} />
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[280px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={
              tab === 'feeds'
                ? 'Search feeds…'
                : tab === 'actors'
                  ? 'Search actors, aliases, techniques…'
                  : 'Search IOC value, tag, source…'
            }
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9 pl-9"
          />
        </div>
        {tab === 'iocs' && (
          <>
            <Button variant="outline" size="sm">
              <Clock size={14} className="mr-2" />
              Last 24 hours
              <ChevronDown size={12} className="ml-1.5 opacity-60" />
            </Button>
            <Button variant="outline" size="sm">
              <ListFilter size={14} className="mr-2" />
              Filters
            </Button>
          </>
        )}
        <button
          title="Refresh"
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} />
        </button>
      </div>

      <div className="mt-3">
        {tab === 'iocs' && <IocTable iocs={iocs} onOpen={setOpenIoc} />}
        {tab === 'actors' && <ActorsList actors={actors} onOpen={setOpenActor} />}
        {tab === 'feeds' && <FeedsList />}
      </div>

      <IocDrawer id={openIoc} onClose={() => setOpenIoc(null)} />
      <ActorDrawer id={openActor} onClose={() => setOpenActor(null)} />
    </div>
  )
}
