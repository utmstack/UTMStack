import { useTranslation } from "react-i18next";
import { cn } from "@/shared/lib/utils";
import { fullName, initials } from "../lib/team-utils";
import type { UserBase, UserStatus } from "../types/team.types";

const STATUS_STYLE: Record<UserStatus, { badge: string; dot: string }> = {
  active: {
    badge:
      "bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300",
    dot: "bg-emerald-500",
  },
  pending: {
    badge:
      "bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300",
    dot: "bg-amber-500",
  },
  suspended: {
    badge: "bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300",
    dot: "bg-red-500",
  },
  inactive: {
    badge: "bg-muted text-muted-foreground ring-border",
    dot: "bg-zinc-400",
  },
};

export function StatusBadge({ status }: { status: UserStatus }) {
  const { t } = useTranslation();
  const style = STATUS_STYLE[status] ?? STATUS_STYLE.inactive;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset",
        style.badge,
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full", style.dot)} />
      {t(`team.status.${status}`, { defaultValue: status })}
    </span>
  );
}

export function Avatar({
  user: u,
  size = 36,
}: {
  user: UserBase;
  size?: number;
}) {
  if (u.image_url) {
    return (
      <img
        src={u.image_url}
        alt={fullName(u)}
        className="shrink-0 rounded-full object-cover"
        style={{ height: size, width: size }}
      />
    );
  }
  return (
    <span
      className="flex shrink-0 items-center justify-center rounded-full bg-primary text-[11px] font-semibold text-primary-foreground"
      style={{ height: size, width: size }}
    >
      {initials(u)}
    </span>
  );
}
