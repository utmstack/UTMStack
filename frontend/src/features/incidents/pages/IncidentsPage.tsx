import { useEffect, useMemo, useState } from 'react'
import { AlertTriangle, Loader2, Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { StatusTabs } from '@/shared/components/ui/status-tabs'
import { ALL_TIME, resolveRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import type { CustomFilter } from '@/shared/components/filters/custom-filter.types'
import { useIncidentsList } from '../hooks/use-incidents-list'
import { useIncidentSync } from '../hooks/use-incident-sync'
import { STATUSES, statusKey } from '../lib/incident-meta'
import { incidentsHttpService as svc, IncidentsHttpError } from '../services/incidents-http.service'
import type { Incident, IncidentSeverity, IncidentStatus } from '../types/incident.types'
import { IncidentsHeader, type IncidentsLayout } from '../components/incidents-header'
import { IncidentsToolbar } from '../components/incidents-toolbar'
import { ASSIGNEE_FIELD, IncidentsFilterBar } from '../components/incidents-filter-bar'
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
  const [severity, setSeverity] = useState<IncidentSeverity | 'all'>('all')
  const [range, setRange] = useState<TimeRange>(ALL_TIME)
  const [rangeTick, setRangeTick] = useState(0)
  const [customFilters, setCustomFilters] = useState<CustomFilter[]>([])
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

  // Relative windows ("last 7 days") are resolved to instants once per change
  // or refresh, not per render, or the request would differ on every render.
  const resolved = useMemo(() => resolveRange(range), [range.from, range.to, rangeTick])
  const dateFrom = resolved.from ?? ''
  const dateTo = resolved.from ? resolved.to : ''
  const assignee = customFilters.find((f) => f.field === ASSIGNEE_FIELD)?.value ?? 'all'

  const isBoard = layout === 'board'
  // The board is already grouped by status, so the status tabs don't apply to it.
  const status = isBoard ? 'all' : statusFilter
  const { incidents, total, loading, error, counts, assigneeOptions, refresh } = useIncidentsList({
    search: debounced,
    statusFilter: status,
    severity,
    assignee,
    dateFrom,
    dateTo,
    page,
    pageSize,
    isBoard,
  })

  const refreshAndSync = useIncidentSync(refresh, setOpen)
  const reload = () => {
    setRangeTick((n) => n + 1)
    refresh()
  }

  const exportCsv = async () => {
    try {
      await svc.exportCsv({
        incidentName: debounced || undefined,
        incidentStatus: status === 'all' ? undefined : status,
        incidentSeverity: severity === 'all' ? undefined : severity,
        incidentAssignedTo: assignee === 'all' ? undefined : assignee,
        createdDateStart: dateFrom || undefined,
        createdDateEnd: dateTo || undefined,
      })
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.toast.exportFailed'))
    }
  }

  // The list takes one assignee, so a new assignee filter replaces the old one.
  const addFilter = (f: CustomFilter) => {
    setCustomFilters((c) => [...c.filter((x) => x.field !== f.field), f])
    setPage(0)
  }
  const updateFilter = (i: number, f: CustomFilter) => {
    setCustomFilters((c) => c.map((x, idx) => (idx === i ? f : x)))
    setPage(0)
  }
  const removeFilter = (i: number) => {
    setCustomFilters((c) => c.filter((_, idx) => idx !== i))
    setPage(0)
  }

  const createButton = (
    <Button size="sm" onClick={() => setCreating(true)}>
      <Plus size={14} className="mr-1.5" />
      {t('incidents.toolbar.create')}
    </Button>
  )

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <IncidentsHeader
        total={total}
        openCount={counts.Open}
        layout={layout}
        onLayout={(l) => {
          setLayout(l)
          setPage(0)
        }}
      />

      <div className="mt-3 shrink-0">
        <IncidentsToolbar
          search={search}
          onSearch={setSearch}
          severity={severity}
          onSeverity={(s) => {
            setSeverity(s)
            setPage(0)
          }}
          range={range}
          onRange={(r) => {
            setRange(r)
            setPage(0)
          }}
          onExport={() => void exportCsv()}
          onRefresh={reload}
          loading={loading}
        />
      </div>

      <div className="mt-2 shrink-0">
        <IncidentsFilterBar
          filters={customFilters}
          assigneeOptions={assigneeOptions}
          onAdd={addFilter}
          onUpdate={updateFilter}
          onRemove={removeFilter}
          onClear={() => {
            setCustomFilters([])
            setPage(0)
          }}
        />
      </div>

      <div className="mt-4 shrink-0">
        {isBoard ? (
          <div className="flex justify-end">{createButton}</div>
        ) : (
          <StatusTabs
            current={statusFilter}
            onChange={(s) => {
              setStatusFilter(s)
              setPage(0)
            }}
            tabs={[
              { id: 'all' as const, label: t('incidents.statusTabs.all') },
              ...STATUSES.map((s) => ({ id: s, label: t(`incidents.status.${statusKey(s)}`), count: counts[s] })),
            ]}
            actions={createButton}
          />
        )}
      </div>

      {isBoard ? (
        error ? (
          <Center>
            <AlertTriangle size={16} className="text-amber-500" /> {t('incidents.loadError')}
            <button onClick={reload} className="ml-2 text-primary hover:underline">
              {t('incidents.retry')}
            </button>
          </Center>
        ) : loading && incidents.length === 0 ? (
          <Center>
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </Center>
        ) : incidents.length === 0 ? (
          <Center>{t('incidents.empty')}</Center>
        ) : (
          <IncidentsBoard incidents={incidents} onOpen={setOpen} />
        )
      ) : (
        <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
          <div className="min-h-0 flex-1 overflow-auto">
            <IncidentsTable
              incidents={incidents}
              onOpen={setOpen}
              onChanged={(id) => void refreshAndSync(id)}
              loading={loading}
              error={error}
              onRetry={reload}
            />
            {incidents.length > 0 && (
              <InfiniteScrollSentinel
                onReach={() => setPage((p) => p + 1)}
                hasMore={incidents.length < total}
                loading={loading}
                endLabel={t('common.allLoaded', { count: total })}
              />
            )}
          </div>
        </div>
      )}

      {open && (
        <IncidentDrawer incident={open} onClose={() => setOpen(null)} onChanged={(id) => void refreshAndSync(id)} />
      )}

      {creating && (
        <CreateIncidentWizard
          onClose={() => setCreating(false)}
          onCreated={() => {
            setCreating(false)
            refresh()
          }}
        />
      )}
    </div>
  )
}
