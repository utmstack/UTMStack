import { useEffect, useState } from 'react'
import { AlertTriangle, Flame, Loader2, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { useIncidentsList } from '../hooks/use-incidents-list'
import { useIncidentSync } from '../hooks/use-incident-sync'
import type { Incident, IncidentStatus } from '../types/incident.types'
import { IncidentsStatCards } from '../components/incidents-stat-cards'
import { IncidentsToolbar } from '../components/incidents-toolbar'
import { IncidentsLayoutToggle, type IncidentsLayout } from '../components/incidents-layout-toggle'
import { IncidentsTable } from '../components/incidents-table'
import { IncidentsBoard } from '../components/incidents-board'
import { IncidentDrawer } from '../components/incident-drawer'
import { CreateIncidentWizard } from '../components/create-incident-wizard'
import { Center } from '../components/ui-primitives'

export function IncidentsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [statusFilter, setStatusFilter] = useState<IncidentStatus | 'all'>('all')
  const [assignee, setAssignee] = useState<string>('all')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [layout, setLayout] = useState<IncidentsLayout>('table')
  const [page, setPage] = useState(0)
  const [pageSize] = useState(50)
  const [open, setOpen] = useState<Incident | null>(null)
  const [creating, setCreating] = useState(false)

  // Debounce the free-text search.
  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const isBoard = layout === 'board'
  const { incidents, total, loading, error, counts, assigneeOptions, refresh } = useIncidentsList({
    search: debounced,
    statusFilter,
    assignee,
    dateFrom,
    dateTo,
    page,
    pageSize,
    isBoard,
  })

  const refreshAndSync = useIncidentSync(refresh, setOpen)

  const setStatus = (s: IncidentStatus | 'all') => {
    setStatusFilter(s)
    setPage(0)
  }
  const hasFilters =
    !!dateFrom || !!dateTo || !!debounced || statusFilter !== 'all' || assignee !== 'all'
  const clearFilters = () => {
    setSearch('')
    setStatusFilter('all')
    setAssignee('all')
    setDateFrom('')
    setDateTo('')
    setPage(0)
  }

  if (creating) {
    return (
      <CreateIncidentWizard
        onClose={() => setCreating(false)}
        onCreated={() => {
          setCreating(false)
          refresh()
        }}
      />
    )
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Flame size={14} strokeWidth={1.75} className="text-red-500" />
          <span><span className="font-medium text-foreground">{total}</span> {t('incidents.title').toLowerCase()}</span>
        </div>
        <div className="flex items-center gap-2">
          <IncidentsLayoutToggle
            value={layout}
            onChange={(l) => {
              setLayout(l)
              setPage(0)
            }}
          />
          <button
            onClick={refresh}
            title={t('incidents.refresh')}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
          </button>
        </div>
      </header>

      <IncidentsStatCards
        counts={counts}
        statusFilter={statusFilter}
        onToggle={(s) => setStatus(statusFilter === s ? 'all' : s)}
      />

      <IncidentsToolbar
        search={search}
        onSearch={setSearch}
        statusFilter={statusFilter}
        onStatusFilter={setStatus}
        assignee={assignee}
        onAssignee={(v) => {
          setAssignee(v)
          setPage(0)
        }}
        assigneeOptions={assigneeOptions}
        dateFrom={dateFrom}
        onDateFrom={(v) => {
          setDateFrom(v)
          setPage(0)
        }}
        dateTo={dateTo}
        onDateTo={(v) => {
          setDateTo(v)
          setPage(0)
        }}
        hasFilters={hasFilters}
        onClear={clearFilters}
        onCreate={() => setCreating(true)}
      />

      {error ? (
        <Center>
          <AlertTriangle size={16} className="text-amber-500" /> {t('incidents.loadError')}
          <button onClick={refresh} className="ml-2 text-primary hover:underline">
            {t('incidents.retry')}
          </button>
        </Center>
      ) : loading && incidents.length === 0 ? (
        <Center>
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </Center>
      ) : incidents.length === 0 ? (
        <Center>{t('incidents.empty')}</Center>
      ) : isBoard ? (
        <IncidentsBoard incidents={incidents} onOpen={setOpen} />
      ) : (
        <div className="min-h-0 flex-1 overflow-y-auto">
          <IncidentsTable incidents={incidents} onOpen={setOpen} />
          <InfiniteScrollSentinel
            onReach={() => setPage((p) => p + 1)}
            hasMore={incidents.length < total}
            loading={loading}
            endLabel={t('common.allLoaded', { count: total })}
          />
        </div>
      )}

      {open && (
        <IncidentDrawer incident={open} onClose={() => setOpen(null)} onChanged={(id) => void refreshAndSync(id)} />
      )}
    </div>
  )
}
