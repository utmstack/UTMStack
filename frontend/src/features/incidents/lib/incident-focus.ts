import type { SocAiFocus } from '@/features/soc-ai'
import type { Incident } from '../types/incident.types'

/** What the assistant is told about the incident a drawer has open. */
export function incidentFocus(i: Incident): SocAiFocus {
  const facts = [
    `id ${i.id}`,
    `name "${i.incidentName}"`,
    `status ${i.incidentStatus}`,
    i.incidentSeverity && `severity ${i.incidentSeverity}`,
    `${i.alertCount} linked alerts`,
  ].filter(Boolean)

  return {
    kind: 'incident',
    id: i.id,
    label: i.incidentName,
    context:
      `The user has an incident open in the drawer beside this chat (${facts.join('; ')}). ` +
      `When they say "this incident" or "it" they mean this one — do not ask for its id. ` +
      `Read it with incidents.get, its alerts with incident_alerts.list, ` +
      `and its notes and history with incident_notes.list and incident_history.list, all with this incident id.`,
  }
}
