import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Check, Copy, FilePlus2, Search, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'

/** One thing that can be copied. Keeps the modal usable for both catalogues. */
export interface StartFromOption {
  id: string
  name: string
}

/**
 * Asked before a new definition is created: start empty, or start from one that
 * already exists.
 *
 * Copying is the common case — a customer wants their PCI, not a blank page —
 * so it is offered where the work begins rather than hidden as an action on
 * something they were only browsing.
 *
 * Two steps, because they are two decisions. Putting the list on the same
 * screen as the choice makes the page argue with itself: a long list beside a
 * single button reads as though scrolling were the expected move.
 */
export function StartFromModal({
  title,
  options,
  onScratch,
  onCopy,
  onClose,
}: {
  title: string
  options: StartFromOption[]
  onScratch: () => void
  onCopy: (id: string) => void
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [step, setStep] = useState<'choose' | 'pick'>('choose')
  const [search, setSearch] = useState('')

  const shown = useMemo(() => {
    const q = search.trim().toLowerCase()
    const list = q
      ? options.filter((o) => o.name.toLowerCase().includes(q) || o.id.toLowerCase().includes(q))
      : options
    // Long catalogues are searched, not scrolled; the cap keeps the list from
    // becoming a second problem.
    return list.slice(0, 60)
  }, [options, search])

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className={cn(
          'flex w-full flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl',
          step === 'choose' ? 'max-w-xl' : 'max-h-[78vh] max-w-md',
        )}
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex shrink-0 items-center gap-3 border-b border-border px-5 py-4">
          {step === 'pick' && (
            <button
              onClick={() => setStep('choose')}
              aria-label={t('compliance.startFrom.back')}
              className="-ml-1 flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <ArrowLeft size={15} />
            </button>
          )}
          <h2 className="min-w-0 flex-1 truncate text-base font-semibold">
            {step === 'choose' ? title : t('compliance.startFrom.pickTitle')}
          </h2>
          <button
            onClick={onClose}
            aria-label={t('compliance.schedule.cancel')}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        {step === 'choose' ? (
          <div className="grid gap-3 p-5 sm:grid-cols-2">
            <ChoiceCard
              icon={<FilePlus2 size={19} />}
              title={t('compliance.startFrom.scratch')}
              body={t('compliance.startFrom.scratchHint')}
              onClick={onScratch}
            />
            <ChoiceCard
              icon={<Copy size={19} />}
              title={t('compliance.startFrom.copy')}
              body={t('compliance.startFrom.copyHint')}
              badge={t('compliance.startFrom.count', { n: options.length })}
              disabled={options.length === 0}
              onClick={() => setStep('pick')}
            />
          </div>
        ) : (
          <>
            <div className="shrink-0 px-5 pt-4">
              <div className="relative">
                <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder={t('compliance.startFrom.search')}
                  className="pl-8"
                  autoFocus
                />
              </div>
            </div>
            <div className="mt-2 min-h-0 flex-1 overflow-y-auto px-3 pb-4">
              {shown.length === 0 ? (
                <div className="py-10 text-center text-[12px] text-muted-foreground">{t('compliance.startFrom.none')}</div>
              ) : (
                shown.map((o) => (
                  <button
                    key={o.id}
                    onClick={() => onCopy(o.id)}
                    className="group flex w-full items-center gap-2 rounded-md px-2.5 py-2.5 text-left transition-colors hover:bg-muted/60"
                  >
                    {/* Just the name. An id and a description beside it turn a
                        choice into a reading exercise. */}
                    <span className="min-w-0 flex-1 truncate text-[13px]">{o.name}</span>
                    <Check size={14} className="shrink-0 text-primary opacity-0 transition-opacity group-hover:opacity-100" />
                  </button>
                ))
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

function ChoiceCard({
  icon,
  title,
  body,
  badge,
  disabled,
  onClick,
}: {
  icon: React.ReactNode
  title: string
  body: string
  badge?: string
  disabled?: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={cn(
        'group flex h-full flex-col items-start rounded-xl border border-border bg-background/40 p-4 text-left transition-all',
        disabled
          ? 'cursor-not-allowed opacity-50'
          : 'hover:-translate-y-0.5 hover:border-primary/50 hover:bg-muted/30 hover:shadow-sm',
      )}
    >
      <span
        className={cn(
          'flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary transition-colors',
          !disabled && 'group-hover:bg-primary group-hover:text-primary-foreground',
        )}
      >
        {icon}
      </span>
      <span className="mt-3 flex items-center gap-2 text-[13px] font-semibold">
        {title}
        {badge && (
          <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">{badge}</span>
        )}
      </span>
      <span className="mt-1 text-[11px] leading-relaxed text-muted-foreground">{body}</span>
    </button>
  )
}
