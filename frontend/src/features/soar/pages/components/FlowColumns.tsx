import type { TFunction } from "i18next";
import type { ColumnDef } from "@tanstack/react-table";
import { Lock, Play, Terminal } from "lucide-react";
import { cn } from "@/shared/lib/utils";
import {
  PlatformBroadcastButton,
  broadcast,
  BULK_PATHS,
} from "@/features/platform-broadcast";
import { flowName, platformsOf, relativeTime, type FlowStat } from "../../lib/flows-page-utils";
import type { Flow } from "../../types/soar.types";

const TH = "whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium";
const TD = "whitespace-nowrap px-3 py-2.5 align-middle";
const stopRowClick = (e: React.MouseEvent) => e.stopPropagation();

export function buildFlowColumns(
  t: TFunction,
  stats: Record<string, FlowStat>,
  onToggle: (f: Flow) => void,
  onTrigger: (f: Flow) => void,
): ColumnDef<Flow>[] {
  const count = `${TD} font-mono text-[11px] text-muted-foreground`;
  return [
    {
      id: "flow",
      header: t("soar.cols.flow"),
      size: 360,
      minSize: 120,
      meta: {
        headerClassName: TH,
        cellClassName: TD,
        cellProps: (f) => ({ title: f.relPath }),
      },
      cell: ({ row }) => {
        const f = row.original;
        return (
          <div className="flex min-w-0 items-center gap-2">
            <Terminal size={14} className="shrink-0 text-muted-foreground" />
            <div className="min-w-0">
              <div className="truncate text-[13px]">
                {f.name || flowName(f.relPath)}
              </div>
              {f.description && (
                <div className="truncate text-[11px] text-muted-foreground">
                  {f.description}
                </div>
              )}
            </div>
          </div>
        );
      },
    },
    {
      id: "platform",
      header: t("soar.cols.platform"),
      size: 130,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const platforms = platformsOf(row.original);
        if (platforms.length === 0)
          return <span className="text-[11px] text-muted-foreground">—</span>;
        return (
          <div className="flex flex-wrap gap-1">
            {platforms.map((p) => (
              <span
                key={p}
                className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]"
              >
                {p}
              </span>
            ))}
          </div>
        );
      },
    },
    {
      id: "conditions",
      header: t("soar.cols.conditions"),
      size: 100,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: count },
      cell: ({ row }) => row.original.conditions?.length ?? 0,
    },
    {
      id: "commands",
      header: t("soar.cols.commands"),
      size: 100,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: count },
      cell: ({ row }) =>
        row.original.nodes ? Object.keys(row.original.nodes).length : 0,
    },
    {
      id: "lastRun",
      header: t("soar.cols.lastRun"),
      size: 140,
      minSize: 90,
      meta: {
        headerClassName: TH,
        cellClassName: `${TD} text-[11px]`,
        cellProps: (f) => {
          const last = stats[f.relPath]?.last;
          return { title: last ? new Date(last).toLocaleString() : undefined };
        },
      },
      cell: ({ row }) => {
        const stat = stats[row.original.relPath];
        if (!stat?.last)
          return (
            <span className="text-muted-foreground/50">
              {t("soar.neverRun")}
            </span>
          );
        return (
          <div className="flex items-center gap-1.5">
            <span className="font-mono text-muted-foreground">
              {relativeTime(stat.last)}
            </span>
            {stat.failed > 0 && (
              <span
                className="rounded bg-red-500/15 px-1 py-0.5 text-[9px] font-semibold text-red-600 dark:text-red-400"
                title={t("soar.cols.failedRuns", { count: stat.failed })}
              >
                {stat.failed} ✕
              </span>
            )}
          </div>
        );
      },
    },
    {
      id: "active",
      header: t("soar.cols.active"),
      size: 100,
      minSize: 90,
      enableResizing: false,
      meta: {
        headerClassName: `${TH} text-center`,
        cellClassName: `${TD} text-center`,
        cellProps: () => ({ onClick: stopRowClick }),
      },
      cell: ({ row }) => (
        <div className="flex items-center justify-center">
          <Toggle
            t={t}
            checked={row.original.active}
            onChange={() => onToggle(row.original)}
            flow={row.original}
          />
        </div>
      ),
    },
    {
      id: "view",
      header: () => null,
      size: 70,
      minSize: 70,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: `${TD} text-right` },
      cell: ({ row }) => (
        <div className="flex items-center justify-end gap-1.5">
          {row.original.systemOwner && (
            <Lock size={11} className="text-muted-foreground/60" />
          )}
          <span className="text-[11px] text-muted-foreground">
            {t("soar.view")}
          </span>
        </div>
      ),
    },
    {
      id: "trigger",
      header: t("soar.cols.actions"),
      size: 40,
      minSize: 40,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onTrigger(row.original);
          }}
          title={t("soar.trigger.btn")}
          aria-label={t("soar.trigger.btn")}
          className="mx-auto flex h-7 w-7 items-center justify-center rounded text-muted-foreground/60 transition hover:bg-background hover:text-primary"
        >
          <Play size={13} />
        </button>
      ),
    },
  ];
}

function Toggle({
  t,
  checked,
  onChange,
  flow,
}: {
  t: TFunction;
  checked: boolean;
  onChange: () => void;
  flow: Flow;
}) {
  return (
    <div
      className="flex items-center gap-1.5"
      onClick={(e) => e.stopPropagation()}
    >
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        // The row opens the editor on click. Without stopping here, switching a
        // flow on also opens it — two things from one press, and the panel lands
        // on top of the row you were aiming at.
        onClick={(e) => {
          e.stopPropagation();
          onChange();
        }}
        className={cn(
          "relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors",
          checked ? "bg-primary" : "bg-muted-foreground/30",
        )}
      >
        <span
          className={cn(
            "inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform",
            checked ? "translate-x-4" : "translate-x-0.5",
          )}
        />
      </button>
      <PlatformBroadcastButton
        label={t("platformBroadcast.button")}
        title={t("platformBroadcast.action.enable", {
          resource: t("platformBroadcast.resource.soarFlow"),
        })}
        onBroadcast={(selector) =>
          broadcast(BULK_PATHS.soarRules.enable, selector, {
            relPath: flow.relPath,
            enabled: !checked,
          })
        }
        size="sm"
      />
    </div>
  );
}
