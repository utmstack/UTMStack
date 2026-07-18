/**
 * Shared types for the parsing-filter/correlation-rule "playground" — a
 * one-shot test of a raw log against either the live (deployed) pipeline
 * or an unsaved draft, without persisting anything.
 */
export type PlaygroundMode = 'filter' | 'rule'

export interface PlaygroundLog {
  dataType: string
  raw: string
}

export interface PlaygroundFilterInput {
  content: string
}

export interface TestFilterRequest {
  log: PlaygroundLog
  /** Present when testing an unsaved draft; omitted to test the live pipeline. */
  filter?: PlaygroundFilterInput
}

export interface TestFilterResponse {
  uuid: string
  /** The parsed event, when the pipeline produced one. */
  event?: Record<string, unknown>
  alerts: unknown[]
  stopReason?: string
  timedOut: boolean
}

export interface PlaygroundRuleInput {
  content: string
}

export interface TestRuleRequest {
  log: PlaygroundLog
  /** Present when testing an unsaved draft; omitted to test all deployed rules. */
  rule?: PlaygroundRuleInput
}

/** Backend returns the same `PlaygroundResponse` DTO for both test-filter and
 * test-rule — `TestRule` just also populates `alerts`. */
export type TestRuleResponse = TestFilterResponse
