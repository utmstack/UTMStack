import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Clock, ListFilter, RefreshCw, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { toast } from 'sonner'
import { useTiConfigStatus } from '../hooks/use-ti-config-status'
import { useTiFeeds } from '../hooks/use-ti-feeds'
import { useTiSearchAdvanced } from '../hooks/use-ti-search-advanced'
import { useTiEntityLookup } from '../hooks/use-ti-entity-lookup'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError, isNotFound } from '../services/ti-errors'
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
import { FiltersPanel, EMPTY_FILTERS, filtersToRequest } from '../components/FiltersPanel'
import type { FiltersState } from '../components/FiltersPanel'
import type {
  EntitySummary,
  AdvancedSearchRequest,
  AdvancedCondition,
} from '../domain/threat-intel.types'

const MULTI_MATCH_FIELDS = [
  'attributes.ip', 'attributes.domain', 'attributes.hostname', 'attributes.url',
  'attributes.md5', 'attributes.sha1', 'attributes.sha256', 'attributes.sha3-256',
  'attributes.cve', 'attributes.email-address', 'attributes.text', 'tags',
]

type TimeRange = 'all' | '15m' | '1h' | '24h' | '7d' | '30d'

const TIME_RANGE_OPTIONS: { value: TimeRange; label: string; expr: string | null }[] = [
  { value: '15m', label: 'Last 15 min', expr: 'now-15m' },
  { value: '1h',  label: 'Last 1 hour', expr: 'now-1h' },
  { value: '24h', label: 'Last 24 hours', expr: 'now-24h' },
  { value: '7d',  label: 'Last 7 days', expr: 'now-7d' },
  { value: '30d', label: 'Last 30 days', expr: 'now-30d' },
  { value: 'all', label: 'All time', expr: null },
]

function textQueryFragment(q: string): AdvancedSearchRequest | undefined {
  if (!q || q === '*') return undefined
  return { query: { must: [{ multi_match: { query: q, fields: MULTI_MATCH_FIELDS } }] } }
}

function timeRangeFragment(range: TimeRange): AdvancedSearchRequest | undefined {
  const expr = TIME_RANGE_OPTIONS.find((o) => o.value === range)?.expr
  if (!expr) return undefined
  return { query: { filter: [{ range: { lastSeen: { gte: expr, lte: 'now' } } }] } }
}

function composeBody(
  filterFragment: AdvancedSearchRequest,
  q: string,
  range: TimeRange,
): AdvancedSearchRequest {
  const must: AdvancedCondition[] = [...(filterFragment.query?.must ?? [])]
  const should: AdvancedCondition[] = [...(filterFragment.query?.should ?? [])]
  const mustNot: AdvancedCondition[] = [...(filterFragment.query?.must_not ?? [])]
  const filter: AdvancedCondition[] = [...(filterFragment.query?.filter ?? [])]
  const text = textQueryFragment(q)
  if (text?.query?.must) must.push(...text.query.must)
  const time = timeRangeFragment(range)
  if (time?.query?.filter) filter.push(...time.query.filter)
  const query: NonNullable<AdvancedSearchRequest['query']> = {}
  if (must.length) query.must = must
  if (should.length) query.should = should
  if (mustNot.length) query.must_not = mustNot
  if (filter.length) query.filter = filter
  return {
    query: Object.keys(query).length ? query : undefined,
    aggs: filterFragment.aggs,
  }
}

