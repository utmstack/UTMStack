import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  Loader2,
  RefreshCw,
  Search,
  XCircle,
} from "lucide-react";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/components/ui/button";
import { Input } from "@/shared/components/ui/input";
import { InfiniteScrollSentinel } from "@/shared/components/ui/infinite-scroll";
import {
  presetRange,
  resolveRange,
  TimeRangePicker,
  type TimeRange,
} from "@/shared/components/ui/time-range-picker";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/shared/components/ui/tooltip";
import { useDateFormat } from "@/shared/lib/datetime";
import { datasourcesHttpService } from "@/features/datasources/services/datasources-http.service";
import { soarExecutionsService } from "../services/soar-executions.service";
import { soarFlowsService } from "../services/soar-flows.service";
import type {
  Execution,
  ExecutionOrigin,
  ExecutionStatus,
  ExecutionListQuery,
  Flow,
  FlowNode,
} from "../types/soar.types";

const STATUSES: (ExecutionStatus | "all")[] = [
  "all",
  "EXECUTED",
  "PENDING",
  "WAITING",
  "EXECUTING",
  "FAILED",
  "DEAD",
];
const ORIGINS: (ExecutionOrigin | "all")[] = ["all", "FLOW", "MANUAL"];
const COLS =
  "90px 100px minmax(160px,1.2fr) minmax(180px,1.6fr) 120px 150px 60px";

const STATUS_META: Record<
  ExecutionStatus,
  { icon: typeof CheckCircle2; cls: string }
> = {
  EXECUTED: { icon: CheckCircle2, cls: "text-emerald-500" },
  PENDING: { icon: Clock, cls: "text-amber-500" },
  WAITING: { icon: Clock, cls: "text-muted-foreground" },
  EXECUTING: { icon: Loader2, cls: "text-sky-500 [&_svg]:animate-spin" },
  FAILED: { icon: XCircle, cls: "text-red-500" },
  DEAD: { icon: AlertTriangle, cls: "text-muted-foreground" },
};

