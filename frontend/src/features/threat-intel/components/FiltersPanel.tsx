import { useState } from 'react'
import { Filter, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { cn } from '@/shared/lib/utils'
import {
  DateRangePickerDialog,
  formatRange,
  type DateRangeValue,
} from '@/shared/components/ui/date-range-picker'
import type {
  EntityType,
  AdvancedSearchRequest,
  AdvancedCondition,
} from '@/features/threat-intel/domain/threat-intel.types'

const ALL_TYPES: EntityType[] = [
  'ip', 'domain', 'hostname', 'url',
  'md5', 'sha1', 'sha256', 'sha3-256',
  'cve', 'email-address', 'threat', 'malware',
]

export interface FiltersState {
  types: EntityType[]
  reputation: { min: number; max: number } | null
  accuracy: { min: number; max: number } | null
  tagsInclude: string[]
  tagsExclude: string[]
  dateFrom: string | null
  dateTo: string | null
}

export const EMPTY_FILTERS: FiltersState = {
  types: [],
  reputation: null,
  accuracy: null,
  tagsInclude: [],
  tagsExclude: [],
  dateFrom: null,
  dateTo: null,
}

export function filtersToRequest(
  state: FiltersState,
  extra?: AdvancedSearchRequest,
): AdvancedSearchRequest {
  const must: AdvancedCondition[] = [...(extra?.query?.must ?? [])]
  const mustNot: AdvancedCondition[] = [...(extra?.query?.must_not ?? [])]
  const filter: AdvancedCondition[] = [...(extra?.query?.filter ?? [])]
  const should: AdvancedCondition[] = [...(extra?.query?.should ?? [])]

  if (state.types.length > 0) {
    must.push({ terms: { type: state.types } })
  }
  if (state.reputation !== null) {
    filter.push({
      range: {
        reputation: {
          gte: String(state.reputation.min),
          lte: String(state.reputation.max),
        },
      },
    })
  }
  if (state.accuracy !== null) {
    filter.push({
      range: {
        accuracy: {
          gte: String(state.accuracy.min),
          lte: String(state.accuracy.max),
        },
      },
    })
  }
  if (state.tagsInclude.length > 0) {
    must.push({ terms: { tags: state.tagsInclude } })
  }
  if (state.tagsExclude.length > 0) {
    mustNot.push({ terms: { tags: state.tagsExclude } })
  }
  if (state.dateFrom || state.dateTo) {
    const rangeVal: { gte?: string; lte?: string } = {}
    if (state.dateFrom) rangeVal.gte = state.dateFrom
    if (state.dateTo) rangeVal.lte = state.dateTo
    filter.push({ range: { '@timestamp': rangeVal } })
  }

  const query: AdvancedSearchRequest['query'] = {}
  if (must.length) query.must = must
  if (should.length) query.should = should
  if (mustNot.length) query.must_not = mustNot
  if (filter.length) query.filter = filter

  return {
    query: Object.keys(query).length ? query : undefined,
    aggs: extra?.aggs,
  }
}

function countActive(state: FiltersState): number {
  let n = 0
  if (state.types.length) n++
  if (state.reputation) n++
  if (state.accuracy) n++
  if (state.tagsInclude.length) n++
  if (state.tagsExclude.length) n++
  if (state.dateFrom || state.dateTo) n++
  return n
}

export interface FiltersPanelProps {
  value: FiltersState
  onChange: (next: FiltersState) => void
  onApply: (request: AdvancedSearchRequest) => void
  onClear: () => void
  onClose?: () => void
}

export function FiltersPanel({ value, onChange, onApply, onClear, onClose }: FiltersPanelProps) {
  const [datePickerOpen, setDatePickerOpen] = useState(false)

  const dateRangeValue: DateRangeValue = {
    from: value.dateFrom ? new Date(value.dateFrom) : null,
    to: value.dateTo ? new Date(value.dateTo) : null,
  }

  const toggleType = (t: EntityType) => {
    const next = value.types.includes(t)
      ? value.types.filter((x) => x !== t)
      : [...value.types, t]
    onChange({ ...value, types: next })
  }

  const setRepField = (field: 'min' | 'max', raw: string) => {
    const n = raw === '' ? (field === 'min' ? -3 : 3) : Math.max(-3, Math.min(3, Number(raw)))
    const base = value.reputation ?? { min: -3, max: 3 }
    const next = { ...base, [field]: n }
    onChange({ ...value, reputation: next.min === -3 && next.max === 3 ? null : next })
  }

  const setAccField = (field: 'min' | 'max', raw: string) => {
    const n = raw === '' ? (field === 'min' ? 0 : 3) : Math.max(0, Math.min(3, Number(raw)))
    const base = value.accuracy ?? { min: 0, max: 3 }
    const next = { ...base, [field]: n }
    onChange({ ...value, accuracy: next.min === 0 && next.max === 3 ? null : next })
  }

  const parseTags = (raw: string) =>
    raw.split(',').map((s) => s.trim()).filter(Boolean)

  const active = countActive(value)

  return (
    <div className="rounded-xl border border-border bg-card p-4 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Filter size={12} className="text-fuchsia-500" />
          Filters
          {active > 0 && (
            <span className="rounded-full bg-fuchsia-500/20 px-1.5 py-0.5 text-[10px] text-fuchsia-400">
              {active}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={onClear}>
            Clear
          </Button>
          <Button size="sm" onClick={() => onApply(filtersToRequest(value))}>
            Apply
          </Button>
          {onClose && (
            <button
              type="button"
              onClick={onClose}
              className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={14} />
            </button>
          )}
        </div>
      </div>

      <div className="space-y-1">
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Types</p>
        <div className="flex flex-wrap gap-1.5">
          {ALL_TYPES.map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => toggleType(t)}
              className={cn(
                'rounded-full border px-2.5 py-0.5 text-xs transition-colors',
                value.types.includes(t)
                  ? 'border-fuchsia-500 bg-fuchsia-500/20 text-fuchsia-300'
                  : 'border-border text-muted-foreground hover:border-fuchsia-400 hover:text-foreground',
              )}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Reputation <span className="normal-case">(−3 to 3)</span>
          </p>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min={-3}
              max={3}
              placeholder="min"
              value={value.reputation?.min ?? ''}
              onChange={(e) => setRepField('min', e.target.value)}
              className="h-8 text-sm"
            />
            <span className="text-muted-foreground">–</span>
            <Input
              type="number"
              min={-3}
              max={3}
              placeholder="max"
              value={value.reputation?.max ?? ''}
              onChange={(e) => setRepField('max', e.target.value)}
              className="h-8 text-sm"
            />
          </div>
        </div>

        <div className="space-y-1">
          <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Accuracy <span className="normal-case">(0 to 3)</span>
          </p>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min={0}
              max={3}
              placeholder="min"
              value={value.accuracy?.min ?? ''}
              onChange={(e) => setAccField('min', e.target.value)}
              className="h-8 text-sm"
            />
            <span className="text-muted-foreground">–</span>
            <Input
              type="number"
              min={0}
              max={3}
              placeholder="max"
              value={value.accuracy?.max ?? ''}
              onChange={(e) => setAccField('max', e.target.value)}
              className="h-8 text-sm"
            />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Tags include</p>
          <Input
            placeholder="tag1, tag2"
            value={value.tagsInclude.join(', ')}
            onChange={(e) => onChange({ ...value, tagsInclude: parseTags(e.target.value) })}
            className="h-8 text-sm"
          />
        </div>
        <div className="space-y-1">
          <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Tags exclude</p>
          <Input
            placeholder="tag1, tag2"
            value={value.tagsExclude.join(', ')}
            onChange={(e) => onChange({ ...value, tagsExclude: parseTags(e.target.value) })}
            className="h-8 text-sm"
          />
        </div>
      </div>

      <div className="space-y-1">
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Date range</p>
        <button
          type="button"
          onClick={() => setDatePickerOpen(true)}
          className={cn(
            'w-full rounded-md border border-border bg-background px-3 py-1.5 text-left text-sm transition-colors hover:border-fuchsia-400',
            !value.dateFrom && !value.dateTo && 'text-muted-foreground',
          )}
        >
          {formatRange(dateRangeValue, 'Pick a date range…')}
        </button>
      </div>

      <DateRangePickerDialog
        open={datePickerOpen}
        value={dateRangeValue}
        title="Date range (@timestamp)"
        onClose={() => setDatePickerOpen(false)}
        onConfirm={(next) => {
          onChange({
            ...value,
            dateFrom: next.from ? next.from.toISOString() : null,
            dateTo: next.to ? next.to.toISOString() : null,
          })
          setDatePickerOpen(false)
        }}
      />
    </div>
  )
}
