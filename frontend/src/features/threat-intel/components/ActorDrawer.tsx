import { Crosshair, ExternalLink, Search, UserX, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { useTiEntity } from '../hooks/use-ti-entity'
import { REPUTATION_STYLE, reputationTone, typeMeta } from './utils/severity-style'
import { absTimestamp } from './utils/time-format'
import { Section } from './Section'
import { Stat } from './Stat'

interface ActorDrawerProps {
  id: string | null
  onClose: () => void
}

export function ActorDrawer({ id, onClose }: ActorDrawerProps) {
  const { data, isLoading } = useTiEntity(id)

  if (!id) return null
  if (data?.kind === 'not-configured') return null
  if (isLoading || !data || data.kind !== 'ok') {
    return (
      <div
        className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
        onClick={onClose}
      >
        <div
          className="flex w-full max-w-[860px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="border-b border-border px-6 py-4">
            <div className="h-20 animate-pulse rounded bg-muted" />
          </div>
          <div className="flex-1 space-y-4 p-6">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-16 animate-pulse rounded bg-muted" />
            ))}
          </div>
        </div>
      </div>
    )
  }

  const a = data.value.attributes
  const tone = reputationTone(a.reputation_score)
  const rep = REPUTATION_STYLE[tone]
  const dangerous = tone === 'danger'
  const associations = data.value.latest_associations ?? []

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[860px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <div
                className={cn(
                  'flex h-11 w-11 shrink-0 items-center justify-center rounded-lg ring-2',
                  dangerous
                    ? 'bg-red-500/15 text-red-500 ring-red-500/40'
                    : 'bg-muted text-foreground/80 ring-border'
                )}
              >
                <UserX size={20} />
              </div>
              <div className="min-w-0 flex-1">
                <h2 className="truncate text-xl font-semibold">{a.value}</h2>
                {a.tags.length > 0 && (
                  <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
                    {a.tags.join(', ')}
                  </div>
                )}
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('rounded-md px-1.5 py-0.5 font-medium ring-1', rep.tone)}>
                    {a.reputation}
                  </span>
                  <span className="text-muted-foreground">Accuracy: {a.accuracy}</span>
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex items-center gap-2">
            <Button size="sm">
              <Crosshair size={13} className="mr-1.5" />
              Track campaign
            </Button>
            <Button size="sm" variant="outline">
              <Search size={13} className="mr-1.5" />
              View matched IOCs
            </Button>
            <Button size="sm" variant="outline">
              <ExternalLink size={13} className="mr-1.5" />
              MITRE ATT&CK
            </Button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
          <div className="space-y-4">
            {a.description && (
              <Section title="About">
                <p className="text-sm leading-relaxed">{a.description}</p>
              </Section>
            )}

            <Section title="Activity">
              <div className="grid grid-cols-3 gap-3 text-xs">
                <Stat label="Reputation" value={`${a.reputation_score}`} />
                <Stat label="Accuracy" value={a.accuracy} />
                <Stat label="Last seen" value={absTimestamp(a.last_seen)} />
              </div>
            </Section>

            {associations.length > 0 && (
              <Section title={`Associations (${associations.length})`}>
                <ul className="space-y-1.5">
                  {associations.slice(0, 10).map((r) => {
                    const rt = typeMeta(r.type)
                    const RIcon = rt.icon
                    return (
                      <li
                        key={r.id}
                        className="flex items-center justify-between gap-3 rounded border border-border bg-background/40 px-3 py-2 text-xs"
                      >
                        <div className="flex min-w-0 items-center gap-2">
                          <RIcon size={11} className={rt.tone} />
                          <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
                            {rt.label}
                          </span>
                          <span className="truncate font-mono">{r.value}</span>
                        </div>
                        <span className={cn('shrink-0 text-[11px]', REPUTATION_STYLE[reputationTone(r.reputation_score)].tone)}>
                          {r.reputation}
                        </span>
                      </li>
                    )
                  })}
                </ul>
              </Section>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
