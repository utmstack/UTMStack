import { useTranslation } from "react-i18next";
import { Crown, Shield } from "lucide-react";
import { cn } from "@/shared/lib/utils";
import { roleLabel } from "../lib/team-utils";

export function RoleBadge({ name, label }: { name: string; label: string }) {
  const { t } = useTranslation();
  const isAdmin = name === "ROLE_ADMIN";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset",
        isAdmin
          ? "bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300"
          : "bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300",
      )}
    >
      {isAdmin ? <Crown size={10} /> : <Shield size={10} />}
      {roleLabel(t, name, label)}
    </span>
  );
}
