import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { cn } from "@/shared/lib/utils";
import { soarFlowsService } from "../services/soar-flows.service";
import { soarExecutionsService } from "../services/soar-executions.service";

// At-a-glance runtime KPIs. Flows carry no execution telemetry, so these come from
// cheap count queries (size:1, read total) against the flows + executions endpoints.
export function FlowKpis({ refreshKey }: { refreshKey: number }) {
  const { t } = useTranslation();
  const [k, setK] = useState<{
    flows: number;
    active: number;
    runs24h: number;
    failed24h: number;
  } | null>(null);

  useEffect(() => {
    let cancelled = false;
    const since = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    Promise.allSettled([
      soarFlowsService.list({ page: 0, size: 1 }),
      soarFlowsService.list({ page: 0, size: 1, active: true }),
      soarExecutionsService.list({ page: 0, size: 1, startedAtFrom: since }),
      soarExecutionsService.list({
        page: 0,
        size: 1,
        status: "FAILED",
        startedAtFrom: since,
      }),
    ]).then(([f, a, r, x]) => {
      if (cancelled) return;
      const tot = (res: PromiseSettledResult<{ total: number }>) =>
        res.status === "fulfilled" ? res.value.total : 0;
      setK({
        flows: tot(f),
        active: tot(a),
        runs24h: tot(r),
        failed24h: tot(x),
      });
    });
    return () => {
      cancelled = true;
    };
  }, [refreshKey]);

  const cards = [
    { key: "flows", n: k?.flows, tone: "text-foreground" },
    {
      key: "active",
      n: k?.active,
      tone: "text-emerald-600 dark:text-emerald-400",
    },
    { key: "runs24h", n: k?.runs24h, tone: "text-foreground" },
    {
      key: "failed24h",
      n: k?.failed24h,
      tone:
        (k?.failed24h ?? 0) > 0
          ? "text-red-600 dark:text-red-400"
          : "text-foreground",
    },
  ];
  return (
    <div className="mb-3 grid shrink-0 grid-cols-2 gap-3 sm:grid-cols-4">
      {cards.map((c) => (
        <div
          key={c.key}
          className="rounded-lg border border-border bg-card px-4 py-3"
        >
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {t(`soar.kpi.${c.key}`)}
          </div>
          <div
            className={cn("mt-1 text-2xl font-semibold tabular-nums", c.tone)}
          >
            {c.n == null ? "—" : c.n.toLocaleString()}
          </div>
        </div>
      ))}
    </div>
  );
}
