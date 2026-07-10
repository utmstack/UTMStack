import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Clock, ListFilter, RefreshCw, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { toast } from 'sonner'
import { useTiConfigStatus } from '../hooks/use-ti-config-status'
import { useTiFeeds } from '../hooks/use-ti-feeds'
import { useTiSearch } from '../hooks/use-ti-search'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError } from '../services/ti-errors'
import { downloadCsv, toCsv } from '../services/csv'
import { searchItemValue } from '../domain/threat-intel.types'
import { NotConfiguredState } from '../components/NotConfiguredState'
import { ThreatIntelHeader } from '../components/ThreatIntelHeader'
import { MatchOverviewCard } from '../components/MatchOverviewCard'
import { LookupBar } from '../components/LookupBar'
import { TabsRow, type TabKey } from '../components/TabsRow'
import { IocTable } from '../components/IocTable'
import { IocDrawer } from '../components/IocDrawer'
import { ActorsList } from '../components/ActorsList'
import { ActorDrawer } from '../components/ActorDrawer'
import { FeedsList } from '../components/FeedsList'
import type { EntitySearchResponse, EntitySummary } from '../domain/threat-intel.types'

export function ThreatIntelPage() {
  const { isConfigured, isLoading: configLoading } = useTiConfigStatus()
  const feedsQuery = useTiFeeds()
  const searchMutation = useTiSearch()

  const [query, setQuery] = useState<string>('*')
  const [page, setPage] = useState(0)
  const [size, setSize] = useState(20)
  const [results, setResults] = useState<EntitySearchResponse | null>(null)
  const [iocs, setIocs] = useState<EntitySummary[]>([])

  // 'replace' when the user searches, jumps pages, or changes page size.
  // 'append' when infinite scroll asks for the next page.
  const modeRef = useRef<'replace' | 'append'>('replace')
  // Guards against out-of-order responses (older mutation lands after newer).
  const seqRef = useRef(0)

  const [tab, setTab] = useState<TabKey>('iocs')
  const [openIoc, setOpenIoc] = useState<string | null>(null)
  const [openActor, setOpenActor] = useState<string | null>(null)
  const [uiSearch, setUiSearch] = useState('')
  const [isExporting, setIsExporting] = useState(false)

  useEffect(() => {
    if (!isConfigured) return
    const my = ++seqRef.current
    const mode = modeRef.current
    searchMutation.mutate(
      { query, page: page + 1, size },
      {
        onSuccess: (data) => {
          if (my !== seqRef.current) return
          if (data?.kind === 'not-configured') return
          if (data?.kind !== 'ok') return
          setResults(data.value)
          setIocs((prev) =>
            mode === 'append' ? [...prev, ...data.value.results] : data.value.results
          )
        },
      }
    )
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query, page, size, isConfigured])

  if (configLoading) return null
  if (isConfigured === false) return <NotConfiguredState />

  const actors: EntitySummary[] = iocs.filter((e) => e.type === 'threat')
  const feedsCount = feedsQuery.data?.kind === 'ok' ? feedsQuery.data.value.length : 0
  const totalItems = results?.items ?? 0
  const totalPages = results?.pages ?? 0
  const hasMore = page + 1 < totalPages

  const handleSearch = (q: string) => {
    modeRef.current = 'replace'
    setPage(0)
    setQuery(q)
  }

  const handlePageChange = (p: number) => {
    modeRef.current = 'replace'
    setPage(p)
  }

  const handlePageSizeChange = (s: number) => {
    modeRef.current = 'replace'
    setSize(s)
    setPage(0)
  }

  const handleLoadMore = () => {
    if (!hasMore || searchMutation.isPending) return
    modeRef.current = 'append'
    setPage((p) => p + 1)
  }

  const handleExport = async () => {
    if (!totalItems || isExporting) return
    setIsExporting(true)
    try {
      const res = await threatIntelHttpService.search({ query, page: 1, size: totalItems })
      if (res.kind !== 'ok') return
      const rows = res.value.results.map((r) => [
        r.id,
        r.type,
        searchItemValue(r),
        r.tags.join('|'),
        r.lastSeen,
        r.reputation,
        r.bestReputation,
        r.worstReputation,
        r.accuracy,
      ])
      const csv = toCsv(
        ['id', 'type', 'value', 'tags', 'lastSeen', 'reputation', 'bestReputation', 'worstReputation', 'accuracy'],
        rows,
      )
      downloadCsv(`iocs-${new Date().toISOString().replace(/[:.]/g, '-')}.csv`, csv)
    } catch (e) {
      toast.error(describeError(e))
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <ThreatIntelHeader
        matchedCount={totalItems}
        onExport={handleExport}
        isExporting={isExporting}
      />

      <div className="mt-5">
        <MatchOverviewCard />
      </div>

      <div className="mt-5">
        <LookupBar onSearch={handleSearch} isPending={searchMutation.isPending} />
      </div>

      <div className="mt-5">
        <TabsRow
          active={tab}
          onChange={setTab}
          counts={{ iocs: totalItems, actors: actors.length, feeds: feedsCount }}
        />
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
            value={uiSearch}
            onChange={(e) => setUiSearch(e.target.value)}
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
          onClick={handleSearch.bind(null, query)}
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} />
        </button>
      </div>

      <div className="mt-3">
        {tab === 'iocs' && (
          <IocTable
            iocs={iocs}
            onOpen={setOpenIoc}
            isLoading={searchMutation.isPending}
            page={page}
            pageSize={size}
            totalItems={totalItems}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
            hasMore={hasMore}
            onLoadMore={handleLoadMore}
          />
        )}
        {tab === 'actors' && <ActorsList actors={actors} onOpen={setOpenActor} />}
        {tab === 'feeds' && <FeedsList />}
      </div>

      <IocDrawer id={openIoc} onClose={() => setOpenIoc(null)} />
      <ActorDrawer id={openActor} onClose={() => setOpenActor(null)} />
    </div>
  )
}
