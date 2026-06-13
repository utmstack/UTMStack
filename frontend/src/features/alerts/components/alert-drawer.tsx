import { useState } from 'react'
import { ChevronDown, Flame, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { SEV_META, ST_META, TS, absTime, flagEmoji, relativeTime, riskOf, sevKey, statusKey } from '../lib/alert-meta'
import { combineUserNote, isAiNote, parseAiNote, userNotePart } from '../lib/ai-note'
import { STATUS_BY_INT, STATUS_INT, type Alert, type AlertTag, type Side, type StatusKey } from '../types/alert.types'
import { useRelatedLogs } from '../hooks/use-related-logs'
import { Menu, Row, Section, TagChip } from './ui-primitives'
import { AlertTagEditor } from './alert-tag-editor'
import { AlertAiAssessment } from './alert-ai-assessment'
import { AlertRelatedEvents } from './alert-related-events'

type Tab = 'summary' | 'parties' | 'events' | 'history'

export function AlertDrawer({
  alert: a,
  tagCatalog,
  onClose,
  onStatus,
  onTags,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
  onIncident,
  onNotes,
}: {
  alert: Alert
  tagCatalog: AlertTag[]
  onClose: () => void
  onStatus: (status: number, observation: string, fp: boolean) => void
  onTags: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
  onIncident: () => void
  onNotes: (notes: string) => void
}) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('summary')
  const { loading: loadingRelated, viewRelatedLogs } = useRelatedLogs()
  // `notes` stores only the analyst's own free text; any AI assessment block is
  // kept separate so the SOC AI note is preserved when the analyst saves.
  const [notes, setNotes] = useState(userNotePart(a.notes))
  // The SOC AI agent may write its assessment into notes OR statusObservation.
  const aiNote = parseAiNote(a.notes) ?? parseAiNote(a.statusObservation)
  const sm = SEV_META[sevKey(a)]
  const stm = ST_META[statusKey(a)]
  const smLabel = t(`alerts.severity.${sevKey(a)}`)
  const stmLabel = t(`alerts.status.${statusKey(a)}`)
  const tags = a.tags ?? []

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', sm.pill)}>
                  {smLabel}
                </span>
                <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', stm.pill)}>
                  {stmLabel}
                </span>
                {a.category && <span>· {a.category}</span>}
                {a.isIncident && (
                  <span className="rounded bg-red-500/15 px-1 py-0.5 text-[9px] font-semibold uppercase text-red-500">
                    {t('alerts.badge.incident')}
                  </span>
                )}
              </div>
              <h2 className="mt-1 text-xl font-semibold">{a.name || '—'}</h2>
              <div className="mt-1 flex items-center gap-2 font-mono text-[11px] text-muted-foreground">
                {a.technique && <span>{a.technique}</span>}
                {a.dataSource && <span>· {a.dataSource}</span>}
                <span>· {absTime(a[TS])}</span>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          {/* Actions */}
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Menu trigger={<>{t('alerts.drawer.setStatus')} <ChevronDown size={12} /></>}>
              {(['open', 'in_review', 'completed'] as StatusKey[]).map((k) => (
                <button
                  key={k}
                  onClick={() => onStatus(STATUS_INT[k], '', false)}
                  className="block w-full px-3 py-1.5 text-left text-sm hover:bg-muted"
                >
                  {t(`alerts.status.${k}`)}
                </button>
              ))}
              <button
                onClick={() => onStatus(STATUS_INT.completed, 'Marked as false positive', true)}
                className="block w-full border-t border-border px-3 py-1.5 text-left text-sm text-muted-foreground hover:bg-muted"
              >
                {t('alerts.drawer.completeFalsePositive')}
              </button>
            </Menu>
            <AlertTagEditor
              tags={tags}
              catalog={tagCatalog}
              onTags={onTags}
              onCreateTag={onCreateTag}
              onUpdateTag={onUpdateTag}
              onDeleteTag={onDeleteTag}
            />
            <Button size="sm" variant="outline" onClick={onIncident}>
              <Flame size={13} className="mr-1.5 text-red-500" /> {t('alerts.drawer.addToIncident')}
            </Button>
          </div>

          {tags.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1.5">
              {tags.map((tg) => (
                <TagChip key={tg} name={tg} catalog={tagCatalog} />
              ))}
            </div>
          )}
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          {(['summary', 'parties', 'events', 'history'] as Tab[]).map((id) => (
            <button
              key={id}
              onClick={() => setTab(id)}
              className={cn(
                'relative px-3 py-2.5 text-xs capitalize transition-colors',
                tab === id ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {id === 'events'
                ? `${t('alerts.drawer.tabs.events')}${a.events?.length ? ` (${a.events.length})` : ''}`
                : t(`alerts.drawer.tabs.${id}`)}
              {tab === id && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          ))}
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/10 p-6">
          {tab === 'summary' && (
            <div className="space-y-4">
              {a.isIncident && a.incidentDetail && (
                <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-4">
                  <div className="flex items-center gap-2 text-xs font-medium text-red-600 dark:text-red-300">
                    <Flame size={13} /> {t('alerts.drawer.partOfIncident')}
                  </div>
                  <dl className="mt-2 grid grid-cols-[120px_1fr] gap-y-1.5 text-xs">
                    {a.incidentDetail.incidentName && (
                      <Row k={t('alerts.drawer.incidentDetail.incident')}>{a.incidentDetail.incidentName}</Row>
                    )}
                    {a.incidentDetail.incidentId != null && (
                      <Row k={t('alerts.drawer.incidentDetail.id')}>
                        <span className="font-mono">#{String(a.incidentDetail.incidentId)}</span>
                      </Row>
                    )}
                    {a.incidentDetail.createdBy && (
                      <Row k={t('alerts.drawer.incidentDetail.createdBy')}>{a.incidentDetail.createdBy}</Row>
                    )}
                    {a.incidentDetail.creationDate && (
                      <Row k={t('alerts.drawer.incidentDetail.created')}>{absTime(a.incidentDetail.creationDate)}</Row>
                    )}
                  </dl>
                </div>
              )}
              {a.description && (
                <Section title={t('alerts.drawer.section.description')}>
                  <p className="text-xs leading-relaxed text-muted-foreground">{a.description}</p>
                </Section>
              )}
              {a.solution && (
                <Section title={t('alerts.drawer.section.solution')}>
                  <p className="text-xs leading-relaxed text-muted-foreground">{a.solution}</p>
                </Section>
              )}
              {(a.reference ?? []).length > 0 && (
                <Section title={t('alerts.drawer.section.references')}>
                  <ul className="space-y-1 text-xs">
                    {(a.reference ?? []).map((r) => (
                      <li key={r}>
                        <a
                          href={r}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="break-all text-primary hover:underline"
                        >
                          {r}
                        </a>
                      </li>
                    ))}
                  </ul>
                </Section>
              )}
              <Section title={t('alerts.drawer.section.details')}>
                <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                  <Row k={t('alerts.drawer.details.risk')}>{riskOf(a)}</Row>
                  <Row k={t('alerts.drawer.details.category')}>{a.category || '—'}</Row>
                  <Row k={t('alerts.drawer.details.technique')}>
                    <span className="font-mono">{a.technique || '—'}</span>
                  </Row>
                  <Row k={t('alerts.drawer.details.sensor')}>{a.dataSource || '—'}</Row>
                  <Row k={t('alerts.drawer.details.dataType')}>
                    <span className="font-mono">{a.dataType || '—'}</span>
                  </Row>
                  <Row k={t('alerts.drawer.details.alertId')}>
                    <span className="break-all font-mono">{a.id}</span>
                  </Row>
                  {a.statusObservation && !isAiNote(a.statusObservation) && (
                    <Row k={t('alerts.drawer.details.statusNote')}>{a.statusObservation}</Row>
                  )}
                </dl>
              </Section>
              {/* The AI SOC assessment is rendered read-only above; the analyst's
                  own notes are always editable below it and saved alongside the
                  AI block (which is preserved on save). */}
              {aiNote && <AlertAiAssessment note={aiNote} />}
              <Section title={aiNote ? t('alerts.drawer.section.yourNotes') : t('alerts.drawer.section.notes')}>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  rows={3}
                  placeholder={t('alerts.drawer.notesPlaceholder')}
                  className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
                <div className="mt-2 flex justify-end">
                  <Button
                    size="sm"
                    disabled={notes.trim() === userNotePart(a.notes)}
                    onClick={() => onNotes(combineUserNote(notes, a.notes))}
                  >
                    {t('alerts.drawer.saveNotes')}
                  </Button>
                </div>
              </Section>
            </div>
          )}
          {tab === 'parties' && (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <PartyCard title={t('alerts.party.source')} ep={a.target} />
              <PartyCard title={t('alerts.party.adversary')} ep={a.adversary} accent />
            </div>
          )}
          {tab === 'events' && (
            <AlertRelatedEvents events={a.events ?? []} onViewAll={() => void viewRelatedLogs(a)} loadingAll={loadingRelated} />
          )}
          {tab === 'history' && <HistoryTab alert={a} />}
        </div>
      </div>
    </div>
  )
}

function PartyCard({ title, ep, accent }: { title: string; ep?: Side; accent?: boolean }) {
  const { t } = useTranslation()
  return (
    <div className={cn('rounded-lg border bg-card p-4', accent ? 'border-red-500/30' : 'border-border')}>
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {!ep ? (
        <p className="text-xs text-muted-foreground">{t('alerts.party.noData')}</p>
      ) : (
        <dl className="grid grid-cols-[80px_1fr] gap-y-2 text-xs">
          {ep.ip && <Row k={t('alerts.party.ip')}><span className="font-mono">{ep.ip}</span></Row>}
          {ep.host && <Row k={t('alerts.party.host')}><span className="font-mono">{ep.host}</span></Row>}
          {ep.user && <Row k={t('alerts.party.user')}><span className="font-mono">{ep.user}</span></Row>}
          {ep.domain && <Row k={t('alerts.party.domain')}><span className="font-mono">{ep.domain}</span></Row>}
          {ep.geolocation?.country && (
            <Row k={t('alerts.party.country')}>
              {flagEmoji(ep.geolocation.countryCode)} {ep.geolocation.country}
              {ep.geolocation.city ? ` · ${ep.geolocation.city}` : ''}
            </Row>
          )}
        </dl>
      )}
    </div>
  )
}

function HistoryTab({ alert: a }: { alert: Alert }) {
  const { t } = useTranslation()
  const history = a.history ?? []
  if (history.length === 0) return <p className="text-xs text-muted-foreground">{t('alerts.history.empty')}</p>
  return (
    <div className="space-y-2">
      {history
        .slice()
        .reverse()
        .map((h, i) => {
          const detail = historyDetail(h, t)
          return (
            <div key={i} className="rounded-md border border-border bg-card px-3 py-2 text-xs">
              <div className="flex items-center justify-between gap-2">
                <span className="font-medium">{actionLabel(h.action, t)}</span>
                <span className="font-mono text-[10px] text-muted-foreground">{relativeTime(h.timestamp)}</span>
              </div>
              {(detail || h.user) && (
                <div className="mt-0.5 text-muted-foreground">
                  {detail}
                  {h.user && (
                    <span>
                      {detail ? ' · ' : ''}
                      {t('alerts.history.by', { user: h.user })}
                    </span>
                  )}
                </div>
              )}
            </div>
          )
        })}
    </div>
  )
}

const ACTION_KEYS = ['UPDATE_STATUS', 'UPDATE_TAGS', 'UPDATE_NOTES', 'UPDATE_SOLUTION', 'MARK_AS_INCIDENT']
function actionLabel(a: string | undefined, t: TFunction) {
  if (a && ACTION_KEYS.includes(a)) return t(`alerts.history.actions.${a}`)
  return (a ?? 'change').replace(/_/g, ' ').toLowerCase()
}

// A clean one-line detail — never dumps the raw AI assessment or the newValue JSON.
function historyDetail(h: { message?: string; newValue?: string }, t: TFunction): string {
  const msg = (h.message ?? '').trim()
  if (msg && !isAiNote(msg) && !msg.startsWith('{')) return msg
  try {
    const v = JSON.parse(h.newValue || '{}') as Record<string, unknown>
    const parts: string[] = []
    if (typeof v.status === 'number') {
      const stKey = STATUS_BY_INT[v.status]
      parts.push(
        t('alerts.history.detail.statusTo', { status: stKey ? t(`alerts.status.${stKey}`) : String(v.status) })
      )
    }
    if (v.tags != null)
      parts.push(
        t('alerts.history.detail.tags', {
          tags: Array.isArray(v.tags) ? (v.tags as string[]).join(', ') : String(v.tags),
        })
      )
    if (typeof v.statusObservation === 'string' && v.statusObservation)
      parts.push(
        isAiNote(v.statusObservation)
          ? t('alerts.history.detail.aiAdded')
          : t('alerts.history.detail.observationAdded')
      )
    if (v.notes != null && !isAiNote(String(v.notes))) parts.push(t('alerts.history.detail.notesUpdated'))
    return parts.join(' · ')
  } catch {
    return ''
  }
}
