import { useMemo, useRef } from 'react'
import Prism from 'prismjs'
import 'prismjs/components/prism-json'
import { cn } from '@/shared/lib/utils'

interface Props {
  value: string
  readOnly?: boolean
  placeholder?: string
  invalid?: boolean
  onChange: (v: string) => void
  onBlur?: () => void
  /** Optional external ref — the parent can drive caret-insertions with it. */
  textareaRef?: React.RefObject<HTMLTextAreaElement | null>
}

// Prism-highlighted JSON textarea. Same overlay pattern SqlQueryEditor uses:
// aria-hidden <pre> renders highlighted tokens; transparent <textarea> owns
// input and caret. Simpler — no autocomplete.
export function JsonCodeEditor({ value, readOnly, placeholder, invalid, onChange, onBlur, textareaRef }: Props) {
  const preRef = useRef<HTMLPreElement>(null)
  const internalTaRef = useRef<HTMLTextAreaElement>(null)
  const taRef = textareaRef ?? internalTaRef

  const highlighted = useMemo(() => {
    const grammar = Prism.languages.json
    if (!grammar) return value
    return Prism.highlight(value + '\n', grammar, 'json')
  }, [value])

  const syncScroll = () => {
    const ta = taRef.current
    const pre = preRef.current
    if (!ta || !pre) return
    pre.scrollTop = ta.scrollTop
    pre.scrollLeft = ta.scrollLeft
  }

  return (
    <div
      className={cn(
        'relative h-40 overflow-hidden rounded-md border bg-background',
        invalid ? 'border-red-500' : 'border-input',
      )}
    >
      <style>{`
        .soar-json-pre, .soar-json-textarea {
          font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
          font-size: 11px;
          line-height: 16px;
          padding: 8px 10px;
          margin: 0;
          tab-size: 2;
          white-space: pre;
        }
        .soar-json-pre .token.property { color: rgb(14 165 233); }
        .soar-json-pre .token.string { color: rgb(16 185 129); }
        .soar-json-pre .token.number { color: rgb(245 158 11); }
        .soar-json-pre .token.boolean { color: rgb(236 72 153); }
        .soar-json-pre .token.null { color: rgb(148 163 184); font-style: italic; }
        .soar-json-pre .token.punctuation { color: rgb(100 116 139); }
        .dark .soar-json-pre .token.property { color: rgb(125 211 252); }
        .dark .soar-json-pre .token.string { color: rgb(110 231 183); }
        .dark .soar-json-pre .token.number { color: rgb(252 211 77); }
        .dark .soar-json-pre .token.boolean { color: rgb(244 114 182); }
      `}</style>
      <pre
        ref={preRef}
        aria-hidden
        className="soar-json-pre pointer-events-none absolute inset-0 overflow-auto"
      >
        <code
          className="language-json block"
          dangerouslySetInnerHTML={{ __html: highlighted }}
        />
      </pre>
      <textarea
        ref={taRef}
        value={value}
        readOnly={readOnly}
        spellCheck={false}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        onBlur={onBlur}
        onScroll={syncScroll}
        className={cn(
          'soar-json-textarea relative block h-full w-full resize-none border-0 bg-transparent',
          'text-transparent caret-foreground placeholder:text-muted-foreground/60 outline-none',
        )}
      />
    </div>
  )
}