export function ExecutionsView() {
  const { t } = useTranslation();
  const df = useDateFormat();
  const [search, setSearch] = useState("");
  const [debounced, setDebounced] = useState("");
  const [status, setStatus] = useState<ExecutionStatus | "all">("all");
  const [origin, setOrigin] = useState<ExecutionOrigin | "all">("FLOW");
  const [agent, setAgent] = useState<string>("");
  const [agents, setAgents] = useState<string[]>([]);
  const [range, setRange] = useState<TimeRange>(presetRange("7d"));
  const [items, setItems] = useState<Execution[]>([]);
  const [total, setTotal] = useState(0);
  // Flows of the runs currently on screen, keyed by rulePath. Used to render
  // each node's position in the flow's DAG (its ancestor chain) in the Node
  // column — the flow itself carries no per-run state, only its shape.
  const [runFlows, setRunFlows] = useState<Record<string, Flow>>({});
  const [page, setPage] = useState(0);
  const [pageSize] = useState(50);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim());
      setPage(0);
    }, 300);
    return () => clearTimeout(h);
  }, [search]);

  // Same source as FlowEditor / InteractiveConsole — Execution.agent stores the datasource name.
  useEffect(() => {
    datasourcesHttpService
      .list({ page: 1, size: 1000, kind: "agent", sort: "asset_name.asc" })
      .then((r) =>
        setAgents((r.items ?? []).map((d) => d.name).filter(Boolean)),
      )
      .catch(() => {});
  }, []);

  const query = useMemo<ExecutionListQuery>(() => {
    const { from, to } = resolveRange(range);
    return {
      alertId: debounced || undefined,
      status: status === "all" ? undefined : status,
      origin: origin === "all" ? undefined : origin,
      agent: agent || undefined,
      startedAtFrom: from ?? undefined,
      startedAtTo: to,
      page,
      size: pageSize,
    };
  }, [debounced, status, origin, agent, range, page, pageSize]);

  const load = useCallback(() => {
    setLoading(true);
    setError(false);
    soarExecutionsService
      .list(query)
      .then((r) => {
        setItems((prev) =>
          page === 0 ? (r.data ?? []) : [...prev, ...(r.data ?? [])],
        );
        setTotal(r.total ?? 0);
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, [query, page]);
  useEffect(() => {
    load();
  }, [load]);

  // Fetch the flows behind the runs currently on screen so the Node column can
  // show where each node sits in the DAG. One GET per unseen rulePath; a failure
  // leaves the cell showing the bare node id (the flow may have been deleted).
  const neededPaths = useMemo(() => {
    const seen = new Set<string>();
    for (const e of items) {
      if (
        e.origin === "FLOW" &&
        e.rulePath &&
        !runFlows[e.rulePath] &&
        !seen.has(e.rulePath)
      )
        seen.add(e.rulePath);
    }
    return [...seen];
  }, [items, runFlows]);
  useEffect(() => {
    if (neededPaths.length === 0) return;
    let cancelled = false;
    Promise.allSettled(
      neededPaths.map(async (p) => {
        const f = await soarFlowsService.get(p);
        return { p, f };
      }),
    ).then((results) => {
      if (cancelled) return;
      const found: Record<string, Flow> = {};
      for (const res of results) {
        if (res.status === "fulfilled") found[res.value.p] = res.value.f;
      }
      if (Object.keys(found).length > 0)
        setRunFlows((prev) => ({ ...prev, ...found }));
    });
    return () => {
      cancelled = true;
    };
  }, [neededPaths]);

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="mb-3 flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search
            size={14}
            className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground"
          />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("soar.executions.search")}
            className="w-[260px] pl-8"
          />
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {STATUSES.map((s) => (
            <button
              key={s}
              onClick={() => {
                setStatus(s);
                setPage(0);
              }}
              className={cn(
                "rounded px-2.5 py-1 text-xs transition-colors",
                status === s
                  ? "bg-muted font-medium text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {s === "all"
                ? t("soar.executions.all")
                : t(`soar.executionStatus.${s}`)}
            </button>
          ))}
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {ORIGINS.map((o) => (
            <button
              key={o}
              onClick={() => {
                setOrigin(o);
                setPage(0);
              }}
              className={cn(
                "rounded px-2.5 py-1 text-xs transition-colors",
                origin === o
                  ? "bg-muted font-medium text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {o === "all"
                ? t("soar.executions.all")
                : t(`soar.executionOrigin.${o}`)}
            </button>
          ))}
        </div>
        <select
          value={agent}
          onChange={(e) => {
            setAgent(e.target.value);
            setPage(0);
          }}
          className="h-9 cursor-pointer rounded-md border border-input bg-popover px-2 text-xs text-foreground"
          title={t("soar.executions.filters.source")}
        >
          <option value="">{t("soar.executions.filters.allSources")}</option>
          {agents.map((a) => (
            <option key={a} value={a}>
              {a}
            </option>
          ))}
        </select>
        <TimeRangePicker
          value={range}
          onChange={(r) => {
            setRange(r);
            setPage(0);
          }}
          align="right"
        />
        <Button
          variant="outline"
          size="sm"
          onClick={load}
          disabled={loading}
          title={t("soar.refresh")}
        >
          <RefreshCw size={14} className={cn(loading && "animate-spin")} />
        </Button>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: COLS }}
        >
          <div>{t("soar.executions.cols.status")}</div>
          <div>{t("soar.executions.cols.node")}</div>
          <div>{t("soar.executions.cols.flow")}</div>
          <div>{t("soar.executions.cols.command")}</div>
          <div>{t("soar.executions.cols.agent")}</div>
          <div>{t("soar.executions.cols.date")}</div>
          <div className="text-center">{t("soar.executions.cols.retries")}</div>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center>
              <Loader2 className="h-4 w-4 animate-spin" />{" "}
              {t("soar.executions.loading")}
            </Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" />{" "}
              {t("soar.executions.loadError")}
              <Button
                variant="outline"
                size="sm"
                className="ml-2"
                onClick={load}
              >
                {t("soar.executions.retry")}
              </Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">
              {t("soar.executions.empty")}
            </div>
          ) : (
            <>
              {items.map((e) => (
                <ExecutionRow
                  key={e.id}
                  e={e}
                  flow={e.rulePath ? runFlows[e.rulePath] : undefined}
                  df={df}
                  t={t}
                />
              ))}
              <InfiniteScrollSentinel
                onReach={() => setPage((p) => p + 1)}
                hasMore={items.length < total}
                loading={loading}
                endLabel={t("common.allLoaded", { count: total })}
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function ExecutionRow({
  e,
  flow,
  df,
  t,
}: {
  e: Execution;
  flow?: Flow;
  df: ReturnType<typeof useDateFormat>;
  t: ReturnType<typeof useTranslation>["t"];
}) {
  const meta = STATUS_META[e.status];
  const Icon = meta?.icon ?? Clock;
  // A manual run has no flow: what identifies it is who typed it.
  const source =
    e.origin === "MANUAL"
      ? e.triggeredBy || t("soar.executions.manual")
      : ((e.rulePath ?? "").split("/").pop() ?? "").replace(/\.ya?ml$/i, "") ||
        "—";

  // Node column: the flow carries no per-run state, only its DAG shape — so the
  // node's place in the run is its ancestor chain, read off the live flow.
  const nodeLabel =
    e.origin === "FLOW" && e.nodeId
      ? [ancestorPath(e.nodeId, flow?.nodes ?? {}), e.nodeId]
          .filter(Boolean)
          .join(" ← ")
      : e.origin === "MANUAL"
        ? t("soar.executions.manual")
        : "—";

  return (
    <div
      className="grid items-center gap-3 border-b border-border px-4 py-2.5 text-sm last:border-0"
      style={{ gridTemplateColumns: COLS }}
    >
      <div
        className={cn(
          "inline-flex items-center gap-1.5 text-[11px] font-medium",
          meta?.cls,
        )}
      >
        <Icon size={13} /> {t(`soar.executionStatus.${e.status}`)}
      </div>
      <div
        className="min-w-0 truncate text-[11px] text-muted-foreground"
        title={nodeLabel}
      >
        {nodeLabel}
      </div>
      <div className="min-w-0">
        <div
          className="truncate text-[13px]"
          title={e.rulePath ?? e.triggeredBy}
        >
          {source}
        </div>
        {e.alertId && (
          <div
            className="truncate font-mono text-[10px] text-muted-foreground"
            title={e.alertId}
          >
            {e.alertId}
          </div>
        )}
      </div>
      <div className="min-w-0">
        <CommandCell text={e.command} />
        {e.nonExecutionCause && (
          <div className="text-[10px] text-red-500">
            {t(`soar.nonExecutionCause.${e.nonExecutionCause}`)}
          </div>
        )}
      </div>
      <div
        className="truncate font-mono text-[11px] text-muted-foreground"
        title={e.agent}
      >
        {e.agent || "—"}
      </div>
      <div className="text-[11px] text-muted-foreground">
        {df.formatDateTime(e.startedAt)}
      </div>
      <div className="text-center text-[11px] text-muted-foreground">
        {e.retries || 0}
      </div>
    </div>
  );
}

// The command column is always filled for every node: flow executors that
// carry no shell command (http, mail, llm, notify, incident, conditional) get
// a derived action summary from the backend. Long values clamp to one line;
// the tooltip carries the full text.
function CommandCell({ text }: { text?: string }) {
  const value = (text ?? "").trim();
  if (!value) return <span className="text-muted-foreground/50">—</span>;
  return (
    <div className="min-w-0">
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="cursor-default truncate font-mono text-[11px]">
            {value}
          </div>
        </TooltipTrigger>
        <TooltipContent
          side="top"
          className="max-w-[440px] whitespace-pre-wrap break-all"
        >
          {value}
        </TooltipContent>
      </Tooltip>
    </div>
  );
}

// Walk the flow's DAG backward from `startId` and return the ancestor chain
// root → … → parent as ids. Bounded by a visited-set so cycles can't spin.
// Returns '' when the node isn't in this flow (renamed or deleted since the run).
function ancestorPath(
  startId: string,
  nodes: Record<string, FlowNode>,
): string {
  if (!nodes[startId]) return "";
  const chain: string[] = [];
  const seen = new Set<string>([startId]);
  let cur: string | undefined = startId;
  while (cur) {
    const parents = parentIds(cur, nodes);
    if (parents.length === 0) break;
    const next = parents[0]; // arbitrary but stable: first declared parent
    if (seen.has(next)) break;
    seen.add(next);
    chain.unshift(next);
    cur = next;
  }
  return chain.join(" ← ");
}

function parentIds(id: string, nodes: Record<string, FlowNode>): string[] {
  const out: string[] = [];
  for (const [pid, n] of Object.entries(nodes)) {
    if (pid === id) continue;
    if ((n.onSuccess ?? []).includes(id) || (n.onError ?? []).includes(id))
      out.push(pid);
  }
  return out;
}

function Center({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
      {children}
    </div>
  );
}
