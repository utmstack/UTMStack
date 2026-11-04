import { describe, expect, it } from 'vitest'
import {
  compactMessages,
  digestGroups,
  estimateTokens,
  groupMessages,
  renderForSummary,
  DEFAULT_COMPACTION,
} from './compaction'
import type { ChatMessage } from './types'

const user = (c: string): ChatMessage => ({ role: 'user', content: c })
const asst = (c: string): ChatMessage => ({ role: 'assistant', content: c })
const asstCall = (id: string, c = ''): ChatMessage => ({
  role: 'assistant',
  content: c || null,
  tool_calls: [{ id, type: 'function', function: { name: 'lookup_indicator', arguments: '{"value":"8.8.8.8"}' } }],
})
const toolMsg = (id: string, c: string): ChatMessage => ({ role: 'tool', tool_call_id: id, content: c })

describe('compaction', () => {
  it('counts tool_calls arguments in the estimate', () => {
    const plain = [user('x')]
    const withCalls = [asstCall('a1')]
    expect(estimateTokens(withCalls)).toBeGreaterThan(estimateTokens(plain))
  })

  it('groups an assistant tool call with every matching tool message', () => {
    const groups = groupMessages([user('q'), asstCall('a1'), toolMsg('a1', 'r1'), toolMsg('a1', 'r2'), asst('final')])
    expect(groups.length).toBe(3)
    expect(groups[1].messages.length).toBe(3) // asst + 2 tool messages
    expect(groups[2].messages[0].role).toBe('assistant')
  })

  it('does not compact under budget', async () => {
    const res = await compactMessages([user('hi'), asst('hello')], DEFAULT_COMPACTION)
    expect(res.compacted).toBe(false)
    expect(res.droppedGroups).toBe(0)
  })

  it('folds oldest groups into a summary, keeps recent verbatim', async () => {
    const big = 'filler '.repeat(2000)
    const messages: ChatMessage[] = [
      user('q1 ' + big),
      asst('a1'),
      user('q2 ' + big),
      asst('a2'),
      user('q3'),
      asst('a3'),
      user('q4'),
      asst('a4'),
    ]
    const config = { ...DEFAULT_COMPACTION, maxInputTokens: 1000, keepRecentGroups: 2 }
    const res = await compactMessages(messages, config)
    expect(res.compacted).toBe(true)
    expect(res.tokensAfter).toBeLessThanOrEqual(res.tokensBefore)
    // Summary message is present, recent groups verbatim.
    const summary = res.messages.find((m) => m.content?.includes('[Earlier conversation, condensed]'))
    expect(summary).toBeDefined()
    expect(res.messages.some((m) => m.content === 'a4')).toBe(true)
    expect(res.messages.some((m) => m.content === 'q1 ' + big)).toBe(false)
  })

  it('digest records questions and tools run', () => {
    const groups = groupMessages([user('what is 8.8.8.8?'), asstCall('a1'), toolMsg('a1', 'r'), asst('it is good')])
    const digest = digestGroups(groups)
    expect(digest).toContain('what is 8.8.8.8?')
    expect(digest).toContain('lookup_indicator')
    expect(digest).toContain('1 earlier answer(s)')
  })

  it('bounds the summary render to budget, oldest end first', () => {
    const groups = groupMessages([user('old ' + 'x'.repeat(5000)), user('new ' + 'y'.repeat(5000))])
    const out = renderForSummary(groups, 100) // ~400 chars budget
    expect(out.length).toBeLessThanOrEqual(100 * 4)
    expect(out).toContain('y') // recent survives
  })
})
