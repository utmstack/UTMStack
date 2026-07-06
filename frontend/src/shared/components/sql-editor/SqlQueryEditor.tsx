import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import Prism from 'prismjs'
import 'prismjs/components/prism-sql'
import { cn } from '@/shared/lib/utils'
import { useSqlAutocomplete } from './useSqlAutocomplete'
import type {
  SqlAutocompleteField,
  SqlAutocompletePattern,
} from './useSqlAutocomplete'
import { SqlAutocompleteDropdown } from './SqlAutocompleteDropdown'
import type { Suggestion } from './autocomplete-trie.service'

interface Props {
  value: string
  onChange: (v: string) => void
  onRun?: () => void
  fields: SqlAutocompleteField[]
  patterns: SqlAutocompletePattern[]
  placeholder?: string
  /** Minimum visible rows. Defaults to 4. */
  minRows?: number
  /** Maximum visible rows before scrolling. Defaults to 12. */
  maxRows?: number
}

const DEFAULT_MIN_ROWS = 4
const DEFAULT_MAX_ROWS = 12
const LINE_HEIGHT_PX = 18
const PADDING_Y = 8
const TOKEN_CHARS = /[A-Za-z0-9_.@-]/

interface TokenSpan {
  text: string
  start: number
  end: number
}

function tokenAtCaret(value: string, caret: number): TokenSpan | null {
  let start = caret
  while (start > 0 && TOKEN_CHARS.test(value[start - 1])) start--
  if (start === caret) return null
  return { text: value.slice(start, caret), start, end: caret }
}

function getCaretViewportCoords(textarea: HTMLTextAreaElement): { left: number; top: number } {
  const mirror = document.createElement('div')
  const cs = window.getComputedStyle(textarea)
  const copyProps = [
    'boxSizing', 'width', 'fontFamily', 'fontSize', 'fontWeight', 'fontStyle',
    'letterSpacing', 'lineHeight', 'paddingTop', 'paddingRight', 'paddingBottom',
    'paddingLeft', 'borderTopWidth', 'borderRightWidth', 'borderBottomWidth',
    'borderLeftWidth', 'tabSize', 'textIndent', 'textTransform', 'wordSpacing',
  ] as const
  for (const p of copyProps) {
    mirror.style.setProperty(p as string, cs.getPropertyValue(p as string))
  }
  mirror.style.position = 'absolute'
  mirror.style.visibility = 'hidden'
  mirror.style.whiteSpace = 'pre-wrap'
  mirror.style.wordWrap = 'break-word'
  mirror.style.overflow = 'hidden'
  mirror.style.top = '0'
  mirror.style.left = '-9999px'
  document.body.appendChild(mirror)

  const caret = textarea.selectionStart ?? textarea.value.length
  mirror.textContent = textarea.value.slice(0, caret)
  const marker = document.createElement('span')
  marker.textContent = textarea.value.slice(caret, caret + 1) || '.'
  mirror.appendChild(marker)

  const markerRect = marker.getBoundingClientRect()
  const mirrorRect = mirror.getBoundingClientRect()
  const taRect = textarea.getBoundingClientRect()
  document.body.removeChild(mirror)

  return {
    left: taRect.left + (markerRect.left - mirrorRect.left) - textarea.scrollLeft,
    top: taRect.top + (markerRect.top - mirrorRect.top) - textarea.scrollTop,
  }
}

