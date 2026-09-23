import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'
import type { FlowNode } from '../types/soar.types'
import { clearHttpParamsCache, HttpParamsEditor } from './HttpParamsEditor'

const NODE_ID = 'http'
const nodes: Record<string, FlowNode> = {
  [NODE_ID]: { kind: 'enrichment', executor: 'http' },
}
const defaultParams = {
  method: 'GET',
  url: '',
  headers: {
    'Content-Type': 'application/json',
    Accept: '*/*',
    'Accept-Encoding': 'gzip, deflate, br',
    Connection: 'keep-alive',
  },
}

afterEach(() => {
  cleanup()
  clearHttpParamsCache(NODE_ID)
})

describe('HTTP parameter row cache', () => {
  test('does not restore deleted node configuration when its id is reused', () => {
    const configured = render(
      <HttpParamsEditor
        nodeId={NODE_ID}
        nodes={nodes}
        params={{
          method: 'GET',
          url: 'https://example.test/path?oldParam=old-query-value',
          headers: { 'X-Old-Header': 'old-header-value' },
        }}
        onChange={vi.fn()}
      />,
    )

    expect(screen.getByDisplayValue('old-header-value')).toBeInTheDocument()
    expect(screen.getByDisplayValue('old-query-value')).toBeInTheDocument()
    configured.unmount()

    const remounted = render(
      <HttpParamsEditor
        nodeId={NODE_ID}
        nodes={nodes}
        params={defaultParams}
        onChange={vi.fn()}
      />,
    )

    expect(screen.getByDisplayValue('old-header-value')).toBeInTheDocument()
    expect(screen.getByDisplayValue('old-query-value')).toBeInTheDocument()

    clearHttpParamsCache(NODE_ID)
    remounted.unmount()

    render(
      <HttpParamsEditor
        nodeId={NODE_ID}
        nodes={nodes}
        params={defaultParams}
        onChange={vi.fn()}
      />,
    )

    expect(screen.queryByDisplayValue('old-header-value')).not.toBeInTheDocument()
    expect(screen.queryByDisplayValue('old-query-value')).not.toBeInTheDocument()
    expect(screen.getByDisplayValue('application/json')).toBeInTheDocument()
  })
})