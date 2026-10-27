import type { ChatMessage } from './types'

/**
 * Conversation compaction for the agent loop.
 *
 * A long investigation accumulates tool traffic fast and the model re-sends it
 * every turn. Left alone the request exceeds the model's input limit and the
 * run fails mid-investigation. Compaction replaces the oldest turns with a
 * summary and keeps the recent ones verbatim.
 *
 * Pure: it does no networking. The caller supplies a summarizer (or omits one,
 * in which case a deterministic digest is used). This keeps the logic testable
 * and free of a dependency on the transport.
 */

export interface CompactionConfig {
  /** Compact once the estimated input exceeds this many tokens. */
  maxInputTokens: number
  /** How much of the budget to aim for after compacting, as a fraction. */
  targetRatio: number
  /** Turn groups always kept verbatim, however large. */
  keepRecentGroups: number
}

export const DEFAULT_COMPACTION: CompactionConfig = {
  // The server-side model's advertised limit is not plumbed to the client, so
  // this is the standing budget. Overridable per run.
  maxInputTokens: 96_000,
  targetRatio: 0.5,
  keepRecentGroups: 4,
}

/** ~4 characters per token; the per-message constant covers role/delimiter overhead. */
export function estimateTokens(messages: ChatMessage[]): number {
  const PER_MESSAGE_OVERHEAD = 4
  let chars = 0
  for (const message of messages) {
    chars += (message.content ?? '').length
    if (message.role === 'assistant' && message.tool_calls) {
      for (const call of message.tool_calls) {
        chars += (call.function?.name ?? '').length
        chars += (call.function?.arguments ?? '').length
      }
    }
    chars += PER_MESSAGE_OVERHEAD * 4
  }
  return Math.ceil(chars / 4)
}

/**
 * A contiguous run of messages that must survive or be dropped together.
 * The API rejects a request where an assistant message with `tool_calls` is
 * not followed by a `tool` message for every call id. Naively slicing by count
 * will cut between the two and produce a 400 that looks like a model problem.
 */
export interface MessageGroup {
  messages: ChatMessage[]
}

export function groupMessages(messages: ChatMessage[]): MessageGroup[] {
  const groups: MessageGroup[] = []
  let current: ChatMessage[] = []
  const flush = () => {
    if (current.length > 0) {
      groups.push({ messages: current })
      current = []
    }
  }
  for (const message of messages) {
    if (message.role === 'tool') {
      // Belongs with the assistant turn that requested it.
      current.push(message)
      continue
    }
    flush()
    current = [message]
  }
  flush()
  return groups
}

/** Deterministic digest used when no summarizer is supplied or one fails. */
export function digestGroups(groups: MessageGroup[]): string {
  const questions: string[] = []
  const toolCalls: string[] = []
  let assistantAnswers = 0
  for (const group of groups) {
    for (const message of group.messages) {
      if (message.role === 'user' && message.content) {
        questions.push(message.content.slice(0, 200))
      }
      if (message.role === 'assistant') {
        if (message.tool_calls) {
          for (const call of message.tool_calls) {
            const args = (call.function?.arguments ?? '').slice(0, 120)
            toolCalls.push(`${call.function?.name ?? 'unknown'}(${args})`)
          }
        } else if (message.content) {
          assistantAnswers += 1
        }
      }
    }
  }
  const lines = ['[Earlier conversation, condensed]']
  if (questions.length > 0) lines.push(`Questions asked: ${questions.join(' | ')}`)
  if (toolCalls.length > 0) lines.push(`Tools already run: ${toolCalls.join(', ')}`)
  if (assistantAnswers > 0) lines.push(`${assistantAnswers} earlier answer(s) were given.`)
  lines.push('Do not repeat tool calls already made above unless the data may have changed.')
  return lines.join('\n')
}

/** Fraction of the input budget a summarizer's own request may use. */
export const SUMMARY_INPUT_RATIO = 0.5
/** ~4 characters per token, matching estimateTokens. */
const CHARS_PER_TOKEN = 4
const DIGEST_RESERVE = 0.3