export function ThreatIntelPage() {
  const { t } = useTranslation()
  const { isConfigured, isLoading: configLoading } = useTiConfigStatus()
  const feedsQuery = useTiFeeds()
  const searchAdvancedMutation = useTiSearchAdvanced()
  const lookupMutation = useTiEntityLookup()

  const [query, setQuery] = useState<string>('*')
  const [page, setPage] = useState(0)
  const [size, setSize] = useState(20)
  const [iocs, setIocs] = useState<EntitySummary[]>([])
  const [totalItems, setTotalItems] = useState(0)
  const [totalPages, setTotalPages] = useState(0)

  const [filters, setFilters] = useState<FiltersState>(EMPTY_FILTERS)
  const [lastBody, setLastBody] = useState<AdvancedSearchRequest>({})
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [timeRange, setTimeRange] = useState<TimeRange>('24h')

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
    const body = composeBody(filtersToRequest(filters), query, timeRange)
    setLastBody(body)
    searchAdvancedMutation.mutate(
      { body, limit: size, page: page + 1 },
      {
        onSuccess: (data) => {
          if (my !== seqRef.current) return
          if (data?.kind === 'not-configured') return
          if (data?.kind !== 'ok') return
          setTotalItems(data.value.items)
          setTotalPages(data.value.pages)
          setIocs((prev) =>
            mode === 'append' ? [...prev, ...data.value.results] : data.value.results
          )
        },
        onError: (e) => {
          if (my !== seqRef.current) return
          if (!isNotFound(e) || mode === 'append') return
          setIocs([])
          setTotalItems(0)
          setTotalPages(0)
        },
      }
    )
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query, page, size, isConfigured, timeRange])

  if (configLoading) return null
  if (isConfigured === false) return <NotConfiguredState />

  const actors: EntitySummary[] = iocs.filter((e) => e.type === 'threat')
  const feedsCount = feedsQuery.data?.kind === 'ok' ? feedsQuery.data.value.length : 0
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
    if (!hasMore || searchAdvancedMutation.isPending) return
    modeRef.current = 'append'
    setPage((p) => p + 1)
  }

  const handleFiltersApply = (request: AdvancedSearchRequest) => {
    const body = composeBody(request, query, timeRange)
    modeRef.current = 'replace'
    setPage(0)
    setLastBody(body)
    setIocs([])
    setTotalItems(0)
    setTotalPages(0)
    const my = ++seqRef.current
    searchAdvancedMutation.mutate(
      { body, limit: size, page: 1 },
      {
        onSuccess: (data) => {
          if (my !== seqRef.current) return
          if (data?.kind === 'not-configured') return
          if (data?.kind !== 'ok') return
          setTotalItems(data.value.items)
          setTotalPages(data.value.pages)
          setIocs(data.value.results)
        },
      }
    )
  }

  const handleLookup = (input: { type: string; value: string }) => {
    lookupMutation.mutate(
      { type: input.type, value: input.value },
      {
        onSuccess: (result) => {
          if (result?.kind !== 'ok') return
          setOpenIoc(result.value.id)
        },
      }
    )
  }

  const handleExport = async () => {
    if (!totalItems || isExporting) return
    setIsExporting(true)
    try {
      const res = await threatIntelHttpService.searchAdvanced(lastBody, { limit: totalItems, page: 1 })
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
        <LookupBar
          onSearch={handleSearch}
          onLookup={handleLookup}
          isPending={searchAdvancedMutation.isPending}
          isLookupPending={lookupMutation.isPending}
        />
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
                ? t('threatIntel.toolbar.searchPlaceholders.feeds')
                : tab === 'actors'
                  ? t('threatIntel.toolbar.searchPlaceholders.actors')
                  : t('threatIntel.toolbar.searchPlaceholders.iocs')
            }
            value={uiSearch}
            onChange={(e) => setUiSearch(e.target.value)}
            className="h-9 pl-9"
          />
        </div>
        {tab === 'iocs' && (
          <>
            <div className="relative">
              <Clock
                size={14}
                className="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground"
              />
              <select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value as TimeRange)}
                className="h-8 rounded-md border border-border bg-background pl-7 pr-2 text-xs"
              >
                {TIME_RANGE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setFiltersOpen((o) => !o)}
            >
              <ListFilter size={14} className="mr-2" />
              {t('threatIntel.toolbar.filters')}
            </Button>
          </>
        )}
        <button
          title={t('threatIntel.toolbar.refresh')}
          onClick={handleSearch.bind(null, query)}
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} />
        </button>
      </div>

      {filtersOpen && tab === 'iocs' && (
        <div className="mt-3">
          <FiltersPanel
            value={filters}
            onChange={setFilters}
            onApply={handleFiltersApply}
            onClear={() => {
              setFilters(EMPTY_FILTERS)
              handleFiltersApply({})
            }}
            onClose={() => setFiltersOpen(false)}
          />
        </div>
      )}

      <div className="mt-3">
        {tab === 'iocs' && (
          <IocTable
            iocs={iocs}
            onOpen={setOpenIoc}
            isLoading={searchAdvancedMutation.isPending}
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
