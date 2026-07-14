import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowUp } from 'lucide-react'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { useSocAiConfigured } from '@/features/soc-ai/lib/useSocAiConfig'

const SUGGESTION_KEYS = ['threatHunt', 'observableTriage', 'writeRule', 'systemStatus'] as const

export function ChatHero() {
  const { t } = useTranslation()
  const configured = useSocAiConfigured()
  const { submit, homeMessages } = useSocAi()
  const [value, setValue] = useState('')

  // Hidden entirely until SOC-AI has a provider configured — same gate the
  // floating composer/panel use, so we never show a dead chat box.
  if (!configured) return null

  // No queueing — block sending while the last message is still being answered.
  const last = homeMessages[homeMessages.length - 1]
  const isPending = last?.role === 'ai' && !!last.pending

  const send = () => {
    const text = value.trim()
    if (!text || isPending) return
    submit(text, { openPanel: false, scope: 'home' })
    setValue('')
  }

  // Once a conversation is underway, dock the composer at the bottom —
  // centered, ChatGPT/Claude-style — instead of the big top "hero" box.
  const docked = homeMessages.length > 0

  if (docked) {
    return (
      <div className="fixed bottom-5 left-1/2 z-30 w-[min(760px,calc(100vw-2rem))] -translate-x-1/2">
        <div className="relative rounded-xl border border-border bg-card p-3 shadow-lg shadow-black/10">
          <textarea
            rows={1}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            placeholder={t('home.hero.placeholder')}
            className="max-h-40 w-full resize-none border-0 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
          />
          <div className="mt-1.5 flex items-center justify-end">
            <button
              onClick={send}
              disabled={!value.trim() || isPending}
              className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm hover:opacity-90 disabled:opacity-40"
              aria-label={t('home.hero.send')}
            >
              <ArrowUp size={15} strokeWidth={2.5} />
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <section className="flex flex-col items-center">
      <h2 className="mb-4 text-center text-base font-medium text-foreground">{t('home.hero.title')}</h2>
      <div className="relative w-full max-w-3xl">
        <div
          aria-hidden
          className="absolute -inset-px rounded-xl bg-[conic-gradient(from_180deg_at_50%_50%,#1a8cff_0deg,#0ea5e9_120deg,#2563eb_240deg,#1a8cff_360deg)] opacity-50 blur-[1px]"
        />
        <div className="relative rounded-xl bg-card p-4">
          <textarea
            rows={3}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            placeholder={t('home.hero.placeholder')}
            className="w-full resize-none border-0 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
          />
          <div className="mt-2 flex items-center justify-end">
            <button
              onClick={send}
              disabled={!value.trim()}
              className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm hover:opacity-90 disabled:opacity-40"
              aria-label={t('home.hero.send')}
            >
              <ArrowUp size={15} strokeWidth={2.5} />
            </button>
          </div>
        </div>
      </div>
      <div className="mt-4 flex flex-wrap justify-center gap-2">
        {SUGGESTION_KEYS.map((key) => {
          const label = t(`home.hero.suggestions.${key}`)
          return (
            <button
              key={key}
              onClick={() => submit(label, { openPanel: false, scope: 'home' })}
              className="rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              {label}
            </button>
          )
        })}
      </div>
    </section>
  )
}
