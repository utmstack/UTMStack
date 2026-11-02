import type { Flow } from '../types/soar.types'

export type ListTab = "all" | "active" | "inactive" | "system" | "user";
export const LIST_TABS: ListTab[] = ["all", "active", "inactive", "system", "user"];

export function flowName(relPath: string): string {
  return (relPath.split("/").pop() ?? relPath).replace(/\.ya?ml$/i, "");
}

// Per-flow runtime telemetry, aggregated client-side from the most recent
// executions (flows themselves carry no run history). Keyed by rulePath = relPath.
export interface FlowStat {
  last?: string;
  runs: number;
  failed: number;
}

export function relativeTime(iso?: string): string {
  if (!iso) return "—";
  const diff = Date.now() - new Date(iso).getTime();
  if (Number.isNaN(diff)) return "—";
  const m = Math.round(diff / 60_000);
  if (m < 1) return "now";
  if (m < 60) return `${m}m`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h}h`;
  return `${Math.round(h / 24)}d`;
}

// The distinct platforms across the flow's shell nodes — a DAG can span
// several platforms now, so this is the set rather than a flow-level field
// that no longer exists.
export function platformsOf(f: Flow): string[] {
  return Array.from(
    new Set(
      Object.values(f.nodes ?? {})
        .map((n) => n.platform?.trim())
        .filter((p): p is string => !!p),
    ),
  );
}
