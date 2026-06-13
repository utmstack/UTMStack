import { useState } from 'react'
import { Send, UserCircle2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { useDateFormat } from '@/shared/lib/datetime'
import { useIncidentNotesTab } from '../hooks/use-incident-notes-tab'
import { TabEmpty, TabError, TabLoader } from './ui-primitives'

export function IncidentNotesTab({ incidentId }: { incidentId: number }) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [text, setText] = useState('')
  const { rows, error, busy, reload, add } = useIncidentNotesTab(incidentId)

  const submit = async () => {
    await add(text)
    setText('')
  }

  return (
    <div className="space-y-3">
      <div className="rounded-lg border border-border bg-card p-3">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          rows={2}
          maxLength={1000}
          placeholder={t('incidents.notes.placeholder')}
          className="w-full resize-none rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
        <div className="mt-2 flex justify-end">
          <Button size="sm" onClick={() => void submit()} disabled={!text.trim() || busy}>
            <Send size={13} className="mr-1.5" /> {t('incidents.notes.add')}
          </Button>
        </div>
      </div>
      {error ? (
        <TabError onRetry={reload} />
      ) : rows === null ? (
        <TabLoader />
      ) : rows.length === 0 ? (
        <TabEmpty>{t('incidents.notes.empty')}</TabEmpty>
      ) : (
        <ul className="space-y-2">
          {rows.map((n) => (
            <li key={n.id} className="rounded-lg border border-border bg-card p-3">
              <p className="whitespace-pre-wrap text-xs leading-relaxed">{n.noteText}</p>
              <div className="mt-1.5 flex items-center gap-2 text-[10px] text-muted-foreground">
                <UserCircle2 size={11} /> {n.noteSendBy || t('incidents.unknownUser')} ·{' '}
                {df.formatDateTime(n.noteSendDate)}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
