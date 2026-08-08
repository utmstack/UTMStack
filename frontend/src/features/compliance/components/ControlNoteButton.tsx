import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, MessageSquare, MessageSquareText } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import type { ControlRow, Report } from '../types/compliance.types'

/**
 * Annotates a control without touching its verdict.
 *
 * Writing a note and overriding the engine are different acts: "we know, here
 * is the remediation ticket" claims nothing about compliance. Having to pick a
 * status to record one is how a row ends up carrying a verdict nobody meant to
 * give — so this sends the note alone, and the row keeps tracking the engine.
 *
 * It lands in the same place a verdict does: the control row inside the report.
 */
export function ControlNoteButton({
  frameworkKey,
  row,
  onSaved,
  className,
}: {
  frameworkKey: string
  row: ControlRow
  onSaved?: (updated: Report) => void
  className?: string
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState('')
  const [busy, setBusy] = useState(false)

  const start = () => {
    setDraft(row.note ?? '')
    setOpen(true)
  }

  const save = async () => {
    if (!draft.trim()) return
    setBusy(true)
    try {
      const updated = await complianceService.editControl(frameworkKey, row.controlId, { note: draft.trim() })
      setOpen(false)
      onSaved?.(updated)
    } catch (e) {
      toast.error(
        e instanceof ComplianceHttpError
          ? e.message
          : t('compliance.noteError', { defaultValue: 'Could not save the note' }),
      )
    } finally {
      setBusy(false)
    }
  }

  const Icon = row.note ? MessageSquareText : MessageSquare

  return (
    <>
      <button
        onClick={(e) => {
          e.stopPropagation()
          start()
        }}
        title={row.note || t('compliance.addNote', { defaultValue: 'Add a note' })}
        className={cn(
          'rounded p-1 transition-colors',
          row.note ? 'text-sky-500 hover:text-sky-400' : 'text-muted-foreground/50 hover:text-foreground',
          className,
        )}
      >
        <Icon size={13} />
      </button>

      {open && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
          onClick={(e) => {
            e.stopPropagation()
            setOpen(false)
          }}
        >
          <div
            className="w-full max-w-md overflow-hidden rounded-xl border border-border bg-card shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <header className="border-b border-border px-5 py-4">
              <h2 className="text-base font-semibold">{t('compliance.noteModal.title', { defaultValue: 'Note' })}</h2>
              <p className="mt-0.5 font-mono text-[11px] text-muted-foreground">{row.controlId}</p>
            </header>
            <div className="px-5 py-4">
              <p className="mb-2 text-xs text-muted-foreground">
                {t('compliance.noteModal.description', {
                  defaultValue: 'Recorded on the report. It does not change the control’s status.',
                })}
              </p>
              <textarea
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                rows={5}
                autoFocus
                maxLength={4000}
                placeholder={t('compliance.noteModal.placeholder', {
                  defaultValue: 'Remediation in progress — ticket OPS-4412, due end of quarter.',
                })}
                className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            </div>
            <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/20 px-5 py-3">
              <Button variant="outline" size="sm" onClick={() => setOpen(false)} disabled={busy}>
                {t('compliance.reasonModal.cancel')}
              </Button>
              <Button size="sm" onClick={() => void save()} disabled={busy || !draft.trim()}>
                {busy && <Loader2 size={13} className="mr-1.5 animate-spin" />}
                {t('compliance.reasonModal.confirm')}
              </Button>
            </footer>
          </div>
        </div>
      )}
    </>
  )
}
