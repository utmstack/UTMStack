import type { Flow } from '../types/soar.types'

/**
 * Starting a flow from one that already exists.
 *
 * This replaces the v11 catalogue of loose command templates: copying a whole
 * flow carries its conditions, shell and platform too, not just the command —
 * and it works for the tenant's own flows, which a shipped catalogue never
 * could.
 *
 * Nothing is written when a source is picked: the copy is loaded into the
 * editor and only reaches the backend on save.
 */
export function copyOfFlow(f: Flow): Flow {
  return {
    ...f,
    // Identity is the file, and a copy is a new file that has to be named.
    relPath: '',
    name: '',
    // Provenance stays with the original: the store decides a flow is the
    // tenant's from where the file lands, never from what was sent.
    systemOwner: false,
    // A copy arrives off. Its conditions came from another flow, and firing
    // commands at agents on somebody else's trigger is not a safe default.
    active: false,
  }
}
