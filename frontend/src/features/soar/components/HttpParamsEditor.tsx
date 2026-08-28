import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'
import { cn } from '@/shared/lib/utils'
import { isValidHttpUrl, setHttpBodyError } from '../lib/http-node-validity'
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
  params: unknown
  readOnly?: boolean
  onChange: (params: HttpParams) => void
}

// ponytail: URL split via one regex, body highlighted via Prism (already
// vendored). Validity for save-blocking flows through http-node-validity —
// no context wiring.
export function HttpParamsEditor({ nodeId, params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(params)
  const method = (p.method?.toUpperCase() || 'GET') as string
  const { scheme, rest } = splitUrl(p.url ?? '')
  const showBody = BODY_METHODS.has(method)
  const urlInvalid = Boolean(p.url) && !isValidHttpUrl(p.url ?? '')

  const [bodyText, setBodyText] = useState(() => bodyToText(p.body))
  const [bodyError, setBodyError] = useState<string | null>(null)

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

  return (
    <div className="space-y-2">
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.http.url')}
        </label>
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
      {showBody && (
        <div>
          <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('soar.editor.canvas.http.body')}
          </label>
          <JsonCodeEditor
            value={bodyText}
            readOnly={readOnly}
            placeholder='{"foo":"bar"}'
            invalid={Boolean(bodyError)}
            onChange={setBodyText}
            onBlur={() => commitBody(bodyText)}
          />
          {bodyError && (
            <p className="mt-1 text-[10px] text-red-500">
              {t('soar.editor.canvas.http.bodyInvalid')}: {bodyError}
            </p>
          )}
        </div>
      )}
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
