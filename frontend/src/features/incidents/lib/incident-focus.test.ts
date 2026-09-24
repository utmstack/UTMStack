import { describe, expect, it } from 'vitest'
import type { Incident } from '../types/incident.types'
import { incidentFocus } from './incident-focus'

const incident: Incident = {
  id: 'inc-1',
  incidentName: 'Suspicious login burst',
  incidentStatus: 'In review',
  incidentSeverity: 'high',
  incidentAssignedTo: 'analyst@example.com',
  incidentCreatedDate: '2026-09-23T15:42:08Z',
  alertCount: 3,
}

describe('incidentFocus', () => {
  const focus = incidentFocus(incident)

  it('names the incident and what the chip should show', () => {
    expect(focus.kind).toBe('incident')
    expect(focus.id).toBe('inc-1')
    expect(focus.label).toBe('Suspicious login burst')
  })

  it('tells the agent what is open and which tools read the rest', () => {
    expect(focus.context).toContain('status In review')
    expect(focus.context).toContain('3 linked alerts')
    expect(focus.context).toContain('incidents.get')
    expect(focus.context).toContain('incident_alerts.list')
    expect(focus.context).toContain('do not ask for its id')
  })

  it('keeps the assignee out of the prompt', () => {
    expect(focus.context).not.toContain('analyst@example.com')
  })

  it('leaves severity out when the incident has none yet', () => {
    expect(incidentFocus({ ...incident, incidentSeverity: undefined }).context).not.toContain('severity')
  })
})
