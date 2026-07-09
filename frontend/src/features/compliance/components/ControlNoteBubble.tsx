import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, MessageSquare, MessageSquareText } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import type { ReportControlRow } from '../types/compliance.types'

/**
 * Notes icon on a control row. Filled when a note exists; click opens a floating
 * bubble anchored to the icon with the note text + a textarea to edit or clear it.
 * Stops propagation so it doesn't open the row's drawer.
 */
export function ControlNoteBubble({
  frameworkKey,
  row,
  onSaved,
  className,
}: {
  frameworkKey: string
  row: ReportControlRow
  onSaved?: () => void
  className?: string
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [anchor, setAnchor] = useState<DOMRect | null>(null)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const btnRef = useRef<HTMLButtonElement | null>(null)

  const has = !!row.note

  useEffect(() => {
    if (!open) return
    const onEsc = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    window.addEventListener('keydown', onEsc)
    return () => window.removeEventListener('keydown', onEsc)
  }, [open])

  const openBubble = (e: React.MouseEvent) => {
    e.stopPropagation()
    setAnchor(e.currentTarget.getBoundingClientRect())
    setDraft(row.note ?? '')
    setOpen(true)
  }

  const save = async () => {
    if (saving) return
    setSaving(true)
    try {
      const value = draft.trim()
      if (value === '') {
        if (has) await complianceService.clearControlNote(frameworkKey, row.controlId)
      } else {
        await complianceService.setControlNote(frameworkKey, row.controlId, value)
      }
      onSaved?.()
      setOpen(false)
    } catch (e) {
      toast.error(e instanceof ComplianceHttpError ? e.message : t('compliance.noteError', { defaultValue: 'Failed to save note' }))
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <button
        ref={btnRef}
        type="button"
        onClick={openBubble}
        title={t('compliance.notes.title', { defaultValue: 'Notes' })}
        className={cn(
          'rounded p-1 text-muted-foreground  hover:bg-muted hover:text-foreground',
          has && 'text-sky-500 hover:text-sky-400',
          className,
        )}
      >
        {has ? <MessageSquareText size={22} /> : <MessageSquare size={22} />}
      </button>

      {open && anchor && createPortal(
        <div
          className="fixed inset-0 z-50"
          onClick={() => setOpen(false)}
        >
          <div
            className="fixed w-80 rounded-lg border border-border bg-card p-3 shadow-2xl"
            style={bubblePosition(anchor)}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
              {t('compliance.notes.title', { defaultValue: 'Notes' })} — <span className="font-mono normal-case">{row.controlId}</span>
            </div>
            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              rows={5}
              autoFocus
              placeholder={t('compliance.notes.placeholder', { defaultValue: 'Add a note…' })}
              className="w-full resize-none rounded-md border border-input bg-background/60 p-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring"
            />
            <div className="mt-2 flex items-center justify-end gap-2">
              <Button size="sm" variant="ghost" onClick={() => setOpen(false)} disabled={saving}>
                {t('compliance.notes.cancel', { defaultValue: 'Cancel' })}
              </Button>
              <Button size="sm" onClick={() => void save()} disabled={saving}>
                {saving ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null}
                {t('compliance.notes.save', { defaultValue: 'Save' })}
              </Button>
            </div>
          </div>
        </div>,
        document.body,
      )}
    </>
  )
}

// bubblePosition places the bubble under the icon, flipping when it would overflow.
function bubblePosition(rect: DOMRect): React.CSSProperties {
  const W = 320
  const margin = 8
  const left = Math.min(Math.max(margin, rect.left), window.innerWidth - W - margin)
  const top = rect.bottom + margin
  return { top, left }
}