export function SqlQueryEditor({
  value,
  onChange,
  onRun,
  fields,
  patterns,
  placeholder,
  minRows = DEFAULT_MIN_ROWS,
  maxRows = DEFAULT_MAX_ROWS,
}: Props) {
  const minHeight = minRows * LINE_HEIGHT_PX + PADDING_Y * 2
  const maxHeight = maxRows * LINE_HEIGHT_PX + PADDING_Y * 2
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const preRef = useRef<HTMLPreElement>(null)
  const [height, setHeight] = useState<number>(minHeight)
  const [open, setOpen] = useState(false)
  const [items, setItems] = useState<Suggestion[]>([])
  const [activeIndex, setActiveIndex] = useState(0)
  const [anchor, setAnchor] = useState<{ x: number; y: number }>({ x: 0, y: 0 })
  const tokenRef = useRef<TokenSpan | null>(null)

  const { suggest } = useSqlAutocomplete(fields, patterns)

  const highlighted = useMemo(() => {
    const grammar = Prism.languages.sql
    if (!grammar) return value
    return Prism.highlight(value + '\n', grammar, 'sql')
  }, [value])

  const recalcHeight = useCallback(() => {
    const ta = textareaRef.current
    if (!ta) return
    ta.style.height = 'auto'
    const next = Math.max(minHeight, Math.min(maxHeight, ta.scrollHeight))
    ta.style.height = ''
    setHeight(next)
  }, [minHeight, maxHeight])

  useLayoutEffect(() => {
    recalcHeight()
  }, [value, recalcHeight])

  const closeDropdown = useCallback(() => {
    setOpen(false)
    setItems([])
    tokenRef.current = null
  }, [])

  const refreshSuggestions = useCallback(() => {
    const ta = textareaRef.current
    if (!ta) return
    const caret = ta.selectionStart ?? 0
    const token = tokenAtCaret(value, caret)
    if (!token) {
      closeDropdown()
      return
    }
    const next = suggest(token.text, 30).filter(
      (s) => s.word.toLowerCase() !== token.text.toLowerCase(),
    )
    if (next.length === 0) {
      closeDropdown()
      return
    }
    tokenRef.current = token
    setItems(next)
    setActiveIndex(0)
    const coords = getCaretViewportCoords(ta)
    setAnchor({ x: coords.left, y: coords.top + LINE_HEIGHT_PX + 4 })
    setOpen(true)
  }, [closeDropdown, suggest, value])

  const accept = useCallback(
    (item: Suggestion) => {
      const ta = textareaRef.current
      const span = tokenRef.current
      if (!ta || !span) return
      const next = value.slice(0, span.start) + item.word + value.slice(span.end)
      onChange(next)
      closeDropdown()
      const newCaret = span.start + item.word.length
      requestAnimationFrame(() => {
        const t = textareaRef.current
        if (!t) return
        t.focus()
        t.setSelectionRange(newCaret, newCaret)
      })
    },
    [closeDropdown, onChange, value],
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault()
        closeDropdown()
        onRun?.()
        return
      }
      if (open) {
        if (e.key === 'ArrowDown') {
          e.preventDefault()
          setActiveIndex((i) => (i + 1) % items.length)
          return
        }
        if (e.key === 'ArrowUp') {
          e.preventDefault()
          setActiveIndex((i) => (i - 1 + items.length) % items.length)
          return
        }
        if (e.key === 'Enter' || e.key === 'Tab') {
          e.preventDefault()
          const picked = items[activeIndex]
          if (picked) accept(picked)
          return
        }
        if (e.key === 'Escape') {
          e.preventDefault()
          closeDropdown()
          return
        }
      } else if (e.key === 'Enter' && !e.shiftKey && onRun) {
        e.preventDefault()
        onRun()
      }
    },
    [accept, activeIndex, closeDropdown, items, onRun, open],
  )

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e.target.value)
    },
    [onChange],
  )


  useEffect(() => {
      if (document.activeElement === textareaRef.current) refreshSuggestions()
  }, [value, refreshSuggestions])

  useEffect(() => {
    if (!open) return
    const onMouseDown = (e: MouseEvent) => {
      const root = wrapperRef.current
      if (!root) return
      if (!root.contains(e.target as Node)) closeDropdown()
    }
    document.addEventListener('mousedown', onMouseDown)
    return () => document.removeEventListener('mousedown', onMouseDown)
  }, [open, closeDropdown])

  const syncScroll = useCallback(() => {
    const ta = textareaRef.current
    const pre = preRef.current
    if (!ta || !pre) return
    pre.scrollTop = ta.scrollTop
    pre.scrollLeft = ta.scrollLeft
  }, [])

  return (
    <div
      ref={wrapperRef}
      className="relative w-full"
      style={{ height, transition: 'height 150ms ease-out' }}
    >
      <style>{`
        .sql-editor-pre, .sql-editor-textarea {
          font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
          font-size: 12px;
          line-height: ${LINE_HEIGHT_PX}px;
          padding: ${PADDING_Y}px 10px;
          margin: 0;
          tab-size: 2;
          white-space: pre-wrap;
          word-break: break-word;
        }
        .sql-editor-pre .token.keyword { color: rgb(139 92 246); font-weight: 600; }
        .sql-editor-pre .token.string { color: rgb(16 185 129); }
        .sql-editor-pre .token.number { color: rgb(245 158 11); }
        .sql-editor-pre .token.operator,
        .sql-editor-pre .token.punctuation { color: rgb(100 116 139); }
        .sql-editor-pre .token.function { color: rgb(14 165 233); }
        .sql-editor-pre .token.comment { color: rgb(148 163 184); font-style: italic; }
        .sql-editor-pre .token.boolean { color: rgb(236 72 153); }
        .dark .sql-editor-pre .token.keyword { color: rgb(196 181 253); }
        .dark .sql-editor-pre .token.string { color: rgb(110 231 183); }
        .dark .sql-editor-pre .token.number { color: rgb(252 211 77); }
        .dark .sql-editor-pre .token.operator,
        .dark .sql-editor-pre .token.punctuation { color: rgb(148 163 184); }
        .dark .sql-editor-pre .token.function { color: rgb(125 211 252); }
        .dark .sql-editor-pre .token.boolean { color: rgb(244 114 182); }
      `}</style>
      <pre
        ref={preRef}
        aria-hidden
        className="sql-editor-pre pointer-events-none absolute inset-0 overflow-auto"
      >
        <code
          className="language-sql block"
          dangerouslySetInnerHTML={{ __html: highlighted }}
        />
      </pre>
      <textarea
        ref={textareaRef}
        value={value}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        onKeyUp={refreshSuggestions}
        onClick={refreshSuggestions}
        onScroll={syncScroll}
        spellCheck={false}
        placeholder={placeholder}
        rows={minRows}
        className={cn(
          'sql-editor-textarea relative block h-full w-full resize-none rounded-md',
          'border border-transparent bg-transparent text-transparent',
          'caret-foreground placeholder:text-muted-foreground/60',
          'outline-none focus:border-border',
        )}
      />
      {open && (
        <SqlAutocompleteDropdown
          items={items}
          activeIndex={activeIndex}
          position={anchor}
          onPick={accept}
          onHover={setActiveIndex}
        />
      )}
    </div>
  )
}
