import { useEffect, useState } from 'react'
import { AlertTriangle, Flame, Loader2, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Pagination } from '@/shared/components/ui/pagination'
import { useIncidentsList } from '../hooks/use-incidents-list'
import { useIncidentSync } from '../hooks/use-incident-sync'
import type { Incident, IncidentStatus } from '../types/incident.types'
import { IncidentsStatCards } from '../components/incidents-stat-cards'
import { IncidentsToolbar } from '../components/incidents-toolbar'
import { IncidentsLayoutToggle, type IncidentsLayout } from '../components/incidents-layout-toggle'
import { IncidentsTable } from '../components/incidents-table'
import { IncidentsBoard } from '../components/incidents-board'
import { IncidentDrawer } from '../components/incident-drawer'
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
  const [pageSize, setPageSize] = useState(20)
  const [open, setOpen] = useState<Incident | null>(null)

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

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 py-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-semibold">
            <Flame size={18} strokeWidth={1.75} className="text-red-500" />
            {t('incidents.title')}
            <span className="text-sm font-normal text-muted-foreground">· {total}</span>
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">{t('incidents.subtitle')}</p>
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
        <>
          <IncidentsTable incidents={incidents} onOpen={setOpen} />
          <div className="mt-3 shrink-0">
            <Pagination
              page={page}
              pageSize={pageSize}
              total={total}
              onPageChange={setPage}
              onPageSizeChange={(s) => {
                setPageSize(s)
                setPage(0)
              }}
            />
          </div>
        </>
      )}

      {open && (
        <IncidentDrawer incident={open} onClose={() => setOpen(null)} onChanged={(id) => void refreshAndSync(id)} />
      )}
    </div>
  )
}
