import type { EntitySummary } from '../domain/threat-intel.types'
import { ActorCard } from './ActorCard'

interface ActorsListProps {
  actors: EntitySummary[]
  onOpen: (id: string) => void
  isLoading?: boolean
}

export function ActorsList({ actors, onOpen, isLoading }: ActorsListProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <div
            key={i}
            className="h-48 animate-pulse rounded-xl border border-border bg-muted"
          />
        ))}
      </div>
    )
  }

  if (actors.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
        No actors found.
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
      {actors.map((actor) => (
        <ActorCard
          key={actor.id}
          actor={actor}
          onOpen={() => onOpen(actor.id)}
        />
      ))}
    </div>
  )
}