function renderGroup(group: MessageGroup): string {
  return group.messages
    .map((message) => {
      const calls =
        message.role === 'assistant' && message.tool_calls
          ? message.tool_calls
              .map((call) => `${call.function?.name ?? 'unknown'}(${call.function?.arguments ?? ''})`)
              .join(', ')
          : ''
      return `${message.role}: ${message.content ?? ''}${calls ? ` [tools: ${calls}]` : ''}`
    })
    .join('\n')
}

const clamp = (text: string, max: number) => (text.length > max ? text.slice(0, max) : text)

/**
 * Renders dropped turns as text for a summarizer, bounded to `maxTokens`.
 * Drops from the oldest end, representing the rest with `digestGroups` so
 * nothing vanishes without trace.
 */
export function renderForSummary(groups: MessageGroup[], maxTokens: number): string {
  const budgetChars = Math.max(0, maxTokens) * CHARS_PER_TOKEN
  if (groups.length === 0) return ''

  const digestBudget = Math.floor(budgetChars * DIGEST_RESERVE)
  const verbatimBudget = Math.max(0, budgetChars - digestBudget - 1)

  const rendered = groups.map(renderGroup)
  const kept: string[] = []
  let chars = 0
  let firstKept = groups.length

  for (let i = groups.length - 1; i >= 0; i--) {
    const cost = rendered[i].length + 1
    if (chars + cost > verbatimBudget && kept.length > 0) break
    kept.unshift(rendered[i])
    chars += cost
    firstKept = i
  }

  const dropped = groups.slice(0, firstKept)
  const body = kept.join('\n')

  if (dropped.length === 0) {
    return body.length > budgetChars ? body.slice(0, budgetChars) : body
  }
  return `${clamp(digestGroups(dropped), digestBudget)}\n${clamp(body, verbatimBudget)}`
}

export interface CompactionResult {
  messages: ChatMessage[]
  compacted: boolean
  tokensBefore: number
  tokensAfter: number
  droppedGroups: number
}

export type Summarizer = (groups: MessageGroup[]) => Promise<string>

/**
 * Compacts a conversation if it exceeds the configured budget. The system
 * prompt is always preserved verbatim.
 */
export async function compactMessages(
  messages: ChatMessage[],
  config: CompactionConfig,
  summarize?: Summarizer,
): Promise<CompactionResult> {
  const tokensBefore = estimateTokens(messages)
  if (tokensBefore <= config.maxInputTokens) {
    return { messages, compacted: false, tokensBefore, tokensAfter: tokensBefore, droppedGroups: 0 }
  }

  const system = messages[0]?.role === 'system' ? [messages[0]] : []
  const rest = messages.slice(system.length)
  const groups = groupMessages(rest)

  const maxDroppable = Math.max(0, groups.length - config.keepRecentGroups)
  const target = config.maxInputTokens * config.targetRatio
  const systemTokens = estimateTokens(system)
  let dropCount = 0
  for (let candidate = 1; candidate <= maxDroppable; candidate++) {
    dropCount = candidate
    const remaining = groups.slice(candidate).flatMap((g) => g.messages)
    if (systemTokens + estimateTokens(remaining) <= target) break
  }

  if (dropCount <= 0) {
    return { messages, compacted: false, tokensBefore, tokensAfter: tokensBefore, droppedGroups: 0 }
  }

  const dropped = groups.slice(0, dropCount)
  const retained = groups.slice(dropCount)

  let summary: string
  try {
    summary = summarize ? await summarize(dropped) : digestGroups(dropped)
  } catch {
    // A failed summarization must not fail the investigation.
    summary = digestGroups(dropped)
  }

  const compactedMessages: ChatMessage[] = [
    ...system,
    { role: 'system', content: summary },
    ...retained.flatMap((g) => g.messages),
  ]

  return {
    messages: compactedMessages,
    compacted: true,
    tokensBefore,
    tokensAfter: estimateTokens(compactedMessages),
    droppedGroups: dropCount,
  }
}
