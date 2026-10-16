import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { columnScope, useColumnLayouts } from './useColumnLayouts'

const initial = { dataset: 'logs', dataType: 'wineventlog', columns: ['log.eventCode', 'origin.ip'] }

describe('scoped column layouts', () => {
  it('restores each datatype order, isolates datasets and survives remount', () => {
    const { result, rerender, unmount } = renderHook(({ dataset, type }) => useColumnLayouts(dataset, type, initial), {
      initialProps: { dataset: 'logs', type: 'wineventlog' as string | null },
    })
    expect(result.current.columns).toEqual(initial.columns)
    act(() => result.current.setColumns(['origin.ip', 'log.eventCode']))
    rerender({ dataset: 'logs', type: 'firewall-fortigate-traffic' })
    expect(result.current.columns).toEqual([])
    act(() => result.current.setColumns(['action', 'target.ip']))
    rerender({ dataset: 'logs', type: 'wineventlog' })
    expect(result.current.columns).toEqual(['origin.ip', 'log.eventCode'])
    rerender({ dataset: 'other', type: 'wineventlog' })
    expect(result.current.columns).toEqual([])
    const layouts = JSON.parse(JSON.stringify(result.current.layouts))
    unmount()
    const restored = renderHook(() => useColumnLayouts('logs', 'firewall-fortigate-traffic', { ...initial, layouts }))
    expect(restored.result.current.columns).toEqual(['action', 'target.ip'])
  })

  it('resets and undoes the captured scope even after changing datatype', () => {
    const { result, rerender } = renderHook(({ type }) => useColumnLayouts('logs', type, initial), {
      initialProps: { type: 'wineventlog' },
    })
    const previous = result.current.columns
    const scope = result.current.scope
    act(() => result.current.setColumns([]))
    expect(result.current.columns).toEqual([])
    rerender({ type: 'vmware-esxi' })
    act(() => result.current.setColumns(['log.message']))
    act(() => result.current.setScopedColumns(scope, previous))
    expect(result.current.columns).toEqual(['log.message'])
    expect(result.current.layouts[columnScope('logs', 'wineventlog')]).toEqual(previous)
  })
})
