import { useTranslation } from "react-i18next";
import { KeyRound, Pencil, ShieldCheck, ShieldOff } from "lucide-react";
import { cn } from "@/shared/lib/utils";
import { fullName } from "../lib/team-utils";
import type { UserListItem } from "../types/team.types";
import { Avatar } from "./avatar";
import { RoleBadge } from "./role-badge";
import { StatusBadge } from "./avatar";

export function MemberRow({
  user: u,
  tableCols,
  onOpen,
}: {
  user: UserListItem;
  tableCols: string;
  onOpen: () => void;
}) {
  const { t } = useTranslation();
  return (
    <div
      onClick={onOpen}
      className={cn(
        "grid w-max min-w-full cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs last:border-b-0 hover:bg-muted/40",
        u.status !== "active" && "opacity-70",
      )}
      style={{ gridTemplateColumns: tableCols }}
    >
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
      <div className="flex flex-wrap gap-1">
        {(u.roles ?? []).length === 0 ? (
          <span className="text-[11px] italic text-muted-foreground">
            {t("team.members.noRole")}
          </span>
        ) : (
          u.roles!.map((r) => (
            <RoleBadge key={r.name} name={r.name} label={r.display_name} />
          ))
        )}
      </div>
      <div>
        {u.tfa_enabled ? (
          <span className="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
            <ShieldCheck size={12} /> {t("team.members.on")}
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 text-muted-foreground">
            <ShieldOff size={12} /> {t("team.members.off")}
          </span>
        )}
      </div>
      <div>
        <StatusBadge status={u.status} />
      </div>
      <div className="flex justify-end text-muted-foreground">
        <Pencil size={13} />
      </div>
    </div>
  );
}
