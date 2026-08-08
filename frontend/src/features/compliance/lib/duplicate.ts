import type { Control, Framework } from '../types/compliance.types'

/**
 * Starting a definition from one that already exists.
 *
 * Building a variant of a shipped framework by hand is not a real option — PCI
 * DSS alone is 155 requirements — and the catalogue is read-only by design.
 * Copying is what makes that rule livable.
 *
 * Nothing is written when a source is picked: the copy is loaded into the
 * editor and only reaches the backend when the user saves it. So a half-made
 * definition never exists as a row somebody has to clean up.
 */

/** Identity and name are cleared: a copy is a new thing and has to be named. */
export function copyOfFramework(f: Framework): Framework {
  return {
    ...f,
    key: '',
    name: '',
    // Provenance stays with the original. The store decides a definition is the
    // tenant's from where the file lands, never from what was sent.
    system: undefined,
    locked: undefined,
    source: f.source ? `Adapted from ${f.source}` : undefined,
  }
}

export function copyOfControl(c: Control): Control {
  return {
    ...c,
    id: '',
    name: '',
    system: undefined,
    locked: undefined,
    source: c.source ? `Adapted from ${c.source}` : undefined,
  }
}
