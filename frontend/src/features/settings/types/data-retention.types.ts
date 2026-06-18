/* Mirrors backend opensearch domain.PolicySettings (ISM index lifecycle). */

export interface PolicySettings {
  /** Snapshot the index (restore-only archive) before deleting it. */
  snapshotActive: boolean
  /** OpenSearch ISM index age after which the index is deleted, e.g. "90d". */
  deleteAfter: string
}
