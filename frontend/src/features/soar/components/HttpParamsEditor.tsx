import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Trash2 } from 'lucide-react'
import { Input } from '@/shared/components/ui/input'
import { cn } from '@/shared/lib/utils'
import { isValidHttpUrl, setHttpBodyError } from '../lib/http-node-validity'
import type { FlowNode } from '../types/soar.types'
import { InsertFieldMenu } from './InsertFieldMenu'
import { JsonCodeEditor } from './JsonCodeEditor'

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const
const BODY_METHODS = new Set<string>(['POST', 'PUT', 'PATCH'])
const SCHEMES = ['https', 'http'] as const

const SELECT =
  'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

interface HttpParams {
  method?: string
  url?: string
  body?: unknown
  headers?: Record<string, string>
  timeoutSec?: number
}

interface Props {
  nodeId: string
  nodes: Record<string, FlowNode>
  params: unknown
  readOnly?: boolean
  onChange: (params: HttpParams) => void
}

type PayloadTab = 'body' | 'headers'

// ponytail: URL split via one regex, body highlighted via Prism (already
// vendored). Validity for save-blocking flows through http-node-validity —
// no context wiring.
export function HttpParamsEditor({ nodeId, nodes, params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(params)
  const method = (p.method?.toUpperCase() || 'GET') as string
  const { scheme, rest } = splitUrl(p.url ?? '')
  const showBody = BODY_METHODS.has(method)
  const urlInvalid = Boolean(p.url) && !isValidHttpUrl(p.url ?? '')

  const urlRef = useRef<HTMLInputElement>(null)
  const bodyRef = useRef<HTMLTextAreaElement>(null)
  const [bodyText, setBodyText] = useState(() => bodyToText(p.body))
  const [bodyError, setBodyError] = useState<string | null>(null)
  const [tab, setTab] = useState<PayloadTab>('body')

  useEffect(() => {
    setBodyText(bodyToText(p.body))
    setBodyError(null)
    setHttpBodyError(nodeId, null)
  }, [p.body, nodeId])

  useEffect(() => {
    if (!showBody) {
      setHttpBodyError(nodeId, null)
      setBodyError(null)
    }
  }, [showBody, nodeId])

  useEffect(() => {
    if (showBody) setTab('body')
  }, [showBody])

  useEffect(() => () => setHttpBodyError(nodeId, null), [nodeId])

  const commitUrl = (nextScheme: string, nextRest: string) => {
    const trimmed = nextRest.trim()
    if (!trimmed) {
      onChange({ ...p, url: '' })
      return
    }
    if (/^https?:\/\//i.test(trimmed)) {
      onChange({ ...p, url: trimmed })
      return
    }
    onChange({ ...p, url: `${nextScheme}://${trimmed.replace(/^\/+/, '')}` })
  }

  const insertIntoUrl = (token: string) => {
    const el = urlRef.current
    const cur = rest
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const nextRest = cur.slice(0, start) + token + cur.slice(end)
    commitUrl(scheme, nextRest)
    requestAnimationFrame(() => {
      const el2 = urlRef.current
      if (!el2) return
      el2.focus()
      const pos = start + token.length
      el2.setSelectionRange(pos, pos)
    })
  }

  const insertIntoBody = (token: string) => {
    const el = bodyRef.current
    const cur = bodyText
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    setBodyText(next)
    requestAnimationFrame(() => {
      const el2 = bodyRef.current
      if (!el2) return
      el2.focus()
      const pos = start + token.length
      el2.setSelectionRange(pos, pos)
    })
  }

  const commitBody = (raw: string) => {
    const trimmed = raw.trim()
    if (!trimmed) {
      onChange({ ...p, body: undefined })
      setBodyError(null)
      setHttpBodyError(nodeId, null)
      return
    }
    try {
      const parsed = JSON.parse(trimmed)
      onChange({ ...p, body: parsed })
      setBodyError(null)
      setHttpBodyError(nodeId, null)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'invalid JSON'
      setBodyError(msg)
      setHttpBodyError(nodeId, msg)
    }
  }

  const switchTab = (next: PayloadTab) => {
    if (next === tab) return
    if (tab === 'body') commitBody(bodyText)
    setTab(next)
  }

  const headers = (showLabel: boolean) => (
    <HeaderRows
      headers={p.headers}
      readOnly={readOnly}
      nodes={nodes}
      currentNodeId={nodeId}
      showLabel={showLabel}
      onChange={(next) => onChange({ ...p, headers: next })}
    />
  )

  return (
    <div className="space-y-2">
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.http.url')}
        </label>
        {!readOnly && (
          <div className="mb-1 flex flex-wrap items-center gap-1.5">
            <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoUrl} />
          </div>
        )}
        <div className="flex items-center gap-1">
          <select
            value={scheme}
            disabled={readOnly}
            onChange={(e) => commitUrl(e.target.value, rest)}
            className={SELECT}
          >
            {SCHEMES.map((s) => (
              <option key={s} value={s}>
                {s}://
              </option>
            ))}
          </select>
          <Input
            ref={urlRef}
            value={rest}
            readOnly={readOnly}
            onChange={(e) => commitUrl(scheme, e.target.value)}
            placeholder="api.example.com/path"
            className={cn('h-8 flex-1 font-mono text-[11px]', urlInvalid && 'border-red-500')}
            aria-invalid={urlInvalid}
          />
        </div>
        {urlInvalid && (
          <p className="mt-1 text-[10px] text-red-500">
            {t('soar.editor.canvas.http.urlInvalid')}
          </p>
        )}
      </div>
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.http.method')}
        </label>
        <select
          value={method}
          disabled={readOnly}
          onChange={(e) => onChange({ ...p, method: e.target.value })}
          className={cn(SELECT, 'w-full')}
        >
          {METHODS.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
      </div>
      {showBody ? (
        <div>
          <div className="flex gap-1 border-b border-border">
            <TabButton
              active={tab === 'body'}
              onClick={() => switchTab('body')}
              label={t('soar.editor.canvas.http.body')}
            />
            <TabButton
              active={tab === 'headers'}
              onClick={() => switchTab('headers')}
              label={t('soar.editor.canvas.http.headers')}
            />
          </div>
          {tab === 'body' ? (
            <div>
              {!readOnly && (
                <div className="mb-1 flex flex-wrap items-center gap-1.5">
                  <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={insertIntoBody} />
                </div>
              )}
              <JsonCodeEditor
                value={bodyText}
                readOnly={readOnly}
                placeholder='{"foo":"bar"}'
                invalid={Boolean(bodyError)}
                onChange={setBodyText}
                onBlur={() => commitBody(bodyText)}
                textareaRef={bodyRef}
              />
              {bodyError && (
                <p className="mt-1 text-[10px] text-red-500">
                  {t('soar.editor.canvas.http.bodyInvalid')}: {bodyError}
                </p>
              )}
            </div>
          ) : (
            headers(false)
          )}
        </div>
      ) : (
        headers(true)
      )}
    </div>
  )
}

function TabButton({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'rounded-t border-b-2 px-2 py-1 text-[10px] font-medium uppercase tracking-wider transition-colors',
        active
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground',
      )}
    >
      {label}
    </button>
  )
}

