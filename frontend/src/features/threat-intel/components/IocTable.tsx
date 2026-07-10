import type { EntitySummary } from '../domain/threat-intel.types'
import { IocRow } from './IocRow'

interface IocTableProps {
  iocs: EntitySummary[]
  onOpen: (id: string) => void
  isLoading?: boolean
}

const IOC_COLS = '4px 90px 1fr 130px 1fr 110px 36px'

export function IocTable({ iocs, onOpen, isLoading }: IocTableProps) {
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: IOC_COLS }}
      >
        <div />
        <div>Type</div>
        <div>Indicator</div>
        <div className="text-right">Reputation</div>
        <div>Tags</div>
        <div>Last seen</div>
        <div />
      </div>
      {iocs.map((ioc) => (
        <IocRow key={ioc.id} ioc={ioc} onOpen={() => onOpen(ioc.id)} />
      ))}
      {!isLoading && iocs.length === 0 && (
        <div className="px-6 py-16 text-center text-sm text-muted-foreground">No IOCs match.</div>
      )}
      {isLoading && (
        <div className="px-6 py-16 text-center text-sm text-muted-foreground">Loading…</div>
      )}
    </div>
  )
}
