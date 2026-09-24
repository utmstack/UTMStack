import type { ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { composePage, SocAiProvider, useSocAi, useSocAiFocus, type SocAiFocus } from './SocAiProvider'

const sent: { task: string; page: string }[] = []

vi.mock('./lib/chat-stream', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./lib/chat-stream')>()),
  streamChat: (body: { task: string; page: string }) => {
    sent.push({ task: body.task, page: body.page })
    return new Promise<void>(() => {})
  },
}))

const alert: SocAiFocus = {
  kind: 'alert',
  id: 'a-1',
  label: 'Firewall rule deleted',
  context: 'The user has an alert open (id a-1).',
}

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient()
  return (
    <MemoryRouter initialEntries={['/threat-management/alerts']}>
      <QueryClientProvider client={client}>
        <SocAiProvider>{children}</SocAiProvider>
      </QueryClientProvider>
    </MemoryRouter>
  )
}

const chat = (focus: SocAiFocus | null) => () => {
  useSocAiFocus(focus)
  return useSocAi()
}

describe('composePage', () => {
  it('adds what is open to the page the agent is told about', () => {
    expect(composePage('Alerts page', alert)).toBe('Alerts page\n\nThe user has an alert open (id a-1).')
    expect(composePage('Alerts page', null)).toBe('Alerts page')
  })
})

describe('the item open beside the chat', () => {
  beforeEach(() => {
    sent.length = 0
  })

  it('is sent with the question while the drawer is open', () => {
    const { result } = renderHook(chat(alert), { wrapper })
    expect(result.current.focus?.label).toBe('Firewall rule deleted')

    act(() => result.current.submit('is this a false positive?'))

    expect(sent).toHaveLength(1)
    expect(sent[0].page).toContain('The user has an alert open (id a-1).')
  })

  it('stops being sent once the drawer closes', () => {
    const { result, rerender } = renderHook(({ f }) => chat(f)(), { wrapper, initialProps: { f: alert as SocAiFocus | null } })
    rerender({ f: null })

    expect(result.current.focus).toBeNull()
    act(() => result.current.submit('how many open alerts?'))

    expect(sent[0].page).not.toContain('open (id a-1)')
  })

  it('is dropped when the person removes it from the conversation, and shared again next time', () => {
    const { result, rerender } = renderHook(({ f }) => chat(f)(), { wrapper, initialProps: { f: alert as SocAiFocus | null } })

    act(() => result.current.detachFocus())
    expect(result.current.focus).toBeNull()
    act(() => result.current.submit('what is the weather?'))
    expect(sent[0].page).not.toContain('open (id a-1)')

    rerender({ f: null }) // the drawer closes
    rerender({ f: alert }) // and the same alert opens again
    expect(result.current.focus?.id).toBe('a-1')
  })
})
