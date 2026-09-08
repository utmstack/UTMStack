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
const EMPTY_ROWS: Array<[string, string]> = []
const DEFAULT_HEADER_ROWS: Array<[string, string]> = [
  ['Postman-Token', '<calculated when request is sent>'],
  ['Host', '<calculated when request is sent>'],
  ['User-Agent', 'PostmanRuntime/7.43.0'],
  ['Accept', '*/*'],
  ['Accept-Encoding', 'gzip, deflate, br'],
  ['Connection', 'keep-alive'],
]
type Row = [string, string]
type RowCache = { rows: Row[]; selected: boolean[] }
const HTTP_ROWS_CACHE = new Map<string, RowCache>()

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

type PayloadTab = 'headers' | 'params' | 'body'

// ponytail: URL split via one regex, body highlighted via Prism (already
// vendored). Validity for save-blocking flows through http-node-validity —
// no context wiring.
export function HttpParamsEditor({ nodeId, nodes, params: httpParams, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(httpParams)
  const method = (p.method?.toUpperCase() || 'GET') as string
  const { scheme, rest } = splitUrl(p.url ?? '')
  const queryParams = parseQueryParams(p.url ?? '')
  const showBody = BODY_METHODS.has(method)
  const urlInvalid = Boolean(p.url) && !isValidHttpUrl(p.url ?? '')

  const urlRef = useRef<HTMLInputElement>(null)
  const bodyRef = useRef<HTMLTextAreaElement>(null)
  const queryParamsRef = useRef<Record<string, string>>(queryParams)
  const activeQueryParams = getActiveQueryParams(nodeId, queryParamsRef.current, queryParams)
  const previewBaseUrl = p.url || (rest.trim() ? `${scheme}://${rest}` : '')
  const fullUrlPreview = appendQueryParams(previewBaseUrl, activeQueryParams)
  const [bodyText, setBodyText] = useState(() => bodyToText(p.body))
  const [bodyError, setBodyError] = useState<string | null>(null)
  const [tab, setTab] = useState<PayloadTab>('headers')

  useEffect(() => {
    setBodyText(bodyToText(p.body))
    setBodyError(null)
    setHttpBodyError(nodeId, null)
  }, [p.body, nodeId])

  useEffect(() => {
    queryParamsRef.current = queryParams
  }, [nodeId])

  useEffect(() => {
    if (!showBody) {
      setHttpBodyError(nodeId, null)
      setBodyError(null)
    }
  }, [showBody, nodeId])

  useEffect(() => {
    if (!showBody && tab === 'body') setTab('headers')
  }, [showBody, tab])

  useEffect(() => () => setHttpBodyError(nodeId, null), [nodeId])

  const emit = (changes: Partial<HttpParams>) => {
    const { params: _queryParams, ...current } = p as HttpParams & {
      params?: Record<string, string>
    }
    onChange({ ...current, ...changes })
  }

  useEffect(() => {
    if (!readOnly && !p.headers) {
      emit({ headers: Object.fromEntries(DEFAULT_HEADER_ROWS) })
    }
  }, [p.headers, readOnly])

  const commitUrl = (nextScheme: string, nextRest: string) => {
    const trimmed = nextRest.trim()
    const activeParams = getActiveQueryParams(nodeId, queryParamsRef.current, queryParams)
    if (!trimmed) {
      emit({ url: '' })
      return
    }
    if (/^https?:\/\//i.test(trimmed)) {
      emit({ url: appendQueryParams(trimmed, activeParams) })
      return
    }
    emit({
      url: appendQueryParams(`${nextScheme}://${trimmed.replace(/^\/+/, '')}`, activeParams),
    })
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
      emit({ body: undefined })
      setBodyError(null)
      setHttpBodyError(nodeId, null)
      return
    }
    try {
      const parsed = JSON.parse(trimmed)
      emit({ body: parsed })
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

  const renderParams = (showLabel: boolean) => (
    <KeyValueRows
      title={t('soar.editor.canvas.http.params')}
      values={queryParams}
      defaultRows={EMPTY_ROWS}
      addLabel={t('soar.editor.canvas.http.addParam')}
      readOnly={readOnly}
      nodes={nodes}
      currentNodeId={nodeId}
      cacheKey={`${nodeId}:params`}
      showLabel={showLabel}
      onChange={(next) => {
        queryParamsRef.current = next ?? {}
        emit({ url: appendQueryParams(p.url, next) })
      }}
    />
  )

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
        {fullUrlPreview && (
          <p
            className="mt-1 truncate font-mono text-[10px] text-muted-foreground"
            title={fullUrlPreview}
          >
            {fullUrlPreview}
          </p>
        )}
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
      <div>
        <div className="flex gap-1 border-b border-border">
          <TabButton
            active={tab === 'headers'}
            onClick={() => switchTab('headers')}
            label={t('soar.editor.canvas.http.headers')}
          />
          <TabButton
            active={tab === 'params'}
            onClick={() => switchTab('params')}
            label={t('soar.editor.canvas.http.params')}
          />
          {showBody && (
            <TabButton
              active={tab === 'body'}
              onClick={() => switchTab('body')}
              label={t('soar.editor.canvas.http.body')}
            />
          )}
        </div>
        <div hidden={tab !== 'headers'}>{headers(false)}</div>
        <div
          hidden={tab !== 'params'}
          className={cn(Object.keys(queryParams).length === 0 && 'pb-2')}
        >
          {renderParams(false)}
        </div>
        {showBody && (
          <div hidden={tab !== 'body'}>
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
        )}
      </div>
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

function KeyValueRows({
  title,
  values,
  defaultRows = [],
  addLabel,
  cacheKey,
  readOnly,
  nodes,
  currentNodeId,
  showLabel = true,
  onChange,
}: {
  title: string
  values?: Record<string, string>
  defaultRows?: Array<[string, string]>
  addLabel: string
  cacheKey: string
  readOnly?: boolean
  nodes: Record<string, FlowNode>
  currentNodeId: string
  showLabel?: boolean
  onChange: (next: Record<string, string> | undefined) => void
}) {
  const { t } = useTranslation()

  const incomingRows = values && Object.keys(values).length > 0 ? Object.entries(values) : []
  const cachedRows = HTTP_ROWS_CACHE.get(cacheKey)
  const initialRows = cachedRows?.rows ?? (incomingRows.length > 0 ? incomingRows : defaultRows)
  const initialSelected = cachedRows?.selected ?? initialRows.map(() => true)
  const [rows, setRows] = useState<Row[]>(() => initialRows)
  const [selected, setSelected] = useState<boolean[]>(() => initialSelected)
  const valueRefs = useRef<Array<HTMLInputElement | null>>([])
  const cacheRef = useRef<RowCache>({ rows: [...initialRows], selected: [...initialSelected] })

  useEffect(() => {
    const incomingRows = values ? Object.entries(values) : []
    const incomingByKey = new Map(incomingRows)
    const cachedKeys = new Set(cacheRef.current.rows.map(([key]) => key))
    const nextRows = cacheRef.current.rows.map(([key, value]) =>
      incomingByKey.has(key) ? [key, incomingByKey.get(key) ?? value] : [key, value],
    ) as Array<[string, string]>

    incomingRows.forEach(([key, value]) => {
      if (!cachedKeys.has(key)) nextRows.push([key, value])
    })

    cacheRef.current.rows = nextRows
    setRows(nextRows)
    const selectedByKey = new Map(
      cacheRef.current.rows.map(([key], index) => [key, cacheRef.current.selected[index] ?? true]),
    )
    const nextSelected = nextRows.map(([key]) => selectedByKey.get(key) ?? true)
    cacheRef.current.selected = nextSelected
    HTTP_ROWS_CACHE.set(cacheKey, cacheRef.current)
    setSelected(nextSelected)
  }, [values, defaultRows, cacheKey])

  const commit = (next: Array<[string, string]>, nextSelected: boolean[] = selected) => {
    const out: Record<string, string> = {}
    next.forEach(([k, v], index) => {
      if (!nextSelected[index]) return
      if (k.trim()) out[k.trim()] = v
    })
    onChange(Object.keys(out).length > 0 ? out : undefined)
  }

  const setAt = (i: number, patch: { key?: string; value?: string }) => {
    const next = rows.map(([k, v], j) =>
      j === i ? ([patch.key ?? k, patch.value ?? v] as [string, string]) : ([k, v] as [string, string]),
    )
    cacheRef.current.rows = next
    HTTP_ROWS_CACHE.set(cacheKey, cacheRef.current)
    setRows(next)
    commit(next)
  }

  const visibleRows = rows.map((row, index) => ({ row, selected: selected[index] ?? true }))

  const insertIntoValue = (i: number, token: string) => {
    const el = valueRefs.current[i]
    const cur = rows[i]?.[1] ?? ''
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

  const toggleRow = (index: number) => {
    const nextSelected = selected.map((isSelected, i) => (i === index ? !isSelected : isSelected))
    cacheRef.current.selected = nextSelected
    HTTP_ROWS_CACHE.set(cacheKey, cacheRef.current)
    setSelected(nextSelected)
    commit(rows, nextSelected)
  }

  const removeRow = (index: number) => {
    const nextRows = rows.filter((_, i) => i !== index)
    const nextSelected = selected.filter((_, i) => i !== index)
    cacheRef.current.rows = nextRows
    cacheRef.current.selected = nextSelected
    HTTP_ROWS_CACHE.set(cacheKey, cacheRef.current)
    setRows(nextRows)
    setSelected(nextSelected.length > 0 ? nextSelected : [])
    commit(nextRows, nextSelected)
  }

  const addRow = () => {
    const nextRows: Array<[string, string]> = [...rows, ['', ''] as [string, string]]
    const nextSelected = [...selected, true]
    cacheRef.current.rows = nextRows
    cacheRef.current.selected = nextSelected
    HTTP_ROWS_CACHE.set(cacheKey, cacheRef.current)
    setRows(nextRows)
    setSelected(nextSelected)
  }

  return (
    <div>
      <div className="mb-1 flex items-center justify-between">
        {showLabel && (
          <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {title}
          </label>
        )}
        {!readOnly && (
          <button
            type="button"
            onClick={addRow}
            className="ml-auto mt-1 rounded px-1.5 py-0.5 text-[10px] text-primary hover:bg-muted"
          >
            {addLabel}
          </button>
        )}
      </div>
      {rows.length === 0 && readOnly && (
        <p className="text-[10px] text-muted-foreground">—</p>
      )}
      <div className="space-y-1">
        <div className="grid grid-cols-[28px_minmax(0,1fr)_minmax(0,1.4fr)_24px] items-center gap-2 border-b border-border px-1 pb-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          <span className="text-center"> </span>
          <span>{t('soar.editor.canvas.http.key') || 'Key'}</span>
          <span>{t('soar.editor.canvas.http.value') || 'Value'}</span>
        </div>
        {visibleRows.map(({ row, selected: isSelected }, i) => {
          const [k, v] = row
          return (
            <div
              key={i}
              className={cn(
                'grid grid-cols-[28px_minmax(0,1fr)_minmax(0,1.4fr)_24px] items-center gap-2',
                !isSelected && 'opacity-50'
              )}
            >
              <div className="flex justify-center">
                {!readOnly && (
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => toggleRow(i)}
                    className="h-4 w-4 rounded border-input accent-primary"
                    aria-label={`Toggle ${k || 'header'} ${title.toLowerCase()}`}
                  />
                )}
              </div>
              <Input
                value={k}
                readOnly={readOnly}
                onChange={(e) => setAt(i, { key: e.target.value })}
                placeholder={t('soar.editor.canvas.http.key') || 'Key'}
                className="h-7 font-mono text-[11px]"
              />
              <div className="flex items-center gap-1">
                <Input
                  ref={(el) => {
                    valueRefs.current[i] = el
                  }}
                  value={v}
                  readOnly={readOnly}
                  onChange={(e) => setAt(i, { value: e.target.value })}
                  placeholder={t('soar.editor.canvas.http.value') || 'Value'}
                  className="h-7 flex-1 font-mono text-[11px]"
                />
                {!readOnly && (
                  <InsertFieldMenu
                    nodes={nodes}
                    currentNodeId={currentNodeId}
                    onInsert={(token) => insertIntoValue(i, token)}
                  />
                )}
              </div>
              {!readOnly && (
                <div className="flex justify-center">
                  <button
                    type="button"
                    onClick={() => removeRow(i)}
                    className="rounded p-1 text-muted-foreground hover:text-red-500"
                    title={t('soar.editor.canvas.deleteNode')}
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
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

  return (
    <KeyValueRows
      title={t('soar.editor.canvas.http.headers')}
      values={headers}
      defaultRows={DEFAULT_HEADER_ROWS}
      addLabel={t('soar.editor.canvas.http.addHeader')}
      readOnly={readOnly}
      nodes={nodes}
      currentNodeId={currentNodeId}
      cacheKey={`${currentNodeId}:headers`}
      showLabel={showLabel}
      onChange={onChange}
    />
  )
}

function normalize(params: unknown): HttpParams {
  if (!params || typeof params !== 'object') return {}
  return params as HttpParams
}

function splitUrl(url: string): { scheme: string; rest: string } {
  const m = /^(https?):\/\/(.*)$/i.exec(url.trim())
  const value = m ? m[2] : url
  const queryIndex = value.indexOf('?')
  const hashIndex = value.indexOf('#')
  const suffixIndex = [queryIndex, hashIndex].filter((index) => index >= 0).sort((a, b) => a - b)[0]
  const rest = suffixIndex === undefined ? value : value.slice(0, suffixIndex)
  return { scheme: m ? m[1].toLowerCase() : 'https', rest }
}

function parseQueryParams(url: string): Record<string, string> {
  const queryStart = url.indexOf('?')
  if (queryStart < 0) return {}

  const hashStart = url.indexOf('#', queryStart)
  const query = url.slice(queryStart + 1, hashStart >= 0 ? hashStart : undefined)
  const values: Record<string, string> = {}
  new URLSearchParams(query).forEach((value, key) => {
    values[key] = value
  })
  return values
}

function rowsToRecord(rows?: Row[], selected?: boolean[]): Record<string, string> {
  const values: Record<string, string> = {}
  rows?.forEach(([key, value], index) => {
    if (selected?.[index] === false || !key.trim()) return
    values[key.trim()] = value
  })
  return values
}

function getActiveQueryParams(
  nodeId: string,
  current: Record<string, string>,
  fromUrl: Record<string, string>,
): Record<string, string> {
  if (Object.keys(current).length > 0) return current
  const cached = HTTP_ROWS_CACHE.get(`${nodeId}:params`)
  const cachedValues = rowsToRecord(cached?.rows, cached?.selected)
  if (Object.keys(cachedValues).length > 0) return cachedValues
  return fromUrl
}

function appendQueryParams(url: string | undefined, params?: Record<string, string>): string {
  if (!url) return ''

  const hashIndex = url.indexOf('#')
  const hash = hashIndex >= 0 ? url.slice(hashIndex) : ''
  const withoutHash = hashIndex >= 0 ? url.slice(0, hashIndex) : url
  const base = withoutHash.split('?')[0]
  const query = new URLSearchParams()

  Object.entries(params ?? {}).forEach(([key, value]) => {
    if (key.trim()) query.set(key.trim(), value)
  })

  const serialized = query.toString()
  return `${base}${serialized ? `?${serialized}` : ''}${hash}`
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
