import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Braces } from 'lucide-react'
import type { RegexPattern } from '@/features/regex-patterns/services/regex-patterns-http.service'
import { PatternMenu } from './PatternMenu'

/**
 * Floating "{ } Insert pattern" control for the YAML content editor. Opens the
 * shared pattern picker and drops `{{.name}}` at the textarea's caret — so named
 * regex patterns can be inserted directly into the filter content as text.
 */
export function PatternInsertButton({
  taRef,
  value,
  onChange,
  patternOptions,
}: {
  taRef: React.RefObject<HTMLTextAreaElement | null>
  value: string
  onChange: (v: string) => void
  patternOptions: RegexPattern[]
}) {
  const { t } = useTranslation()
  const wrapRef = useRef<HTMLDivElement>(null)
  const caretRef = useRef<number | null>(null)
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')

  // Close on outside click.
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Restore the caret in the textarea after a programmatic insertion.
  useLayoutEffect(() => {
    if (caretRef.current != null && taRef.current) {
      const pos = caretRef.current
      taRef.current.focus()
      taRef.current.setSelectionRange(pos, pos)
      caretRef.current = null
    }
  })

  const insert = (name: string) => {
    const ref = `{{.${name}}}`
    const ta = taRef.current
    const at = ta ? ta.selectionStart ?? value.length : value.length
    const next = value.slice(0, at) + ref + value.slice(at)
    caretRef.current = at + ref.length
    onChange(next)
    setOpen(false)
  }

  return (
    <div ref={wrapRef} className="relative">
      <button
        type="button"
        onClick={() => {
          setQuery('')
          setOpen((o) => !o)
        }}
        className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card/80 px-2 py-1 text-[11px] font-medium text-muted-foreground shadow-sm backdrop-blur hover:bg-muted hover:text-foreground"
      >
        <Braces size={13} /> {t('parsingFilters.visual.insertPattern')}
      </button>
      {open && (
        <div className="absolute right-0 top-[calc(100%+4px)] z-20 w-[340px] max-w-[90vw] overflow-hidden rounded-md border border-border bg-popover shadow-lg">
          <PatternMenu patternOptions={patternOptions} query={query} onQueryChange={setQuery} onPick={insert} />
        </div>
      )}
    </div>
  )
}
