import type { TFunction } from "i18next";
import type { ColumnDef } from "@tanstack/react-table";
import { KeyRound, Pencil, ShieldCheck, ShieldOff, Trash2 } from "lucide-react";
import { fullName } from "../lib/team-utils";
import type { UserListItem } from "../types/team.types";
import { Avatar, StatusBadge } from "./avatar";
import { RoleBadge } from "./role-badge";

const TH = "whitespace-nowrap px-3 py-2 text-left align-middle font-medium";
const TD = "whitespace-nowrap px-3 py-3 align-middle";

export function buildMemberColumns(t: TFunction, onRemove: (u: UserListItem) => void): ColumnDef<UserListItem>[] {
  return [
    {
      id: "user",
      header: t("team.members.colUser"),
      size: 380,
      minSize: 140,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const u = row.original;
        return (
          <div className="flex min-w-0 items-center gap-3">
            <Avatar user={u} />
            <div className="min-w-0">
              <div className="truncate text-sm font-medium">{fullName(u)}</div>
              <div className="flex min-w-0 items-center gap-1.5 text-[11px] text-muted-foreground">
                {u.name?.trim() && <span className="truncate">{u.email}</span>}
                {u.federated && (
                  <span className="inline-flex shrink-0 items-center gap-1 rounded bg-muted px-1 py-px text-[10px]">
                    <KeyRound size={9} />
                    {t("team.members.federated")}
                  </span>
                )}
              </div>
            </div>
          </div>
        );
      },
    },
    {
      id: "roles",
      header: t("team.members.colRoles"),
      size: 260,
      minSize: 100,
      meta: {
        headerClassName: TH,
        cellClassName: TD,
        cellProps: (u) => ({
          title: (u.roles ?? []).map((r) => r.display_name || r.name).join(", "),
        }),
      },
      cell: ({ row }) => {
        const roles = row.original.roles ?? [];
        return roles.length === 0 ? (
          <span className="text-[11px] italic text-muted-foreground">
            {t("team.members.noRole")}
          </span>
        ) : (
          <div className="flex items-center gap-1">
            {roles.map((r) => (
              <RoleBadge key={r.name} name={r.name} label={r.display_name} />
            ))}
          </div>
        );
      },
    },
    {
      id: "tfa",
      header: t("team.members.col2fa"),
      size: 100,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) =>
        row.original.tfa_enabled ? (
          <span className="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
            <ShieldCheck size={12} /> {t("team.members.on")}
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 text-muted-foreground">
            <ShieldOff size={12} /> {t("team.members.off")}
          </span>
        ),
    },
    {
      id: "status",
      header: t("team.members.colStatus"),
      size: 130,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <StatusBadge status={row.original.status} />,
    },
    {
      id: "edit",
      header: () => null,
      size: 84,
      minSize: 84,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: `${TD} text-muted-foreground` },
      cell: ({ row }) => (
        <div className="flex justify-end items-center gap-1">
          <Pencil size={13} />
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onRemove(row.original);
            }}
            disabled={row.original.status !== "inactive"}
            title={t("team.drawer.remove")}
            className="flex h-7 w-7 items-center justify-center rounded-md hover:bg-red-500/10 hover:text-red-500 disabled:cursor-not-allowed disabled:opacity-40"
          >
            <Trash2 size={13} />
          </button>
        </div>
      ),
    },
  ];
}
