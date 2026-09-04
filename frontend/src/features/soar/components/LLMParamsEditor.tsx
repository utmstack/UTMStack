import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { FlowNode } from '../types/soar.types'
import { InsertFieldMenu } from './InsertFieldMenu'

interface Props {
  nodeId: string
  nodes: Record<string, FlowNode>
  params: unknown
  readOnly?: boolean
  /** executor: 'llm_enrich' | 'llm_action' — the hint differs per kind. */
  executor: string
  onChange: (params: { prompt?: string }) => void
}

// llm_enrich / llm_action params hold a single free-text prompt. This editor
// replaces the raw JSON textarea for those nodes — users see one text box,
// never `{"prompt": ...}`. For llm_enrich the backend injects the mandatory
// output contract (a JSON object with a `result` property) into the task
// itself, so nothing about the return shape is configured here; the hint
// just tells the user how children will read it.
export function LLMParamsEditor({ nodeId, nodes, params, readOnly, executor, onChange }: Props) {
  const { t } = useTranslation()
  const promptRef = useRef<HTMLTextAreaElement>(null)
  const [prompt, setPrompt] = useState(() => extractPrompt(params))

  useEffect(() => {
    setPrompt(extractPrompt(params))
  }, [params, nodeId])

  const isEnrich = executor === 'llm_enrich'

  const commit = () => {
    const trimmed = prompt.trim()
    onChange({ prompt: trimmed })
  }

  const insertIntoPrompt = (token: string) => {
    const el = promptRef.current
    const cur = prompt
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    setPrompt(next)
    requestAnimationFrame(() => {
      const el2 = promptRef.current
      if (!el2) return
      el2.focus()
      const pos = start + token.length
      el2.setSelectionRange(pos, pos)
    })
  }

  return (
    <div className="space-y-1">
      <div className="flex flex-wrap items-center gap-1.5">
        <label className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.llm.prompt')}
        </label>
        {!readOnly && (
          <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoPrompt} />
        )}
      </div>
      <textarea
        ref={promptRef}
        value={prompt}
        readOnly={readOnly}
        onChange={(e) => setPrompt(e.target.value)}
        onBlur={commit}
        rows={8}
        placeholder={
          isEnrich
            ? 'Analyze this alert and classify it: $(alert.name)'
            : 'Investigate $(alert.target.host) and page on-call if needed'
        }
        className="w-full rounded-md border border-input bg-background px-2 py-1.5 font-mono text-[11px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
      />
      {isEnrich && (
        <p className="text-[10px] leading-snug text-muted-foreground">
          {t('soar.editor.canvas.llm.hint', { nodeId })}
        </p>
      )}
    </div>
  )
}

function extractPrompt(params: unknown): string {
  if (!params || typeof params !== 'object') return ''
  const p = (params as { prompt?: unknown }).prompt
  return typeof p === 'string' ? p : ''
}
