import { cn } from "@/shared/lib/utils";
import { roleDesc, roleLabel } from "../lib/team-utils";
import type { Role } from "../types/team.types";
import { useTranslation } from "react-i18next";
import { Check, Crown, Shield } from "lucide-react";

export function RolePicker({
  roles,
  selected,
  onToggle,
  compact,
}: {
  roles: Role[];
  selected: string[];
  onToggle: (name: string) => void;
  compact?: boolean;
}) {
  const { t } = useTranslation();
  if (roles.length === 0) {
    return (
      <p className="text-[11px] text-muted-foreground">
        {t("team.rolePicker.none")}
      </p>
    );
  }
  return (
    <div className="space-y-2">
      {roles.map((r) => {
        const checked = selected.includes(r.name);
        const isAdmin = r.name === "ROLE_ADMIN";
        return (
          <button
            key={r.name}
            type="button"
            onClick={() => onToggle(r.name)}
            className={cn(
              "flex w-full items-start gap-3 rounded-md border p-2.5 text-left transition-colors",
              checked
                ? "border-primary/40 bg-primary/5"
                : "border-border hover:bg-muted/40",
            )}
          >
            <span
              className={cn(
                "mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded border",
                checked
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-input",
              )}
            >
              {checked && <Check size={11} strokeWidth={3} />}
            </span>
            <span className="min-w-0">
              <span className="flex items-center gap-1.5 text-sm font-medium">
                {isAdmin ? (
                  <Crown size={12} className="text-amber-500" />
                ) : (
                  <Shield size={12} className="text-sky-500" />
                )}
                {roleLabel(t, r.name, r.display_name)}
                {!compact && (
                  <code className="font-mono text-[10px] text-muted-foreground">
                    {r.name}
                  </code>
                )}
              </span>
              {roleDesc(t, r.name, r.description) && (
                <span className="text-[11px] text-muted-foreground">
                  {roleDesc(t, r.name, r.description)}
                </span>
              )}
            </span>
          </button>
        );
      })}
    </div>
  );
}
