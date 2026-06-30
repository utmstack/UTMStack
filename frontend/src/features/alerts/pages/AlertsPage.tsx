import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { FILTER_OPS, TS } from '../lib/alert-meta'
import { alertToRuleConditions } from '../lib/tagging-rule-meta'
import {
  STATUS_INT,
  SEVERITY_INT,
  type Alert,
  type CustomFilter,
  type FilterType,
  type SeverityKey,
  type StatusKey,
  type StatusTab,
} from '../types/alert.types'
import { useAlertsList } from '../hooks/use-alerts-list'
import { useAlertStats } from '../hooks/use-alert-stats'
import { useAlertTagCatalog } from '../hooks/use-alert-tag-catalog'
import { useAlertMutations } from '../hooks/use-alert-mutations'
import { AlertsHeader, type AlertsView } from '../components/alerts-header'
import { AlertsToolbar } from '../components/alerts-toolbar'
import { AlertsFilterBar } from '../components/alerts-filter-bar'
import { AlertsStatusTabs } from '../components/alerts-status-tabs'
import { AlertsVolumeCard } from '../components/alerts-volume-card'
import { AlertsBreakdownCard } from '../components/alerts-breakdown-card'
import { AlertsBulkBar } from '../components/alerts-bulk-bar'
import { AlertsTableHeader, AlertRow } from '../components/alerts-table'
import { EchoesTimeline } from '../components/echoes-timeline'
import { AlertDrawer } from '../components/alert-drawer'
import { AlertIncidentModal } from '../components/alert-incident-modal'
import { Center } from '../components/ui-primitives'

