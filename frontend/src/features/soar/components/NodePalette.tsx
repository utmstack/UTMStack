import { useState } from 'react'
import { AlertOctagon, Bell, Brain, ChevronDown, GitBranch, Globe, Mail, Sparkles, Terminal, Zap } from 'lucide-react'
import { EXECUTOR_CATALOG, type ExecutorMeta, type NodeKind } from '../types/soar.types'
import { prebuiltShellNodes } from '../lib/command-templates'

const ICONS: Record<string, typeof Terminal> = {
  shell: Terminal,
  http: Globe,
  llm_enrich: Brain,
  llm_action: Zap,
  notify: Bell,
  incident: AlertOctagon,
  mail: Mail,
  conditional: GitBranch,
}

/** Platforms with their own prebuilt shell-node section in the palette.
 *  The section is collapsible — its presets only matter once expanded. */
const PREBUILT_PLATFORMS = ['windows', 'linux', 'darwin']

/** Palette of draggable node types. Each row is one (executor, kind) pair —
 *  the palette spells them out so the drag payload is unambiguous. */
export function NodePalette({ readOnly }: { readOnly?: boolean }) {
  const rows: Array<{ meta: ExecutorMeta; kind: NodeKind }> = []
  for (const meta of EXECUTOR_CATALOG) {
    for (const kind of meta.kinds) rows.push({ meta, kind })
  }

  const onDragStart = (event: React.DragEvent, payload: Record<string, unknown>) => {
    if (readOnly) return
    event.dataTransfer.setData('application/soar-node', JSON.stringify(payload))
    event.dataTransfer.effectAllowed = 'move'
  }

  return (
    <aside className="flex w-56 shrink-0 flex-col border-r border-border bg-card">
      <div className="border-b border-border px-3 py-2">
        <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Palette</div>
        <div className="mt-0.5 text-[11px] text-muted-foreground">Drag onto the canvas</div>
      </div>
      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {PREBUILT_PLATFORMS.map((platform) => (
          <PrebuiltSection key={platform} platform={platform} readOnly={readOnly} onDragStart={onDragStart} />
        ))}
        {rows.map(({ meta, kind }) => {
          const Icon = ICONS[meta.type] ?? Sparkles
          return (
            <div
              key={`${meta.type}:${kind}`}
              draggable={!readOnly}
              onDragStart={(e) => onDragStart(e, { executor: meta.type, kind, paramsPlaceholder: meta.paramsPlaceholder })}
              className={
                'flex cursor-grab items-center gap-2 rounded-md border border-border bg-background px-2 py-1.5 text-[11px] hover:border-primary/50 hover:bg-muted active:cursor-grabbing ' +
                (readOnly ? 'cursor-not-allowed opacity-50' : '')
              }
            >
              <span
                className={
                  'flex h-6 w-6 items-center justify-center rounded ' +
                  (kind === 'enrichment' ? 'bg-sky-500/15 text-sky-500' : 'bg-emerald-500/15 text-emerald-500')
                }
              >
                <Icon size={11} />
              </span>
              <div className="min-w-0">
                <div className="truncate font-medium">{meta.label}</div>
                <div className="truncate text-[9px] uppercase text-muted-foreground">{kind}</div>
              </div>
            </div>
          )
        })}
      </div>
    </aside>
  )
}

function PrebuiltSection({
  platform,
  readOnly,
  onDragStart,
}: {
  platform: string
  readOnly?: boolean
  onDragStart: (event: React.DragEvent, payload: Record<string, unknown>) => void
}) {
  const [open, setOpen] = useState(false)
  const presets = open ? prebuiltShellNodes(platform) : []

  return (
    <div className="rounded-md border border-border bg-background">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex w-full cursor-pointer items-center gap-2 px-2 py-1.5 text-[11px] hover:bg-muted"
      >
        <Terminal size={11} className="shrink-0 text-amber-500" />
        <span className="min-w-0 flex-1 truncate text-left font-medium capitalize">{platform}</span>
        <ChevronDown size={11} className={'shrink-0 opacity-60 transition-transform ' + (open ? 'rotate-180' : '')} />
      </button>
      {open && (
        <div className="space-y-1 border-t border-border p-1">
          {presets.map((p) => (
            <div
              key={p.id}
              draggable={!readOnly}
              onDragStart={(e) =>
                onDragStart(e, { executor: 'shell', kind: 'executor', command: p.command, platform: p.platform, shell: p.shell })
              }
              className={
                'flex cursor-grab items-center gap-1.5 rounded px-2 py-1 text-[10px] hover:border-primary/50 hover:bg-muted active:cursor-grabbing ' +
                (readOnly ? 'cursor-not-allowed opacity-50' : '')
              }
              title={p.command}
            >
              <span className="truncate font-medium">{p.label}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
