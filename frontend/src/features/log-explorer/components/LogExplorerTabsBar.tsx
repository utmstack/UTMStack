import { useEffect, useRef, useState } from 'react'
import { Plus, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { LogExplorerTabConfig } from '../types/log-explorer.types'

interface LogExplorerTabsBarProps {
  tabs: LogExplorerTabConfig[]
  activeId: string
  onSelect: (id: string) => void
  onAdd: () => void
  onClose: (id: string) => void
  onRename: (id: string, name: string) => void
}

export function LogExplorerTabsBar({
  tabs,
  activeId,
  onSelect,
  onAdd,
  onClose,
  onRename,
}: LogExplorerTabsBarProps) {
  const [editingId, setEditingId] = useState<string | null>(null)

  return (
    <div className="flex items-center gap-1 overflow-x-auto border-b border-border bg-card px-2 py-1.5">
      {tabs.map((tab) => (
        <Tab
          key={tab.id}
          tab={tab}
          active={tab.id === activeId}
          editing={editingId === tab.id}
          canClose={tabs.length > 1}
          onSelect={() => onSelect(tab.id)}
          onStartEdit={() => setEditingId(tab.id)}
          onEndEdit={(name) => {
            setEditingId(null)
            if (name != null) onRename(tab.id, name)
          }}
          onClose={() => onClose(tab.id)}
        />
      ))}
      <button
        onClick={onAdd}
        title="New tab"
        className="ml-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <Plus size={14} />
      </button>
    </div>
  )
}

interface TabProps {
  tab: LogExplorerTabConfig
  active: boolean
  editing: boolean
  canClose: boolean
  onSelect: () => void
  onStartEdit: () => void
  onEndEdit: (name: string | null) => void
  onClose: () => void
}

function Tab({ tab, active, editing, canClose, onSelect, onStartEdit, onEndEdit, onClose }: TabProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [draft, setDraft] = useState(tab.name)

  useEffect(() => {
    if (editing) {
      setDraft(tab.name)
      // Focus and select on next paint.
      requestAnimationFrame(() => {
        inputRef.current?.focus()
        inputRef.current?.select()
      })
    }
  }, [editing, tab.name])

  return (
    <div
      onClick={editing ? undefined : onSelect}
      onDoubleClick={onStartEdit}
      className={cn(
        'group/tab flex h-7 shrink-0 cursor-pointer items-center gap-1.5 rounded-md border px-2.5 text-xs transition-colors',
        active
          ? 'border-primary/30 bg-primary/10 text-foreground'
          : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground'
      )}
    >
      {editing ? (
        <input
          ref={inputRef}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => {
            if (e.key === 'Enter') onEndEdit(draft)
            else if (e.key === 'Escape') onEndEdit(null)
          }}
          onBlur={() => onEndEdit(draft)}
          className="h-5 max-w-[200px] rounded border border-border bg-background px-1.5 text-xs outline-none focus:border-primary"
        />
      ) : (
        <span className="max-w-[200px] truncate" title={tab.name}>
          {tab.name}
        </span>
      )}
      {canClose && !editing && (
        <button
          onClick={(e) => {
            e.stopPropagation()
            onClose()
          }}
          title="Close tab"
          className={cn(
            'flex h-4 w-4 items-center justify-center rounded transition-colors',
            active
              ? 'text-muted-foreground hover:bg-foreground/10 hover:text-foreground'
              : 'text-muted-foreground opacity-0 hover:bg-foreground/10 hover:text-foreground group-hover/tab:opacity-100'
          )}
        >
          <X size={11} />
        </button>
      )}
    </div>
  )
}
