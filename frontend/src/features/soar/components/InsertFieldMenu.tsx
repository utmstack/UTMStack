import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip'
import { ALERT_FIELDS } from '../lib/alert-fields'
import { enrichmentAncestors } from '../lib/ancestors'
import type { FlowNode } from '../types/soar.types'

interface Props {
  /** Full flow node map — needed to walk ancestors. */
  nodes: Record<string, FlowNode>
  /** Which node is currently being edited. Its own outputs are excluded from
   *  the menu (you can't reference yourself). */
  currentNodeId: string
  /** Called with a ready-to-paste token, e.g. `$(alert.foo)` or `$(geoip.country)`. */
  onInsert: (token: string) => void
}

/** Dropdown that lists every value available to `$( ... )` at this node:
 *  alert fields (always in scope) plus one section per enrichment ancestor.
 *  Ancestors with statically-declared fields (currently `select`) get one
 *  item per field; the rest offer a generic `$(id.field)` placeholder the
 *  user renames after clicking. */
export function InsertFieldMenu({ nodes, currentNodeId, onInsert }: Props) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const ancestors = enrichmentAncestors(nodes, currentNodeId)

  const pick = (token: string) => {
    onInsert(token)
    setOpen(false)
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 text-[10px] hover:bg-muted"
      >
        {t('soar.editor.canvas.insertField')} <ChevronDown size={10} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 max-h-72 w-72 overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-lg">
          <SectionHeader>{t('soar.editor.canvas.alert')}</SectionHeader>
          {ALERT_FIELDS.map((af) => (
            <MenuItem
              key={`alert:${af.field}`}
              label={af.label}
              path={`alert.${af.field}`}
              token={`$(alert.${af.field})`}
              onClick={() => pick(`$(alert.${af.field})`)}
            />
          ))}

          {ancestors.length > 0 && <SectionDivider />}

          {ancestors.map((a) => (
            <div key={a.nodeId}>
              <SectionHeader>
                <span className="font-mono">{a.nodeId}</span>
                <span className="ml-1 text-[9px] uppercase text-muted-foreground/60">{a.executor}</span>
              </SectionHeader>
              {a.fields.length > 0 ? (
                a.fields.map((f) => (
                  <MenuItem
                    key={`${a.nodeId}:${f}`}
                    label={f}
                    path={`${a.nodeId}.${f}`}
                    token={`$(${a.nodeId}.${f})`}
                    onClick={() => pick(`$(${a.nodeId}.${f})`)}
                  />
                ))
              ) : (
                <MenuItem
                  label={t('soar.editor.canvas.wholeOutput')}
                  path={`${a.nodeId}.*`}
                  token={`$(${a.nodeId}.)`}
                  onClick={() => pick(`$(${a.nodeId}.)`)}
                />
              )}
            </div>
          ))}

          {ancestors.length === 0 && (
            <div className="mt-1 border-t border-border px-3 pt-2 text-[10px] italic text-muted-foreground">
              {t('soar.editor.canvas.noAncestorsHint')}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function SectionHeader({ children }: { children: React.ReactNode }) {
  return (
    <div className="sticky top-0 bg-popover px-3 py-1 text-[9px] font-semibold uppercase tracking-wider text-muted-foreground">
      {children}
    </div>
  )
}

function SectionDivider() {
  return <div className="my-1 border-t border-border" />
}

function MenuItem({ label, path, token, onClick }: { label: string; path: string; token: string; onClick: () => void }) {
  const { t } = useTranslation()
  return (
    <Tooltip delayDuration={200}>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={onClick}
          className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-[11px] hover:bg-muted"
        >
          <span className="truncate">{label}</span>
          <span className="shrink-0 font-mono text-[9px] text-muted-foreground">{path}</span>
        </button>
      </TooltipTrigger>
      <TooltipContent side="left" className="max-w-xs">
        <div className="text-[11px]">{label}</div>
        <div className="mt-0.5 font-mono text-[10px] opacity-80">{t('soar.editor.canvas.inserts')} {token}</div>
      </TooltipContent>
    </Tooltip>
  )
}
