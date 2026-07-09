import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowUp, Maximize2, Minimize2, Sparkles, Trash2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useSocAi } from '../SocAiProvider'
import { MessageRow } from './MessageRow'

export function SocAiPanel() {
  const { t } = useTranslation()
  const { open, expanded, messages, closePanel, toggleExpand, clear, submit } = useSocAi()
  const [draft, setDraft] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)

  // Stick to the bottom as messages stream in.
  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [messages])

  const send = () => {
    if (!draft.trim()) return
    submit(draft)
    setDraft('')
  }

  return (
    <aside
      role="dialog"
      aria-label="SOC Assistant"
      className={cn(
        'fixed inset-y-0 right-0 z-50 flex flex-col border-l border-border bg-background shadow-2xl transition-transform duration-200',
        open ? 'translate-x-0' : 'translate-x-full',
        expanded ? 'w-[min(720px,95vw)]' : 'w-[min(420px,95vw)]',
      )}
    >
      <header className="flex shrink-0 items-center justify-between border-b border-border px-4 py-3">
        <div className="flex items-center gap-2 text-[15px] font-semibold">
          <Sparkles size={18} className="text-primary" />
          <span>{t('socAi.chat.title')}</span>
        </div>
        <div className="flex items-center gap-0.5">
          <IconBtn label={expanded ? t('socAi.chat.collapse') : t('socAi.chat.expand')} onClick={toggleExpand}>
            {expanded ? <Minimize2 size={16} /> : <Maximize2 size={16} />}
          </IconBtn>
          <IconBtn label={t('socAi.chat.clear')} onClick={clear}>
            <Trash2 size={16} />
          </IconBtn>
          <IconBtn label={t('socAi.chat.close')} onClick={closePanel}>
            <X size={18} />
          </IconBtn>
        </div>
      </header>

      <div ref={scrollRef} className="flex flex-1 flex-col gap-4 overflow-y-auto px-4 py-5 text-[13.5px] leading-relaxed">
        {messages.length === 0 ? (
          <div className="m-auto max-w-[260px] text-center text-sm text-muted-foreground">
            <Sparkles size={22} className="mx-auto mb-3 text-primary/70" />
            <p>{t('socAi.chat.empty')}</p>
            <p className="mt-3 text-xs">
              {t('socAi.chat.tryPrefix')} <span className="text-foreground">“{t('socAi.chat.suggestion')}”</span>
            </p>
          </div>
        ) : (
          messages.map((m) => <MessageRow key={m.id} message={m} />)
        )}
      </div>

      <div className="shrink-0 border-t border-border bg-muted/30 p-3">
        <div className="relative rounded-lg border border-primary/40 bg-background p-2.5 pr-12 transition-colors focus-within:border-primary">
          <textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            rows={1}
            placeholder={t('socAi.chat.placeholder')}
            className="max-h-40 w-full resize-none bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
          />
          <button
            type="button"
            onClick={send}
            aria-label={t('socAi.chat.send')}
            className="absolute bottom-2 right-2 flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground transition-opacity hover:opacity-90 active:scale-95"
          >
            <ArrowUp size={16} strokeWidth={2.5} />
          </button>
        </div>
      </div>
    </aside>
  )
}


function IconBtn({ label, onClick, children }: { label: string; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
    >
      {children}
    </button>
  )
}
