/**
 * Shared types for the parsing-pipeline/correlation-rule "playground" — a
 * one-shot test of a raw log against either the live (deployed) pipeline
 * or an unsaved draft, without persisting anything.
 */
export type PlaygroundMode = 'pipeline' | 'rule'

export interface PlaygroundLog {
  dataType: string
  raw: string
}

export interface PlaygroundPipelineInput {
  content: string
}

export interface TestPipelineRequest {
  log: PlaygroundLog
  /** Present when testing an unsaved draft; omitted to run the deployed ones. */
  pipeline?: PlaygroundPipelineInput
}

export interface TestPipelineResponse {
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

/** Backend returns the same `PlaygroundResponse` DTO for both test-pipeline and
 * test-rule — `TestRule` just also populates `alerts`. */
export type TestRuleResponse = TestPipelineResponse
