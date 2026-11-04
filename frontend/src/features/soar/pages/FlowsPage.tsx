import { useCallback, useEffect, useMemo, useState, useRef } from "react";
import { useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useListScrollReset } from '@/shared/hooks/use-list-scroll-reset'
import {
  AlertTriangle,
  Loader2,
  Plus,
  RefreshCw,
  Search,
} from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/components/ui/button";
import { Input } from "@/shared/components/ui/input";
import { InfiniteScrollSentinel } from "@/shared/components/ui/infinite-scroll";
import { ResizableDataTable } from "@/shared/components/ui/resizable-data-table";
import { soarFlowsService } from "../services/soar-flows.service";
import { soarExecutionsService } from "../services/soar-executions.service";
import { FlowEditor } from "../components/FlowEditor";
import { SoarCreateDialog } from "../components/SoarCreateDialog";
import { useSocAi } from "@/features/soc-ai/SocAiProvider";
import { StartFromModal } from "@/shared/components/StartFromModal";
import { FlowTriggerModal } from "../components/FlowTriggerModal";
import { copyOfFlow } from "../lib/duplicate";
import { LIST_TABS, type FlowStat, type ListTab } from "../lib/flows-page-utils";
import { FlowKpis } from "../components/FlowKpis";
import { buildFlowColumns } from "../components/FlowColumns";
import type { Flow } from "../types/soar.types";

function Center({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
      {children}
    </div>
  );
}