export function AlertsPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<AlertsView>('alerts')
  const [statusTab, setStatusTab] = useState<StatusTab>('all')
  const [severity, setSeverity] = useState<SeverityKey | 'all'>('all')
  const [range, setRange] = useState<TimeRange>(() => presetRange('7d'))
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [tagFilter, setTagFilter] = useState<string[]>([])
  const [customFilters, setCustomFilters] = useState<CustomFilter[]>([])

  const [page, setPage] = useState(0)
  const [pageSize] = useState(50)

  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [expandedEchoes, setExpandedEchoes] = useState<Set<string>>(new Set())
  const [openAlert, setOpenAlert] = useState<Alert | null>(null)
  const [incidentTargets, setIncidentTargets] = useState<Alert[] | null>(null)

  const toggleEchoes = (id: string) =>
    setExpandedEchoes((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })

  // SOC-AI chat navigation: seed the filters + time window the agent emitted.
  const location = useLocation()
  const navigate = useNavigate()
  const seededRef = useRef(false)
  useEffect(() => {
    const state = location.state as { socaiFilters?: FilterType[]; socaiTime?: string } | null
    if (!state?.socaiFilters?.length || seededRef.current) return
    seededRef.current = true
    setCustomFilters(
      state.socaiFilters.map((f) => ({
        field: f.field,
        label: f.field,
        operator: f.operator,
        value: Array.isArray(f.value) ? f.value.join(', ') : String(f.value ?? ''),
      })),
    )
    if (state.socaiTime) setRange(presetRange(state.socaiTime))
    setPage(0)
    // Drop the router state so a refresh doesn't re-seed over the user's edits.
    navigate(location.pathname, { replace: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  // Scope filters (time + search). `parentId DOES_NOT_EXIST` hides deduplicated
  // child "echoes" — they roll up under their parent (matches the legacy default).
  const scopeFilters = useMemo<FilterType[]>(() => {
    const f: FilterType[] = [{ field: 'parentId', operator: 'DOES_NOT_EXIST' }]
    if (range.from) f.push({ field: TS, operator: 'IS_BETWEEN', value: [range.from, range.to] })
    if (debounced) f.push({ field: 'name', operator: 'CONTAIN', value: debounced })
    // Tag selector: match alerts carrying ANY of the chosen tags (OR semantics).
    if (tagFilter.length) f.push({ field: 'tags', operator: 'IS_ONE_OF', value: tagFilter })
    for (const cf of customFilters) {
      const needsValue = FILTER_OPS.find((o) => o.id === cf.operator)?.needsValue ?? true
      f.push({ field: cf.field, operator: cf.operator, value: needsValue ? cf.value : undefined })
    }
    return f
  }, [range.from, range.to, debounced, tagFilter, customFilters])

  // List filters add the tab (status) + severity refinement.
  const listFilters = useMemo<FilterType[]>(() => {
    const f = [...scopeFilters]
    if (statusTab !== 'all') f.push({ field: 'status', operator: 'IS', value: STATUS_INT[statusTab as StatusKey] })
    if (severity !== 'all') f.push({ field: 'severity', operator: 'IS', value: SEVERITY_INT[severity] })
    return f
  }, [scopeFilters, statusTab, severity])

  const { alerts, total, loading, error, refresh: refreshList } = useAlertsList(page, pageSize, listFilters)
  const { sevCounts, statusCounts, timeline, openCount, refresh: refreshStats } = useAlertStats(scopeFilters, range.interval)
  const { tagCatalog, createTag, updateTag, deleteTag } = useAlertTagCatalog((deletedName) =>
    setTagFilter((tf) => tf.filter((tn) => tn !== deletedName))
  )

  const refresh = useCallback(() => {
    refreshList()
    refreshStats()
  }, [refreshList, refreshStats])

  const { applyStatus, applyTags, updateNotes, updateAssignee, exportCsv } = useAlertMutations({
    refresh,
    clearSelection: () => setSelected(new Set()),
    openAlert,
    setOpenAlert,
  })

  const selectedAlerts = useMemo(() => alerts.filter((a) => selected.has(a.id)), [alerts, selected])

  const toggleSel = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  const allChecked = alerts.length > 0 && alerts.every((a) => selected.has(a.id))
  const togglePage = () =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (allChecked) alerts.forEach((a) => next.delete(a.id))
      else alerts.forEach((a) => next.add(a.id))
      return next
    })

  const addFilter = (cf: CustomFilter) => { setCustomFilters((c) => [...c, cf]); setPage(0) }
  const removeFilter = (i: number) => { setCustomFilters((c) => c.filter((_, idx) => idx !== i)); setPage(0) }

  // Register a new tag in the catalog, then apply it to the given alerts.
  const createAndApplyTag = async (ids: string[], tagName: string, tagColor: string, existing: string[]) => {
    const tag = await createTag(tagName, tagColor)
    if (tag) await applyTags(ids, [...existing, tagName])
  }

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 pb-6 pt-3">
      <AlertsHeader total={total} openCount={openCount} view={view} onView={setView} />

      <div className="mt-3 shrink-0">
        <AlertsToolbar
          search={search}
          onSearch={setSearch}
          severity={severity}
          onSeverity={(s) => { setSeverity(s); setPage(0) }}
          range={range}
          onRange={(r) => { setRange(r); setPage(0) }}
          tagCatalog={tagCatalog}
          tagFilter={tagFilter}
          onTagFilter={(tags) => { setTagFilter(tags); setPage(0) }}
          onCreateTag={(name, color) => void createTag(name, color)}
          onUpdateTag={(id, name, color) => void updateTag(id, name, color)}
          onDeleteTag={(id, name) => void deleteTag(id, name)}
          onRefresh={refresh}
          onExport={() => void exportCsv(listFilters)}
          loading={loading}
          showSeverity={view === 'alerts'}
        />
      </div>

      <div className="mt-2 shrink-0">
        <AlertsFilterBar
          filters={customFilters}
          onAdd={addFilter}
          onRemove={removeFilter}
          onClear={() => { setCustomFilters([]); setPage(0) }}
        />
      </div>

      {view === 'overview' ? (
        <div className="mt-4 grid min-h-0 flex-1 grid-cols-12 gap-4">
          <div className="col-span-12 lg:col-span-7">
            <AlertsVolumeCard values={timeline} total={total} />
          </div>
          <div className="col-span-12 lg:col-span-5">
            <AlertsBreakdownCard sevCounts={sevCounts} statusCounts={statusCounts} />
          </div>
        </div>
      ) : (
        <>
          <div className="mt-4 shrink-0">
            <AlertsStatusTabs
              current={statusTab}
              counts={statusCounts}
              onChange={(s) => { setStatusTab(s); setPage(0) }}
            />
          </div>

          {selected.size > 0 && (
            <AlertsBulkBar
              count={selected.size}
              tagCatalog={tagCatalog}
              onClear={() => setSelected(new Set())}
              onStatus={(s) => void applyStatus([...selected], s)}
              onTag={(tag) => void applyTags([...selected], [tag])}
              onIncident={() => setIncidentTargets(selectedAlerts)}
            />
          )}

          <div className={`mt-3 flex min-h-0 flex-1 flex-col overflow-hidden  rounded-xl border border-border bg-card  ${loading && '[&_*]:cursor-wait'}`}
          onClickCapture={(ev)=>{
            if(loading){
              ev.stopPropagation();
              ev.preventDefault();
            }
          }}
          >
            <AlertsTableHeader allChecked={allChecked} onTogglePage={togglePage} />
            <div className="min-h-0 flex-1 overflow-y-auto ">
              {loading && alerts.length === 0 ? (
                <Center>
                  <Loader2 className="h-4 w-4 animate-spin" /> {t('alerts.list.loading')}
                </Center>
              ) : error ? (
                <Center>
                  <AlertTriangle size={16} className="text-amber-500" /> {t('alerts.list.loadError')}
                  <Button variant="outline" size="sm" className="ml-2" onClick={refresh}>
                    {t('alerts.list.retry')}
                  </Button>
                </Center>
              ) : alerts.length === 0 ? (
                <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('alerts.list.empty')}</div>
              ) : (
                alerts.map((a,index) => (
                  <Fragment key={`${a.id}-${index}`}>
                    <AlertRow
                      alert={a}
                      tagCatalog={tagCatalog}
                      checked={selected.has(a.id)}
                      expanded={expandedEchoes.has(a.id)}
                      onToggle={() => toggleSel(a.id)}
                      onOpen={() => setOpenAlert(a)}
                      onCreateRule={(alert) =>
                        navigate('/threat-management/alerts/tagging-rules', {
                          state: { createWithConditions: alertToRuleConditions(alert) },
                        })
                      }
                      onToggleEchoes={() => toggleEchoes(a.id)}
                      onStatus={(s, obs, fp) => void applyStatus([a.id], s, obs, fp)}
                    />
                    {expandedEchoes.has(a.id) && <EchoesTimeline parentId={a.id} />}
                  </Fragment>
                ))
              )}
              {alerts.length > 0 && (
                <InfiniteScrollSentinel
                  onReach={() => setPage((p) => p + 1)}
                  hasMore={alerts.length < total}
                  loading={loading}
                  endLabel={t('common.allLoaded', { count: total })}
                />
              )}
            </div>
          </div>
        </>
      )}

      {openAlert && (
        <AlertDrawer
          alert={openAlert}
          tagCatalog={tagCatalog}
          onClose={() => setOpenAlert(null)}
          onStatus={(s, obs, fp) => void applyStatus([openAlert.id], s, obs, fp)}
          onTags={(tags) => void applyTags([openAlert.id], tags)}
          onCreateTag={(name, color) => void createAndApplyTag([openAlert.id], name, color, openAlert.tags ?? [])}
          onUpdateTag={(id, name, color) => void updateTag(id, name, color)}
          onDeleteTag={(id, name) => void deleteTag(id, name)}
          onIncident={() => setIncidentTargets([openAlert])}
          onNotes={(notes) => void updateNotes(openAlert.id, notes)}
          onAssign={(assignee) => void updateAssignee(openAlert.id, assignee)}
        />
      )}

      {incidentTargets && (
        <AlertIncidentModal
          alerts={incidentTargets}
          onClose={() => setIncidentTargets(null)}
          onDone={() => {
            setIncidentTargets(null)
            setSelected(new Set())
            refresh()
          }}
        />
      )}
    </div>
  )
}
