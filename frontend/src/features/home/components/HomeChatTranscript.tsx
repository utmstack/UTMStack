import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Trash2 } from 'lucide-react'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { MessageRow } from '@/features/soc-ai/components/MessageRow'

export function HomeChatTranscript() {
  const { t } = useTranslation()
  const { homeMessages, clear } = useSocAi()
  const bottomRef = useRef<HTMLDivElement>(null)

  // Stick to bottom: on mount (navigating back with existing messages) and as
  // new messages stream in. Works regardless of which ancestor scrolls.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: 'end' })
  }, [homeMessages])

  return (
    <section className="mx-auto max-w-3xl pb-28">
      <div className="flex justify-end">
        <button
          type="button"
          onClick={() => clear('home')}
          className="inline-flex items-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:border-border hover:bg-muted hover:text-foreground"
        >
          <Trash2 size={13} />
          {t('socAi.chat.clear')}
        </button>
      </div>
      <div className="mt-4 flex flex-col gap-4 text-[13.5px] leading-relaxed">
        {homeMessages.map((m) => (
          <MessageRow key={m.id} message={m} />
        ))}
      </div>
      <div ref={bottomRef} />
    </section>
  )
}
