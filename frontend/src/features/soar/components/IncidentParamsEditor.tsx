import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'
import type { FlowNode } from '../types/soar.types'
import { InsertFieldMenu } from './InsertFieldMenu'

interface IncidentParams {
  name?: string
  description?: string
}

interface Props {
  nodeId: string
  nodes: Record<string, FlowNode>
  params: unknown
  readOnly?: boolean
  onChange: (params: IncidentParams) => void
}

// ponytail: name + description only — alert identity comes from the exec's
// AlertID and the interpolation bag on the backend. InsertFieldMenu drives
// caret-insert of $()-tokens into both fields (matches shell/mail patterns).
export function IncidentParamsEditor({ nodeId, nodes, params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(params)
  const nameRef = useRef<HTMLInputElement>(null)
  const descRef = useRef<HTMLTextAreaElement>(null)

  const insertInto = <T extends HTMLInputElement | HTMLTextAreaElement>(
    el: T | null,
    cur: string,
    commit: (next: string) => void,
    token: string,
  ) => {
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    commit(next)
    requestAnimationFrame(() => {
      if (!el) return
      el.focus()
      const pos = start + token.length
      el.setSelectionRange(pos, pos)
    })
  }

  const insertIntoName = (token: string) =>
    insertInto(nameRef.current, p.name ?? '', (next) => onChange({ ...p, name: next }), token)
  const insertIntoDescription = (token: string) =>
    insertInto(descRef.current, p.description ?? '', (next) => onChange({ ...p, description: next }), token)

  return (
    <div className="space-y-2">
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.incident.name')}
        </label>
        {!readOnly && (
          <div className="mb-1 flex flex-wrap items-center gap-1.5">
            <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoName} />
          </div>
        )}
        <Input
          ref={nameRef}
          value={p.name ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, name: e.target.value })}
          placeholder={t('soar.editor.canvas.incident.namePlaceholder')}
          className="h-8 text-xs"
        />
      </div>
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.incident.description')}
        </label>
        {!readOnly && (
          <div className="mb-1 flex flex-wrap items-center gap-1.5">
            <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoDescription} />
          </div>
        )}
        <textarea
          ref={descRef}
          value={p.description ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, description: e.target.value })}
          rows={4}
          placeholder={t('soar.editor.canvas.incident.descriptionPlaceholder')}
          className="w-full rounded-md border border-input bg-background px-2 py-1.5 text-[11px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>
    </div>
  )
}

function normalize(params: unknown): IncidentParams {
  if (!params || typeof params !== 'object') return {}
  return params as IncidentParams
}
