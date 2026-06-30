import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { relativeTime } from '../lib/alert-meta'
import { isAiNote } from '../lib/ai-note'
import { STATUS_BY_INT, type Alert } from '../types/alert.types'

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

export function HistoryTab({ alert: a }: { alert: Alert }) {
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
