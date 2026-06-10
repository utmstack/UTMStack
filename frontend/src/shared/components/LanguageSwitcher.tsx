import { useEffect, useRef, useState } from 'react'
import { Check, Languages } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SUPPORTED_LANGUAGES } from '@/shared/i18n'

interface LanguageSwitcherProps {
  /** Fired after the language changes — e.g. to persist it to the user's lang_key. */
  onChange?: (code: string) => void
  align?: 'left' | 'right'
  className?: string
}

export function LanguageSwitcher({ onChange, align = 'right', className }: LanguageSwitcherProps) {
  const { i18n } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const current =
    SUPPORTED_LANGUAGES.find((l) => l.code === i18n.language) ??
    SUPPORTED_LANGUAGES.find((l) => i18n.language?.startsWith(l.code)) ??
    SUPPORTED_LANGUAGES[0]

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('mousedown', handler)
    return () => window.removeEventListener('mousedown', handler)
  }, [])

  const pick = (code: string) => {
    void i18n.changeLanguage(code)
    onChange?.(code)
    setOpen(false)
  }

  return (
    <div ref={ref} className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1.5 rounded-md px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <Languages size={14} />
        <span>{current?.label}</span>
      </button>
      {open && (
        <div
          className={cn(
            'absolute z-50 mt-1 min-w-[150px] overflow-hidden rounded-md border border-border bg-popover py-1 text-popover-foreground shadow-lg',
            align === 'right' ? 'right-0' : 'left-0',
          )}
        >
          {SUPPORTED_LANGUAGES.map((l) => {
            const active = l.code === current?.code
            return (
              <button
                key={l.code}
                type="button"
                onClick={() => pick(l.code)}
                className={cn(
                  'flex w-full items-center justify-between gap-3 px-3 py-1.5 text-left text-sm transition-colors hover:bg-muted',
                  active ? 'text-foreground' : 'text-popover-foreground/80',
                )}
              >
                {l.label}
                {active && <Check size={14} className="text-primary" />}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
