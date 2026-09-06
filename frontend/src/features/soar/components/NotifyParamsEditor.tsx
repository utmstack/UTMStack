import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import type { FlowNode } from '../types/soar.types'
import { InsertFieldMenu } from './InsertFieldMenu'

interface NotifyParams {
  message?: string
  type?: string
}

interface Props {
  nodeId: string
  nodes: Record<string, FlowNode>
  params: unknown
  readOnly?: boolean
  onChange: (params: NotifyParams) => void
}

const LEVELS = ['INFO', 'WARNING', 'ERROR'] as const

const SELECT =
  'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

export function NotifyParamsEditor({ nodeId, nodes, params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(params)
  const messageRef = useRef<HTMLTextAreaElement>(null)

  const insertIntoMessage = (token: string) => {
    const el = messageRef.current
    const cur = p.message ?? ''
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    onChange({ ...p, message: next })
    requestAnimationFrame(() => {
      const el2 = messageRef.current
      if (!el2) return
      el2.focus()
      const pos = start + token.length
      el2.setSelectionRange(pos, pos)
    })
  }

  return (
    <div className="space-y-2">
      <Field label={t('soar.editor.canvas.notify.level')}>
        <select
          value={p.type ?? 'INFO'}
          disabled={readOnly}
          onChange={(e) => onChange({ ...p, type: e.target.value })}
          className={SELECT}
        >
          {LEVELS.map((l) => (
            <option key={l} value={l}>
              {l}
            </option>
          ))}
        </select>
      </Field>
      <Field label={t('soar.editor.canvas.notify.message')}>
        {!readOnly && (
          <div className="mb-1 flex flex-wrap items-center gap-1.5">
            <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoMessage} />
          </div>
        )}
        <textarea
          ref={messageRef}
          value={p.message ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, message: e.target.value })}
          rows={6}
          placeholder={t('soar.editor.canvas.notify.messagePlaceholder')}
          className="w-full rounded-md border border-input bg-background px-2 py-1.5 text-[11px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </Field>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}

function normalize(params: unknown): NotifyParams {
  if (!params || typeof params !== 'object') return {}
  return params as NotifyParams
}
