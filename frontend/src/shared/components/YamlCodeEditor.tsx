import { Fragment, useRef, type ReactNode } from 'react'
import { cn } from '@/shared/lib/utils'

interface Props {
  value: string
  readOnly?: boolean
  placeholder?: string
  onChange?: (value: string) => void
  className?: string
  /** Optional outside access to the underlying textarea (e.g. caret-aware insertion). */
  textareaRef?: React.RefObject<HTMLTextAreaElement | null>
}

/**
 * Lightweight YAML editor with syntax highlighting and no external deps: a
 * colored <pre> highlight layer sits behind a transparent <textarea>, and the
 * two scroll in sync. Both share identical font/padding/wrapping so the caret
 * and tokens line up exactly.
 */
export function YamlCodeEditor({ value, readOnly, placeholder, onChange, className, textareaRef }: Props) {
  const taRef = useRef<HTMLTextAreaElement | null>(null)
  const preRef = useRef<HTMLPreElement>(null)

  // Mirror the internal textarea node into the caller's ref, when provided.
  const setTextarea = (node: HTMLTextAreaElement | null) => {
    taRef.current = node
    if (textareaRef) textareaRef.current = node
  }

  const syncScroll = () => {
    if (taRef.current && preRef.current) {
      preRef.current.scrollTop = taRef.current.scrollTop
      preRef.current.scrollLeft = taRef.current.scrollLeft
    }
  }

  // Both layers MUST share these classes so glyphs align perfectly.
  const shared = 'm-0 whitespace-pre-wrap break-words p-3 font-mono text-[11px] leading-relaxed'

  return (
    <div
      className={cn(
        'relative min-h-0 flex-1 overflow-hidden rounded-md border border-input bg-background/40 focus-within:border-ring focus-within:ring-1 focus-within:ring-ring',
        className,
      )}
    >
      <pre ref={preRef} aria-hidden className={cn(shared, 'pointer-events-none absolute inset-0 overflow-hidden')}>
        {highlightYaml(value)}
        {'\n'}
      </pre>
      <textarea
        ref={setTextarea}
        value={value}
        readOnly={readOnly}
        spellCheck={false}
        placeholder={placeholder}
        onChange={(e) => onChange?.(e.target.value)}
        onScroll={syncScroll}
        className={cn(
          shared,
          'absolute inset-0 h-full w-full resize-none overflow-auto bg-transparent text-transparent caret-foreground outline-none placeholder:text-muted-foreground read-only:opacity-80',
        )}
      />
    </div>
  )
}

/* ── Tokenizer ────────────────────────────────────────────────────────────
   Line-based YAML highlighting — enough for the pipeline: filter format. */

const C = {
  comment: 'text-emerald-600 dark:text-emerald-400/80',
  key: 'text-sky-700 dark:text-sky-300',
  string: 'text-amber-700 dark:text-amber-300',
  literal: 'text-violet-700 dark:text-violet-300',
  punct: 'text-muted-foreground',
}

interface Tok {
  text: string
  cls?: string
}

function highlightYaml(code: string): ReactNode {
  const lines = code.split('\n')
  return lines.map((line, i) => (
    <Fragment key={i}>
      {lineTokens(line).map((tk, j) => (tk.cls ? <span key={j} className={tk.cls}>{tk.text}</span> : <Fragment key={j}>{tk.text}</Fragment>))}
      {i < lines.length - 1 ? '\n' : ''}
    </Fragment>
  ))
}

function lineTokens(line: string): Tok[] {
  const out: Tok[] = []
  const m = /^(\s*)(.*)$/.exec(line)!
  const indent = m[1]
  let rest = m[2]
  if (indent) out.push({ text: indent })
  if (rest === '') return out

  // Full-line comment.
  if (rest.startsWith('#')) {
    out.push({ text: rest, cls: C.comment })
    return out
  }

  // List dash(es): "- " or a bare "-".
  if (rest === '-') {
    out.push({ text: '-', cls: C.punct })
    return out
  }
  const dash = /^-\s+/.exec(rest)
  if (dash) {
    out.push({ text: dash[0], cls: C.punct })
    rest = rest.slice(dash[0].length)
  }

  // key: value
  const kv = /^("[^"]*"|'[^']*'|[\w.\-/@]+)(\s*:)(\s?)(.*)$/.exec(rest)
  if (kv) {
    out.push({ text: kv[1], cls: C.key })
    out.push({ text: kv[2], cls: C.punct })
    if (kv[3]) out.push({ text: kv[3] })
    if (kv[4] !== '') pushValue(kv[4], out)
    return out
  }

  pushValue(rest, out)
  return out
}

function pushValue(v: string, out: Tok[]) {
  const trimmed = v.trim()

  // Block scalar indicator (|, >, |-, >+, etc.).
  if (/^[|>][+-]?\d*$/.test(trimmed)) {
    out.push({ text: v, cls: C.punct })
    return
  }
  // Quoted string.
  if (/^".*"$/.test(trimmed) || /^'.*'$/.test(trimmed)) {
    out.push({ text: v, cls: C.string })
    return
  }
  // Number / boolean / null.
  if (/^(true|false|null|~|-?\d+(\.\d+)?)$/i.test(trimmed)) {
    out.push({ text: v, cls: C.literal })
    return
  }
  // Trailing inline comment.
  const ci = v.indexOf(' #')
  if (ci >= 0) {
    out.push({ text: v.slice(0, ci) })
    out.push({ text: v.slice(ci), cls: C.comment })
    return
  }
  out.push({ text: v })
}
