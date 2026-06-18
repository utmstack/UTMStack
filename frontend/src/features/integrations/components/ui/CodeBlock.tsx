import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Copy, Check } from 'lucide-react'

// ── Secret masking ────────────────────────────────────────────────────────────
// Segments parsed from a code string that may contain <secret>…</secret> spans.
type Segment = { text: string; secret: boolean }

function parseSegments(code: string): Segment[] {
  const segments: Segment[] = []
  const re = /<secret>([\s\S]*?)<\/secret>/g
  let last = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(code)) !== null) {
    if (m.index > last) segments.push({ text: code.slice(last, m.index), secret: false })
    segments.push({ text: m[1], secret: true })
    last = m.index + m[0].length
  }
  if (last < code.length) segments.push({ text: code.slice(last), secret: false })
  return segments
}

// ── Bash syntax highlighting ──────────────────────────────────────────────────
type TokType = 'comment' | 'dstring' | 'sstring' | 'var' | 'url' | 'path' | 'flag' | 'op' | 'cmd' | 'num' | 'plain'
interface Token { type: TokType; text: string }

// Priority-ordered alternation: first match wins.
const BASH_RE = /(#[^\n]*)|("(?:[^"\\]|\\.)*")|('(?:[^'\\]|\\.)*')|(\$\{?[\w]+\}?)|(https?:\/\/[^\s"']+)|(\/[^\s"'\\,]+)|(--?[\w][\w-]*)|(&&|\|\||\\\n)|\b(sudo|bash|sh|apt-get|apt|wget|curl|chmod|mkdir|cp|mv|rm|echo|cat|install|update|yes|no)\b|(\b\d+\b)/g

const TOKEN_CLASS: Record<TokType, string> = {
  comment: 'text-slate-400 dark:text-slate-500',
  dstring: 'text-emerald-700 dark:text-emerald-400',
  sstring: 'text-emerald-700 dark:text-emerald-400',
  var:     'text-violet-600 dark:text-violet-400',
  url:     'text-sky-600 dark:text-sky-300',
  path:    'text-sky-600/90 dark:text-sky-200/80',
  flag:    'text-cyan-700 dark:text-cyan-400',
  op:      'text-slate-500 dark:text-slate-400',
  cmd:     'text-amber-700 dark:text-amber-300',
  num:     'text-emerald-700 dark:text-emerald-300',
  plain:   'text-slate-700 dark:text-slate-200',
}

function tokenizeBash(text: string): Token[] {
  const tokens: Token[] = []
  let last = 0
  let m: RegExpExecArray | null
  BASH_RE.lastIndex = 0
  while ((m = BASH_RE.exec(text)) !== null) {
    if (m.index > last) tokens.push({ type: 'plain', text: text.slice(last, m.index) })
    const [full, comment, dstr, sstr, variable, url, path, flag, op, cmd, num] = m
    if (comment  !== undefined) tokens.push({ type: 'comment', text: comment })
    else if (dstr  !== undefined) tokens.push({ type: 'dstring', text: dstr })
    else if (sstr  !== undefined) tokens.push({ type: 'sstring', text: sstr })
    else if (variable !== undefined) tokens.push({ type: 'var',  text: variable })
    else if (url   !== undefined) tokens.push({ type: 'url',     text: url })
    else if (path  !== undefined) tokens.push({ type: 'path',    text: path })
    else if (flag  !== undefined) tokens.push({ type: 'flag',    text: flag })
    else if (op    !== undefined) tokens.push({ type: 'op',      text: op })
    else if (cmd   !== undefined) tokens.push({ type: 'cmd',     text: cmd })
    else if (num   !== undefined) tokens.push({ type: 'num',     text: num })
    else tokens.push({ type: 'plain', text: full })
    last = m.index + full.length
  }
  if (last < text.length) tokens.push({ type: 'plain', text: text.slice(last) })
  return tokens
}

// ── PowerShell syntax highlighting ────────────────────────────────────────────
// Priority-ordered: comment, double/single strings, $variable, url, Verb-Noun
// cmdlet, -Parameter, operators, number.
const PS_RE = /(#[^\n]*)|("(?:[^"\\]|\\.)*")|('(?:[^'\\]|\\.)*')|(\$\{?[\w]+\}?)|(https?:\/\/[^\s"']+)|(\b[A-Z][a-zA-Z]+-[A-Za-z]+\b)|(-[A-Za-z][\w]*)|(&|\||;|::)|(\b\d+\b)/g

function tokenizePowershell(text: string): Token[] {
  const tokens: Token[] = []
  let last = 0
  let m: RegExpExecArray | null
  PS_RE.lastIndex = 0
  while ((m = PS_RE.exec(text)) !== null) {
    if (m.index > last) tokens.push({ type: 'plain', text: text.slice(last, m.index) })
    const [full, comment, dstr, sstr, variable, url, cmdlet, param, op, num] = m
    if (comment  !== undefined) tokens.push({ type: 'comment', text: comment })
    else if (dstr    !== undefined) tokens.push({ type: 'dstring', text: dstr })
    else if (sstr    !== undefined) tokens.push({ type: 'sstring', text: sstr })
    else if (variable !== undefined) tokens.push({ type: 'var',  text: variable })
    else if (url     !== undefined) tokens.push({ type: 'url',     text: url })
    else if (cmdlet  !== undefined) tokens.push({ type: 'cmd',     text: cmdlet })
    else if (param   !== undefined) tokens.push({ type: 'flag',    text: param })
    else if (op      !== undefined) tokens.push({ type: 'op',      text: op })
    else if (num     !== undefined) tokens.push({ type: 'num',     text: num })
    else tokens.push({ type: 'plain', text: full })
    last = m.index + full.length
  }
  if (last < text.length) tokens.push({ type: 'plain', text: text.slice(last) })
  return tokens
}

// ── Config (key: value) highlighting ─────────────────────────────────────────
// Used for endpoint / connection-details blocks.
const CONFIG_LINE_RE = /^(\s*[\w_-]+)(\s*:\s*)(.+)$/gm

function tokenizeConfig(text: string): Array<{ cls: string; text: string }> {
  const parts: Array<{ cls: string; text: string }> = []
  let last = 0
  let m: RegExpExecArray | null
  CONFIG_LINE_RE.lastIndex = 0
  while ((m = CONFIG_LINE_RE.exec(text)) !== null) {
    if (m.index > last) parts.push({ cls: 'text-slate-200', text: text.slice(last, m.index) })
    parts.push({ cls: 'text-cyan-400',  text: m[1] })
    parts.push({ cls: 'text-slate-500', text: m[2] })
    parts.push({ cls: 'text-emerald-300', text: m[3] })
    last = m.index + m[0].length
  }
  if (last < text.length) parts.push({ cls: 'text-slate-200', text: text.slice(last) })
  return parts
}

// ── YAML highlighting ─────────────────────────────────────────────────────────
function tokenizeYaml(text: string): Array<{ cls: string; text: string }> {
  const parts: Array<{ cls: string; text: string }> = []
  const lines = text.split('\n')
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    if (i > 0) parts.push({ cls: '', text: '\n' })
    const trim = line.trim()
    if (trim.startsWith('#')) {
      parts.push({ cls: 'text-slate-400 dark:text-slate-500', text: line })
      continue
    }
    // List item: leading spaces + "- " + optional key: value
    const listM = line.match(/^(\s*)(- )(.*)$/)
    if (listM) {
      const [, indent, dash, rest] = listM
      if (indent) parts.push({ cls: 'text-slate-700 dark:text-slate-200', text: indent })
      parts.push({ cls: 'text-slate-500 dark:text-slate-400', text: dash })
      const kvM = rest.match(/^([\w-]+)(\s*:\s*)(.*)$/)
      if (kvM) {
        parts.push({ cls: 'text-cyan-700 dark:text-cyan-400', text: kvM[1] })
        parts.push({ cls: 'text-slate-500 dark:text-slate-400', text: kvM[2] })
        if (kvM[3]) parts.push({ cls: 'text-emerald-700 dark:text-emerald-400', text: kvM[3] })
      } else {
        parts.push({ cls: 'text-slate-700 dark:text-slate-200', text: rest })
      }
      continue
    }
    // key: value
    const kvM = line.match(/^(\s*)([\w-]+)(\s*:\s*)(.*)$/)
    if (kvM) {
      const [, indent, key, sep, val] = kvM
      if (indent) parts.push({ cls: 'text-slate-700 dark:text-slate-200', text: indent })
      parts.push({ cls: 'text-cyan-700 dark:text-cyan-400', text: key })
      parts.push({ cls: 'text-slate-500 dark:text-slate-400', text: sep })
      if (val) parts.push({ cls: 'text-emerald-700 dark:text-emerald-400', text: val })
      continue
    }
    parts.push({ cls: 'text-slate-700 dark:text-slate-200', text: line })
  }
  return parts
}

// ── Component ─────────────────────────────────────────────────────────────────
export function CodeBlock({ code, lang = 'bash' }: { code: string; lang?: 'bash' | 'config' | 'yaml' | 'powershell' }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const segments = parseSegments(code)
  const copyText = segments.map((s) => s.text).join('')

  const handleCopy = () => {
    navigator.clipboard.writeText(copyText)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 bg-slate-50 font-mono text-[11px] leading-relaxed dark:border-[hsl(222_18%_18%)] dark:bg-[hsl(222_33%_7%)]">
      {/* Terminal chrome */}
      <div className="flex items-center justify-between border-b border-slate-200 px-3 py-1.5 dark:border-[hsl(222_18%_12%)]">
        <div className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-full bg-red-500/60" />
          <span className="h-2 w-2 rounded-full bg-yellow-400/60" />
          <span className="h-2 w-2 rounded-full bg-emerald-500/60" />
        </div>
        <span className="text-[9px] font-medium uppercase tracking-[0.15em] text-slate-400 dark:text-slate-600">
          {lang}
        </span>
        <button
          title={copied ? t('integrations.codeBlock.copied') : t('integrations.codeBlock.copy')}
          onClick={handleCopy}
          className="flex items-center gap-1 rounded px-1.5 py-0.5 text-[9px] text-slate-400 transition-colors hover:bg-slate-200 hover:text-slate-700 dark:text-slate-500 dark:hover:bg-[hsl(222_18%_12%)] dark:hover:text-slate-300"
        >
          {copied ? <Check size={9} className="text-emerald-400" /> : <Copy size={9} />}
          <span className={copied ? 'text-emerald-400' : ''}>
            {copied ? t('integrations.codeBlock.copied') : t('integrations.codeBlock.copy')}
          </span>
        </button>
      </div>

      {/* Code */}
      <pre className="overflow-x-auto p-4">
        {segments.map((seg, i) =>
          seg.secret ? (
            <span
              key={i}
              className="select-none rounded bg-slate-200 px-1 text-slate-400 dark:bg-[hsl(222_25%_14%)] dark:text-slate-600"
              aria-label="hidden value"
            >
              {'•'.repeat(7)}
            </span>
          ) : lang === 'bash' || lang === 'powershell' ? (
            (lang === 'powershell' ? tokenizePowershell : tokenizeBash)(seg.text).map((tok, j) => (
              <span key={`${i}-${j}`} className={TOKEN_CLASS[tok.type]}>
                {tok.text}
              </span>
            ))
          ) : lang === 'yaml' ? (
            tokenizeYaml(seg.text).map((part, j) => (
              <span key={`${i}-${j}`} className={part.cls}>
                {part.text}
              </span>
            ))
          ) : (
            tokenizeConfig(seg.text).map((part, j) => (
              <span key={`${i}-${j}`} className={part.cls}>
                {part.text}
              </span>
            ))
          ),
        )}
      </pre>
    </div>
  )
}
