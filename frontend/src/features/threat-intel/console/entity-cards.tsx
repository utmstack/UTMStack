import { Check, Copy, MapPin } from 'lucide-react'
import { useState } from 'react'
import type { EntityDetailAttributes, EntityHistoryRecord, EntityDetails } from './types'

/**
 * Entity cards for the inspector, ported from the ThreatWinds analyst
 * console. The inspector renders one flat dossier; these break it into the
 * cards TW uses so a verdict, metadata, history and geo each get their own
 * scan-able block.
 */

const SEVERITY: Record<number, { label: string; cls: string; bar: string }> = {
  '-3': { label: 'Alarming', cls: 'text-destructive border-destructive/40 bg-destructive/10', bar: 'bg-destructive' },
  '-2': { label: 'Worrying', cls: 'text-warning border-warning/40 bg-warning/10', bar: 'bg-warning' },
  '-1': { label: 'Concerning', cls: 'text-warning border-warning/40 bg-warning/10', bar: 'bg-warning' },
  0: { label: 'Indefinable', cls: 'text-muted-foreground border-border bg-muted/40', bar: 'bg-muted-foreground' },
  1: { label: 'Adequate', cls: 'text-primary border-primary/40 bg-primary/10', bar: 'bg-primary' },
  2: { label: 'Good', cls: 'text-primary border-primary/40 bg-primary/10', bar: 'bg-primary' },
  3: { label: 'Exceptional', cls: 'text-primary border-primary/40 bg-primary/10', bar: 'bg-primary' },
}

export function severityOf(score: number) {
  return SEVERITY[Math.max(-3, Math.min(3, Math.round(score)))] ?? SEVERITY[0]
}

/**
 * The verdict, stated in words and in figures, coloured by severity. One
 * card instead of four near-identical gauges: the score leads, the accuracy
 * rides along, and the spread shows only when sources disagree.
 */
export function EntityVerdict({ attributes }: { attributes: EntityDetailAttributes }) {
  const [copied, setCopied] = useState(false)
  const s = severityOf(attributes.reputation_score)

  async function copyValue() {
    try {
      await navigator.clipboard.writeText(attributes.value)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard access can be refused outright. Silence is fine — the value
      // is still selectable text.
    }
  }

  const spread =
    attributes.worst_reputation_score !== attributes.best_reputation_score
      ? `${attributes.worst_reputation_score}…${attributes.best_reputation_score}`
      : null

  return (
    <div className="space-y-2 rounded-md border border-border bg-card p-3">
      <div className="flex flex-wrap items-center gap-2">
        <span
          className={`rounded-md border px-2 py-0.5 font-mono text-xs font-bold ${s.cls}`}
        >
          {s.label}
        </span>
        <span className="font-mono text-lg font-bold text-foreground" dir="ltr">
          {attributes.reputation_score > 0 ? `+${attributes.reputation_score}` : attributes.reputation_score}
        </span>
        {spread && (
          <span className="font-mono text-[10px] text-muted-foreground" dir="ltr">
            spread {spread}
          </span>
        )}
        <span
          className="ms-auto font-mono text-[10px] text-muted-foreground"
          title="accuracy of the verdict"
        >
          accuracy {attributes.accuracy_score}/3
        </span>
        <button
          type="button"
          onClick={copyValue}
          className="shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:text-primary"
          aria-label="copy value"
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
        </button>
      </div>
      <div className="h-1 overflow-hidden rounded-full bg-muted/60">
        <div
          className={`h-full ${s.bar}`}
          style={{ width: `${((attributes.reputation_score + 3) / 6) * 100}%` }}
        />
      </div>
      {attributes.description && (
        <p className="font-mono text-xs leading-relaxed text-muted-foreground">{attributes.description}</p>
      )}
    </div>
  )
}

/** Flat metadata chips, as before — only rendered when metadata exists. */
export function MetadataCard({ metadata }: { metadata: Record<string, string> }) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {Object.entries(metadata).map(([k, v]) => (
        <span key={k} className="rounded-md border border-border px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
          {k}: <span className="text-foreground/80">{v}</span>
        </span>
      ))}
    </div>
  )
}

/** Grouped extended metadata, when the corpus carries it. */
export function ExtendedMetadataCard({ entries }: { entries: Record<string, Record<string, string>> }) {
  const groups = Object.entries(entries)
  if (!groups.length) return null
  return (
    <div className="rounded-md border border-border bg-card p-3">
      <h4 className="font-mono text-[10px] uppercase tracking-widest text-primary">Extended metadata</h4>
      {groups.map(([group, fields], i) => (
        <div key={group} className={i > 0 ? 'mt-2 border-t border-border/30 pt-2' : 'mt-2'}>
          <h5 className="font-mono text-[10px] font-semibold uppercase tracking-widest text-primary/70">{group}</h5>
          <dl className="mt-1">
            {Object.entries(fields).map(([key, val]) => (
              <div key={key} className="flex gap-3 py-0.5">
                <dt className="w-28 shrink-0 font-mono text-[10px] text-muted-foreground">{key}</dt>
                <dd className="min-w-0 flex-1 break-words font-mono text-[10px] text-foreground/90">{val}</dd>
              </div>
            ))}
          </dl>
        </div>
      ))}
    </div>
  )
}

/** One reported version of the entity. Reputation movement is the story. */
export function HistoryRecordCard({ record, previous }: { record: EntityHistoryRecord; previous?: EntityHistoryRecord }) {
  const changed = previous !== undefined && previous.reputation !== record.reputation
  const ts = Date.parse(record.timestamp)
  return (
    <div className={`mb-1.5 flex items-center gap-3 rounded-md border px-3 py-2 ${changed ? 'border-primary/40' : 'border-border'} bg-card`}>
      <span
        dir="ltr"
        className={`rounded-md border px-1.5 py-0.5 font-mono text-[10px] ${record.reputation < 0 ? 'border-destructive/40 bg-destructive/10 text-destructive' : 'border-border text-muted-foreground'}`}
      >
        {record.reputation > 0 ? `+${record.reputation}` : record.reputation}
      </span>
      <div className="min-w-0 flex-1">
        <div className="truncate font-mono text-xs text-foreground/90">
          reported by {record.author || 'unknown'}
        </div>
        <div className="font-mono text-[10px] text-muted-foreground">
          {Number.isNaN(ts) ? '—' : new Date(ts).toLocaleString()}
        </div>
      </div>
      {changed && (
        <span className="shrink-0 font-mono text-[10px] text-primary">reputation changed</span>
      )}
    </div>
  )
}

/** One geolocation, as a card rather than a bare row. */
export function GeoCard({ latitude, longitude, radius }: { latitude: number; longitude: number; radius: number }) {
  return (
    <div className="mb-1.5 flex items-center gap-3 rounded-md border border-border bg-card px-3 py-2">
      <MapPin className="h-4 w-4 shrink-0 text-primary/60" />
      <div className="min-w-0 flex-1 font-mono text-xs">
        <div dir="ltr" className="text-foreground/90">
          {latitude.toFixed(6)}°, {longitude.toFixed(6)}°
        </div>
        <div className="text-muted-foreground">r={radius}m</div>
      </div>
    </div>
  )
}

export function firstSeenLabel(details: EntityDetails): string | null {
  const iso = details.attributes.first_seen
  const ts = iso ? Date.parse(iso) : NaN
  return Number.isNaN(ts) ? null : new Date(ts).toLocaleDateString()
}