export function FlowsPage() {
  const { t } = useTranslation();
  const location = useLocation();
  const [listTab, setListTab] = useState<ListTab>("all");
  const [search, setSearch] = useState("");
  const [debounced, setDebounced] = useState("");
  const [items, setItems] = useState<Flow[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [pageSize] = useState(50);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [editing, setEditing] = useState<{
    flow?: Flow;
    creating: boolean;
    initialName?: string;
    initialDescription?: string;
    initialMaxDepth?: number;
  } | null>(null);
  const [starting, setStarting] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [triggering, setTriggering] = useState<Flow | null>(null);
  const { soarEditVersion } = useSocAi();

  const scrollRef = useRef<HTMLDivElement>(null)
  useListScrollReset(page, 0, scrollRef)

  useEffect(() => {
    const relPath = (location.state as { selectFlowId?: string } | null)
      ?.selectFlowId;
    if (!relPath || editing) return;
    soarFlowsService
      .get(relPath)
      .then((f) => setEditing({ flow: f, creating: false }))
      .catch(() => toast.error(t("soar.loadError")));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.state]);
  // Every flow, not the page being browsed: what you can copy must not depend
  // on where you had scrolled to.
  const [copyable, setCopyable] = useState<Flow[]>([]);
  // Bumped whenever a flow's state changes, so the counters refetch. The total
  // cannot serve as that signal: switching one on does not change how many
  // flows exist.
  const [stateVersion, setStateVersion] = useState(0);
  const [stats, setStats] = useState<Record<string, FlowStat>>({});
  const openStartFrom = () => {
    setStarting(true);
    soarFlowsService
      .list({ page: 0, size: 500 })
      .then((r) => setCopyable(r.data ?? []))
      .catch(() => setCopyable([]));
  };

  const startFromFlow = (relPath: string) => {
    const src = copyable.find((f) => f.relPath === relPath);
    setStarting(false);
    if (src) setEditing({ flow: copyOfFlow(src), creating: true });
  };

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim());
      setPage(0);
    }, 300);
    return () => clearTimeout(h);
  }, [search]);

  const query = useMemo(() => {
    const q: {
      name?: string;
      active?: boolean;
      systemOwner?: boolean;
      page: number;
      size: number;
    } = {
      name: debounced || undefined,
      page,
      size: pageSize,
    };
    if (listTab === "active") q.active = true;
    else if (listTab === "inactive") q.active = false;
    else if (listTab === "system") q.systemOwner = true;
    else if (listTab === "user") q.systemOwner = false;
    return q;
  }, [debounced, listTab, page, pageSize]);

  const load = useCallback(() => {
    setLoading(true);
    setError(false);
    soarFlowsService
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
  // Refresh resets to page 0 first so a scrolled list replaces, not appends.
  const handleRefresh = useCallback(() => {
    if (page === 0) load();
    else setPage(0);
  }, [page, load]);
  useEffect(() => {
    load();
  }, [load]);

  // A soc-ai run that touched flows (create or edit) settled — refetch the list.
  useEffect(() => {
    if (soarEditVersion > 0) load();
  }, [soarEditVersion, load]);

  // Aggregate recent executions into per-flow stats (last run / runs / failures).
  // Capped at the most recent 500 executions; refreshed alongside the flows list.
  useEffect(() => {
    let cancelled = false;
    soarExecutionsService
      .list({ page: 0, size: 500 })
      .then((r) => {
        if (cancelled) return;
        const map: Record<string, FlowStat> = {};
        for (const e of r.data ?? []) {
          if (!e.rulePath) continue; // a manual run belongs to no flow
          const s = (map[e.rulePath] ??= { runs: 0, failed: 0 });
          s.runs++;
          if (e.status === "FAILED") s.failed++;
          if (!s.last || (e.startedAt && e.startedAt > s.last))
            s.last = e.startedAt;
        }
        setStats(map);
      })
      .catch(() => {
        if (!cancelled) setStats({});
      });
    return () => {
      cancelled = true;
    };
  }, [load]);

  const toggleActive = async (f: Flow) => {
    const next = !f.active;
    setItems((list) =>
      list.map((x) => (x.relPath === f.relPath ? { ...x, active: next } : x)),
    );
    try {
      await soarFlowsService.setEnabled(f.relPath, next);
      setStateVersion((v) => v + 1);
    } catch {
      setItems((list) =>
        list.map((x) =>
          x.relPath === f.relPath ? { ...x, active: f.active } : x,
        ),
      );
      toast.error(t("soar.toast.activateError"));
    }
  };

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      {!editing && (
        <>
          <FlowKpis refreshKey={total + stateVersion} />
          <div className="flex shrink-0 flex-wrap items-center gap-2">
            <div className="relative">
              <Search
                size={14}
                className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={t("soar.search")}
                className="w-[260px] pl-8"
              />
            </div>
            <div className="inline-flex rounded-md border border-border p-0.5">
              {LIST_TABS.map((tb) => (
                <button
                  key={tb}
                  onClick={() => {
                    setListTab(tb);
                    setPage(0);
                  }}
                  className={cn(
                    "rounded px-2.5 py-1 text-xs transition-colors",
                    listTab === tb
                      ? "bg-muted font-medium text-foreground"
                      : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {t(`soar.tabs.${tb}`)}
                </button>
              ))}
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleRefresh}
              disabled={loading}
              title={t("soar.refresh")}
            >
              <RefreshCw size={14} className={cn(loading && "animate-spin")} />
            </Button>
            <div className="ml-auto">
              <Button size="sm" onClick={openStartFrom}>
                <Plus size={14} className="mr-1.5" /> {t("soar.new")}
              </Button>
            </div>
          </div>

          <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
            <div className="min-h-0 flex-1 overflow-auto" ref={scrollRef}>
              <ResizableDataTable
                columns={buildFlowColumns(t, stats, toggleActive, setTriggering)}
                data={items}
                flexColumnId="flow"
                storageKey="soar-flows-table-sizing"
                getRowId={(f) => f.relPath}
                onRowClick={(f) => setEditing({ flow: f, creating: false })}
                loading={loading && items.length === 0}
                loadingContent={
                  <Center>
                    <Loader2 className="h-4 w-4 animate-spin" />{" "}
                    {t("soar.loading")}
                  </Center>
                }
                error={error}
                errorContent={
                  <Center>
                    <AlertTriangle size={16} className="text-amber-500" />{" "}
                    {t("soar.loadError")}
                    <Button
                      variant="outline"
                      size="sm"
                      className="ml-2"
                      onClick={load}
                    >
                      {t("soar.retry")}
                    </Button>
                  </Center>
                }
                emptyContent={
                  <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                    {t("soar.empty")}
                  </div>
                }
              />
              {items.length > 0 && (
                <InfiniteScrollSentinel
                  onReach={() => setPage((p) => p + 1)}
                  hasMore={items.length < total}
                  loading={loading}
                  endLabel={t("common.allLoaded", { count: total })}
                />
              )}
            </div>
          </div>
        </>
      )}

      {editing && (
        <FlowEditor
          flow={editing.flow}
          creating={editing.creating}
          initialName={editing.initialName}
          initialDescription={editing.initialDescription}
          initialMaxDepth={editing.initialMaxDepth}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            load();
          }}
        />
      )}

      {starting && (
        <StartFromModal
          title={t("soar.new")}
          options={copyable.map((f) => ({ id: f.relPath, name: f.name }))}
          onScratch={() => {
            setStarting(false);
            setCreateOpen(true);
          }}
          onCopy={startFromFlow}
          onClose={() => setStarting(false)}
        />
      )}

      {triggering && (
        <FlowTriggerModal
          key={triggering.relPath}
          flow={triggering}
          onClose={() => setTriggering(null)}
        />
      )}

      <SoarCreateDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onManual={(name, description, maxDepth) => {
          setCreateOpen(false);
          setEditing({
            creating: true,
            initialName: name,
            initialDescription: description,
            initialMaxDepth: maxDepth,
          });
        }}
      />
    </div>
  );
}
