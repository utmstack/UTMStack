import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'
import type { FlowNode } from '../types/soar.types'
import { InsertFieldMenu } from './InsertFieldMenu'

interface MailParams {
  to?: string
  cc?: string
  subject?: string
  body?: string
}

interface Props {
  nodeId: string
  nodes: Record<string, FlowNode>
  params: unknown
  readOnly?: boolean
  onChange: (params: MailParams) => void
}

// ponytail: comma-separated to/cc — backend splits + trims. No dedicated
// address chip UI; $()-templates work in every field, InsertFieldMenu drives
// caret-insert into the body (matches shell/paramsJson pattern).
export function MailParamsEditor({ nodeId, nodes, params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(params)
  const bodyRef = useRef<HTMLTextAreaElement>(null)

  const insertIntoBody = (token: string) => {
    const el = bodyRef.current
    const cur = p.body ?? ''
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    onChange({ ...p, body: next })
    requestAnimationFrame(() => {
      const el2 = bodyRef.current
      if (!el2) return
      el2.focus()
      const pos = start + token.length
      el2.setSelectionRange(pos, pos)
    })
  }

  return (
    <div className="space-y-2">
      <Field label={t('soar.editor.canvas.mail.to')}>
        <Input
          value={p.to ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, to: e.target.value })}
          placeholder="alice@example.com, bob@example.com"
          className="h-8 font-mono text-[11px]"
        />
      </Field>
      <Field label={t('soar.editor.canvas.mail.cc')}>
        <Input
          value={p.cc ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, cc: e.target.value })}
          placeholder="carol@example.com"
          className="h-8 font-mono text-[11px]"
        />
      </Field>
      <Field label={t('soar.editor.canvas.mail.subject')}>
        <Input
          value={p.subject ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, subject: e.target.value })}
          placeholder={t('soar.editor.canvas.mail.subjectPlaceholder')}
          className="h-8 text-xs"
        />
      </Field>
      <Field label={t('soar.editor.canvas.mail.body')}>
        {!readOnly && (
          <div className="mb-1 flex flex-wrap items-center gap-1.5">
            <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoBody} />
          </div>
        )}
        <textarea
          ref={bodyRef}
          value={p.body ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, body: e.target.value })}
          rows={8}
          placeholder={t('soar.editor.canvas.mail.bodyPlaceholder')}
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

function normalize(params: unknown): MailParams {
  if (!params || typeof params !== 'object') return {}
  return params as MailParams
}
