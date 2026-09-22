import { describe, expect, it } from 'vitest'
import { humanizeToolLabel } from './tool-labels'

describe('humanizeToolLabel', () => {
  it('uses the curated label for a known tool', () => {
    expect(humanizeToolLabel('visualizations.create')).toBe('Adding widget')
    expect(humanizeToolLabel('dashboards.list')).toBe('Listing dashboards')
  })

  it('passes through non-tool text unchanged (e.g. the compaction placeholder)', () => {
    expect(humanizeToolLabel('Compacting conversation…')).toBe('Compacting conversation…')
  })

  it('falls back to a verb/resource humanization for an unmapped tool', () => {
    // Any future backend tool not yet added to TOOL_LABELS must still render
    // as a readable phrase, never as a raw dotted tool id.
    expect(humanizeToolLabel('widgets.archive')).toBe('Archive widgets')
    expect(humanizeToolLabel('adaudit.groups.list')).toBe('Listing adaudit groups')
  })
})