function HeaderRows({
  headers,
  readOnly,
  nodes,
  currentNodeId,
  showLabel = true,
  onChange,
}: {
  headers?: Record<string, string>
  readOnly?: boolean
  nodes: Record<string, FlowNode>
  currentNodeId: string
  showLabel?: boolean
  onChange: (next: Record<string, string> | undefined) => void
}) {
  const { t } = useTranslation()
  const entries = Object.entries(headers ?? {})
  const valueRefs = useRef<Array<HTMLInputElement | null>>([])

  const commit = (next: Array<[string, string]>) => {
    const out: Record<string, string> = {}
    for (const [k, v] of next) {
      if (k.trim()) out[k.trim()] = v
    }
    onChange(Object.keys(out).length > 0 ? out : undefined)
  }

  const setAt = (i: number, patch: { key?: string; value?: string }) => {
    const next = entries.map(([k, v], j) =>
      j === i ? ([patch.key ?? k, patch.value ?? v] as [string, string]) : ([k, v] as [string, string]),
    )
    commit(next)
  }

  const insertIntoValue = (i: number, token: string) => {
    const el = valueRefs.current[i]
    const cur = entries[i]?.[1] ?? ''
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    setAt(i, { value: cur.slice(0, start) + token + cur.slice(end) })
    requestAnimationFrame(() => {
      const el2 = valueRefs.current[i]
      if (!el2) return
      el2.focus()
      const pos = start + token.length
      el2.setSelectionRange(pos, pos)
    })
  }

  return (
    <div>
      <div className="mb-1 flex items-center justify-between">
        {showLabel && (
          <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('soar.editor.canvas.http.headers')}
          </label>
        )}
        {!readOnly && (
          <button
            type="button"
            onClick={() => commit([...entries, ['', '']])}
            className="rounded px-1.5 py-0.5 text-[10px] text-primary hover:bg-muted"
          >
            {t('soar.editor.canvas.http.addHeader')}
          </button>
        )}
      </div>
      {entries.length === 0 && readOnly && (
        <p className="text-[10px] text-muted-foreground">—</p>
      )}
      <div className="space-y-1">
        {entries.map(([k, v], i) => (
          <div key={i} className="flex items-center gap-1">
            <Input
              value={k}
              readOnly={readOnly}
              onChange={(e) => setAt(i, { key: e.target.value })}
              placeholder="Authorization"
              className="h-7 w-2/5 font-mono text-[11px]"
            />
            <Input
              ref={(el) => {
                valueRefs.current[i] = el
              }}
              value={v}
              readOnly={readOnly}
              onChange={(e) => setAt(i, { value: e.target.value })}
              placeholder="Bearer $(variables.apiToken)"
              className="h-7 flex-1 font-mono text-[11px]"
            />
            {!readOnly && (
              <>
                <InsertFieldMenu
                  nodes={nodes}
                  currentNodeId={currentNodeId}
                  onInsert={(token) => insertIntoValue(i, token)}
                />
                <button
                  type="button"
                  onClick={() => commit(entries.filter((_, j) => j !== i))}
                  className="rounded p-1 text-muted-foreground hover:text-red-500"
                  title={t('soar.editor.canvas.deleteNode')}
                >
                  <Trash2 size={12} />
                </button>
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

function normalize(params: unknown): HttpParams {
  if (!params || typeof params !== 'object') return {}
  return params as HttpParams
}

function splitUrl(url: string): { scheme: string; rest: string } {
  const m = /^(https?):\/\/(.*)$/i.exec(url.trim())
  if (m) return { scheme: m[1].toLowerCase(), rest: m[2] }
  return { scheme: 'https', rest: url }
}

function bodyToText(body: unknown): string {
  if (body == null) return ''
  if (typeof body === 'string') return body
  try {
    return JSON.stringify(body, null, 2)
  } catch {
    return ''
  }
}
