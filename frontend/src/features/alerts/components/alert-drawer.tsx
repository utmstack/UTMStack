import { useState } from 'react'
import { Flame, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { SEV_BADGE, ST_META, TS, absTime, riskOf, sevKey, statusKey } from '../lib/alert-meta'
import { combineUserNote, isAiNote, parseAiNote, userNotePart } from '../lib/ai-note'
import { type Alert, type AlertTag } from '../types/alert.types'
import { useRelatedLogs } from '../hooks/use-related-logs'
import { Section } from './section'
import { Row } from './row'
import { TagChip } from './tag-chip'
import { AlertTagEditor } from './alert-tag-editor'
import { AlertAiAssessment } from './alert-ai-assessment'
import { AlertRelatedEvents } from './alert-related-events'
import { StatusChangeMenu } from './status-change-menu'
import { TechniqueValue } from './technique-value'
import { AssigneeMenu } from './assignee-menu'
import { PartyCard } from './party-card'
import { HistoryTab } from './history-tab'

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
  onCreateRule,
  onIncident,
  onNotes,
  onAssign,
}: {
  alert: Alert
  tagCatalog: AlertTag[]
  onClose: () => void
  onStatus: (status: number, observation: string, fp: boolean) => void
  onTags: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: string, tagName: string, tagColor: string) => void
  onDeleteTag: (id: string, tagName: string) => void
  onCreateRule: (tg: AlertTag) => void
  onIncident: () => void
  onNotes: (notes: string) => void
  onAssign: (assignee: string) => void
}) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('summary')
  const { loading: loadingRelated, viewRelatedLogs } = useRelatedLogs()
  // `notes` stores only the analyst's own free text; any AI assessment block is
  // kept separate so the SOC AI note is preserved when the analyst saves.
  const [notes, setNotes] = useState(userNotePart(a.notes))
  // The SOC AI agent may write its assessment into notes OR statusObservation.
  const aiNote = parseAiNote(a.notes) ?? parseAiNote(a.statusObservation)
  const sk = sevKey(a)
  const stm = ST_META[statusKey(a)]
  const smLabel = t(`alerts.severity.${sk}`)
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
                <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 font-medium uppercase tracking-wide ring-1 ring-inset', SEV_BADGE[sk])}>
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
            <StatusChangeMenu status={statusKey(a)} variant="action" onStatus={onStatus} />
            <AlertTagEditor
              tags={tags}
              catalog={tagCatalog}
              onTags={onTags}
              onCreateTag={onCreateTag}
              onUpdateTag={onUpdateTag}
              onDeleteTag={onDeleteTag}
              onCreateRule={onCreateRule}
            />
            {!a.isIncident && (
              <Button size="sm" variant="outline" onClick={onIncident}>
                <Flame size={13} className="mr-1.5 text-red-500" /> {t('alerts.drawer.addToIncident')}
              </Button>
            )}
            <AssigneeMenu current={a.assignee} onAssign={onAssign} />
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
              {(a.references ?? []).length > 0 && (
                <Section title={t('alerts.drawer.section.references')}>
                  <ul className="space-y-1 text-xs">
                    {(a.references ?? []).map((r: string) => (
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
                  <Row k={t('alerts.drawer.details.severity')}>
                    <span
                      className={cn(
                        'inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ring-1 ring-inset',
                        SEV_BADGE[sk],
                      )}
                    >
                      {smLabel}
                    </span>
                  </Row>
                  <Row k={t('alerts.drawer.details.risk')}>{riskOf(a)}</Row>
                  <Row k={t('alerts.drawer.details.assignee')}>
                    {a.assignee ? a.assignee : <span className="text-muted-foreground">{t('alerts.drawer.unassigned')}</span>}
                  </Row>
                  <Row k={t('alerts.drawer.details.category')}>{a.category || '—'}</Row>
                  <Row k={t('alerts.drawer.details.technique')}>
                    <TechniqueValue technique={a.technique} />
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
