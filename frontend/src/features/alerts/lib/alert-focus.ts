import type { SocAiFocus } from '@/features/soc-ai'
import type { Alert } from '../types/alert.types'
import { sevKey, statusKey } from './alert-meta'

/**
 * What the assistant is told about the alert a drawer has open. A summary and
 * where to read the rest — the full record is one tool call away, and the
 * assignee (an email) stays out of the prompt.
 */
export function alertFocus(a: Alert): SocAiFocus {
  const facts = [
    `id ${a.id}`,
    a.name && `name "${a.name}"`,
    `severity ${sevKey(a)}`,
    `status ${statusKey(a)}`,
    a.category && `category ${a.category}`,
    a.technique && `technique ${a.technique}`,
    a.dataSource && `data source ${a.dataSource}`,
    a.tags?.length ? `tags ${a.tags.join(', ')}` : '',
  ].filter(Boolean)

  return {
    kind: 'alert',
    id: a.id,
    label: a.name || a.id,
    context:
      `The user has an alert open in the drawer beside this chat (${facts.join('; ')}). ` +
      `When they say "this alert" or "it" they mean this one — do not ask for its id. ` +
      `Read the full record with store.search on the "alerts" dataset filtered by id IS "${a.id}", ` +
      `and call alerts.score with this id for the deterministic score.`,
  }
}
