import { describe, expect, test } from 'vitest'
import { parseAiNote, userNotePart, aiBlock, combineUserNote } from '@/features/alerts/lib/ai-note'

// The exact line the soc-ai prompt mandates (plugins/soc-ai/internal/agent/prompt.go).
const NOTE =
  '[AI SOC Agent] Score: 72/100 - Open - High Risk | Threat Assessment: Privileged group changed off-hours. | ' +
  'Affected Asset: WIN-SRV-08 | Context: Third occurrence this week from the same host. | ' +
  'LLM Analysis: Matches a known escalation pattern. | Action: Escalate to tier 2.'

describe('parseAiNote against the format the agent is told to write', () => {
  test('headline, score and risk', () => {
    const n = parseAiNote(NOTE)!
    expect(n.score).toBe(72)
    expect(n.risk).toBe('high')
  })

  test('every labelled section survives', () => {
    const n = parseAiNote(NOTE)!
    expect(n.sections.map((s) => s.label)).toEqual([
      'Threat Assessment', 'Affected Asset', 'Context', 'LLM Analysis', 'Action',
    ])
    expect(n.sections[1].value).toBe('WIN-SRV-08')
  })

  test('a colon inside a value keeps the label intact', () => {
    const n = parseAiNote('[AI SOC Agent] Score: 10/100 - Open - Low Risk | Context: seen at 10:42 UTC')!
    expect(n.sections[0]).toEqual({ label: 'Context', value: 'seen at 10:42 UTC' })
  })

  test('the analyst note and the AI block stay separable', () => {
    const combined = combineUserNote('looks benign', NOTE)
    expect(userNotePart(combined)).toBe('looks benign')
    expect(aiBlock(combined)).toBe(NOTE)
  })

  // The agent writes free text into these sections, so a pipe can land inside
  // one. A real section opens with a label; a continuation does not.
  test('a pipe inside a value stays in that value', () => {
    const n = parseAiNote('[AI SOC Agent] Score: 5/100 - Open - Low Risk | Action: block the IP | then escalate')!
    expect(n.sections).toEqual([{ label: 'Action', value: 'block the IP | then escalate' }])
  })

  test('several pipes in one value all stay put', () => {
    const n = parseAiNote('[AI SOC Agent] Score: 5/100 - Open - Low Risk | Action: a | b | c | Context: after')!
    expect(n.sections).toEqual([
      { label: 'Action', value: 'a | b | c' },
      { label: 'Context', value: 'after' },
    ])
  })
})
