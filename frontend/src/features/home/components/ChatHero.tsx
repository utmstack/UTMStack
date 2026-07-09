import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AtSign, Paperclip, Sparkles } from 'lucide-react'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { useSocAiConfigured } from '@/features/soc-ai/lib/useSocAiConfig'
import { IconBtn } from './IconBtn'

const SUGGESTION_KEYS = ['threatHunt', 'observableTriage', 'writeRule', 'systemStatus'] as const

export function ChatHero() {
  const { t } = useTranslation()
  const configured = useSocAiConfigured()
  const { submit } = useSocAi()
  const [value, setValue] = useState('')

  // Hidden entirely until SOC-AI has a provider configured — same gate the
  // floating composer/panel use, so we never show a dead chat box.
  if (!configured) return null

  const send = () => {
    const text = value.trim()
    if (!text) return
    submit(text)
    setValue('')
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
          <div className="mt-2 flex items-center justify-between">
            <div className="flex items-center gap-2 text-muted-foreground">
              <IconBtn label={t('home.hero.mention')}><AtSign size={16} strokeWidth={1.75} /></IconBtn>
              <IconBtn label={t('home.hero.attach')}><Paperclip size={16} strokeWidth={1.75} /></IconBtn>
            </div>
            <button
              onClick={send}
              disabled={!value.trim()}
              className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm hover:opacity-90 disabled:opacity-40"
              aria-label={t('home.hero.send')}
            >
              <Sparkles size={14} strokeWidth={2} />
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
              onClick={() => submit(label)}
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
