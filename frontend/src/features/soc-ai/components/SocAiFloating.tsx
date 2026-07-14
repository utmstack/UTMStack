import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowUp, Sparkles } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useSocAi } from '../SocAiProvider'

/** Always-visible chat pill, centered at the bottom. Submitting opens the panel. */
export function SocAiFloating() {
  const { t } = useTranslation()
  const { open, submit, messages } = useSocAi()
  const [value, setValue] = useState('')

  // No queueing — block sending while the last message is still being answered.
  const last = messages[messages.length - 1]
  const isPending = last?.role === 'ai' && !!last.pending

  const send = (e: React.FormEvent) => {
    e.preventDefault()
    if (!value.trim() || isPending) return
    submit(value)
    setValue('')
  }

  return (
    <div
      className={cn(
        'fixed bottom-5 left-1/2 z-40 w-[min(560px,calc(100vw-2rem))] -translate-x-1/2 transition-all duration-200',
        open ? 'pointer-events-none translate-y-5 opacity-0' : 'translate-y-0 opacity-100',
      )}
    >
      <form
        onSubmit={send}
        className="flex items-center gap-2.5 rounded-full border border-border bg-background py-2 pl-4 pr-2 shadow-lg shadow-black/10 transition-colors focus-within:border-primary/60"
      >
        <Sparkles size={16} className="shrink-0 text-primary" />
        <input
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={t('socAi.chat.floatingPlaceholder')}
          aria-label={t('socAi.chat.floatingPlaceholder')}
          className="min-w-0 flex-1 bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
        />
        <button
          type="submit"
          disabled={!value.trim() || isPending}
          aria-label={t('socAi.chat.send')}
          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground transition-opacity hover:opacity-90 active:scale-95 disabled:opacity-40"
        >
          <ArrowUp size={15} strokeWidth={2.5} />
        </button>
      </form>
    </div>
  )
}
