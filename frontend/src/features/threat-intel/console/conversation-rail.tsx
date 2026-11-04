import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Folder, MessageSquarePlus, Trash2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import {
  byRecency,
  createConversation,
  deleteConversation,
  useConsoleStore,
} from './store'
import { runningConversations, unreadCounts } from './use-conversation-loop'
import type { Conversation } from './types'

export function ConversationRail({
  conversationId,
  onSelectConversation,
  className,
}: {
  conversationId: string | null
  onSelectConversation(id: string | null): void
  className?: string
}) {
  const { t } = useTranslation()
  const store = useConsoleStore()
  const [creating, setCreating] = useState(false)
  const [title, setTitle] = useState('')

  const ordered = useMemo(
    () => [...store.conversations].sort(byRecency),
    [store.conversations],
  )

  const startConversation = () => {
    const trimmed = title.trim()
    if (!trimmed) return
    const conversation = createConversation(trimmed)
    setTitle('')
    setCreating(false)
    onSelectConversation(conversation.id)
  }

  return (
    <nav
      aria-label={t('tiAnalyst.conv.title')}
      className={cn(
        'flex min-h-0 w-56 shrink-0 flex-col border-r border-border bg-card',
        className,
      )}
    >
      <div className="shrink-0 border-b border-border p-2">
        <button
          type="button"
          className="flex w-full items-center gap-1.5 rounded-md border border-border bg-muted/50 px-1.5 py-1 font-mono text-[10px] text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground"
        >
          <Folder className="h-3 w-3 shrink-0" />
          <span className="min-w-0 flex-1 truncate text-start">{t('tiAnalyst.conv.allConversations')}</span>
        </button>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-1.5">
        {!ordered.length && (
          <p className="px-1.5 py-3 font-mono text-[10px] text-muted-foreground">{t('tiAnalyst.conv.empty')}</p>
        )}
        <ul className="space-y-0.5">
          {ordered.map((conversation) => (
            <ConversationRow
              key={conversation.id}
              conversation={conversation}
              active={conversation.id === conversationId}
              unread={runningUnread(conversation.id)}
              running={runningNow(conversation.id)}
              onSelect={() => onSelectConversation(conversation.id)}
              onDelete={() => {
                deleteConversation(conversation.id)
                if (conversation.id === conversationId) onSelectConversation(null)
              }}
            />
          ))}
        </ul>
      </div>

      <div className="shrink-0 border-t border-border p-1.5">
        {creating ? (
          <form
            onSubmit={(e) => {
              e.preventDefault()
              startConversation()
            }}
            className="space-y-1"
          >
            <label htmlFor="ti-new-conversation" className="sr-only">
              {t('tiAnalyst.conv.newTitle')}
            </label>
            <input
              id="ti-new-conversation"
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              onBlur={() => !title.trim() && setCreating(false)}
              placeholder={t('tiAnalyst.conv.newTitle')}
              className="w-full rounded-md border border-border bg-muted/50 px-2 py-1 font-mono text-[11px] focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </form>
        ) : (
          <button
            type="button"
            onClick={() => setCreating(true)}
            className="flex w-full items-center gap-1.5 rounded-md border border-border px-2 py-1.5 text-[11px] uppercase tracking-wider text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground"
          >
            <MessageSquarePlus className="h-3 w-3" />
            {t('tiAnalyst.conv.new')}
          </button>
        )}
      </div>
    </nav>
  )
}

function ConversationRow({
  conversation,
  active,
  unread,
  running,
  onSelect,
  onDelete,
}: {
  conversation: Conversation
  active: boolean
  unread: number
  running: boolean
  onSelect(): void
  onDelete(): void
}) {
  const { t } = useTranslation()

  return (
    <li className="group relative">
      <button
        type="button"
        onClick={onSelect}
        aria-current={active ? 'true' : undefined}
        className={cn(
          'flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-start transition-colors',
          active ? 'bg-primary/10 text-foreground' : 'text-muted-foreground hover:bg-muted/50',
        )}
      >
        {running && (
          <span className="h-1.5 w-1.5 shrink-0 animate-pulse rounded-full bg-primary" aria-label={t('tiAnalyst.conv.running')} />
        )}
        <span className="min-w-0 flex-1 truncate font-mono text-[11px]">{conversation.title}</span>
        {unread > 0 && (
          <span
            className="shrink-0 rounded-full bg-primary px-1.5 font-mono text-[9px] font-semibold text-background"
            aria-label={t('tiAnalyst.conv.unread', { count: unread })}
          >
            {unread}
          </span>
        )}
      </button>

      <button
        type="button"
        onClick={onDelete}
        aria-label={t('tiAnalyst.conv.delete', { title: conversation.title })}
        className="absolute end-1 top-1/2 hidden -translate-y-1/2 rounded-md p-1 text-muted-foreground transition-colors hover:text-destructive focus-visible:block group-hover:block"
      >
        <Trash2 className="h-3 w-3" />
      </button>
    </li>
  )
}

function runningNow(id: string): boolean {
  return runningConversations().has(id)
}

function runningUnread(id: string): number {
  return unreadCounts()[id] ?? 0
}
