import { useState } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export interface ChipInputProps {
  values: string[]
  onChange: (v: string[]) => void
  placeholder?: string
  mono?: boolean
  /** Optional per-value validator. Returns null when accepted, or an error message when rejected. */
  validate?: (v: string) => string | null
  /** Called with the error message when a value is rejected. Cleared on next successful add. */
  onInvalid?: (message: string | null) => void
}

export function ChipInput({ values, onChange, placeholder, mono, validate, onInvalid }: ChipInputProps) {
  const [draft, setDraft] = useState('')
  const add = () => {
    const v = draft.trim()
    if (!v) {
      setDraft('')
      return
    }
    if (values.includes(v)) {
      setDraft('')
      onInvalid?.(null)
      return
    }
    if (validate) {
      const err = validate(v)
      if (err) {
        onInvalid?.(err)
        return
      }
    }
    onInvalid?.(null)
    onChange([...values, v])
    setDraft('')
  }
  return (
    <div className="flex min-h-8 flex-wrap items-center gap-1 rounded-md border border-input bg-background p-1">
      {values.map((v) => (
        <span key={v} className={cn('inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-[11px]', mono && 'font-mono')}>
          {v}
          <button onClick={() => onChange(values.filter((x) => x !== v))} className="text-muted-foreground hover:text-foreground"><X size={11} /></button>
        </span>
      ))}
      <input
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ',') { e.preventDefault(); add() } else if (e.key === 'Backspace' && !draft && values.length) onChange(values.slice(0, -1)) }}
        onBlur={add}
        placeholder={values.length ? '' : placeholder}
        className={cn('min-w-[80px] flex-1 bg-transparent px-1 text-xs outline-none', mono && 'font-mono')}
      />
    </div>
  )
}
