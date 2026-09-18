import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Check,
  Crown,
  KeyRound,
  Loader2,
  Lock,
  Pencil,
  Plus,
  Shield,
  Trash2,
} from "lucide-react";
import { Button } from "@/shared/components/ui/button";
import {
  actionLabel,
  permResource,
  resourceLabel,
  roleDesc,
  roleLabel,
} from "../lib/team-utils";
import { rolesHttpService } from "../services/team-http.service";
import type { Permission, Role, RoleDetail } from "../types/team.types";
import { RoleEditor } from "./role-editor";
import { DeleteRoleDialog } from "./delete-role-dialog";

export function RolesView({
  roles,
  onChanged,
}: {
  roles: Role[];
  onChanged: () => void;
}) {
  const { t } = useTranslation();
  const [details, setDetails] = useState<RoleDetail[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [editing, setEditing] = useState<RoleDetail | "new" | null>(null);
  const [deleting, setDeleting] = useState<RoleDetail | null>(null);

  useEffect(() => {
    if (roles.length === 0) return;
    let cancelled = false;
    setLoading(true);
    setError(false);
    Promise.all(roles.map((r) => rolesHttpService.get(r.id)))
      .then((d) => {
        if (!cancelled) setDetails(d);
      })
      .catch(() => {
        if (!cancelled) setError(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [roles]);

  // Union of all permissions across roles, grouped by resource (rows of the matrix).
  const byResource = useMemo(() => {
    const map = new Map<string, Map<string, Permission>>();
    details?.forEach((d) =>
      d.permissions.forEach((p) => {
        const resource = permResource(p.name);
        if (!map.has(resource)) map.set(resource, new Map());
        map.get(resource)!.set(p.name, p);
      }),
    );
    return [...map.entries()]
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([resource, perms]) => ({
        resource,
        perms: [...perms.values()].sort((a, b) => a.name.localeCompare(b.name)),
      }));
  }, [details]);

  const has = (roleName: string, permName: string) =>
    details
      ?.find((d) => d.name === roleName)
      ?.permissions.some((p) => p.name === permName) ?? false;

  if (loading) {
    return (
      <div className="mt-10 flex items-center justify-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t("team.roles.loading")}
      </div>
    );
  }
  if (error || !details) {
    return (
      <div className="mt-6 rounded-md border border-amber-500/30 bg-amber-500/5 p-4 text-sm text-muted-foreground">
        {t("team.roles.loadFailed")}
      </div>
    );
  }

  return (
    <div className="mt-5 space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-border bg-muted/30 px-4 py-3">
        <span className="text-xs text-muted-foreground">
          {t("team.roles.systemNote")}
        </span>
        <Button size="sm" onClick={() => setEditing("new")}>
          <Plus size={14} className="mr-1.5" />
          {t("team.roles.newRole")}
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {details.map((d) => (
          <div
            key={d.name}
            className="rounded-xl border border-border bg-card p-5"
          >
            <div className="flex items-start gap-2">
              {d.name === "ROLE_ADMIN" ? (
                <Crown size={18} className="text-amber-500" />
              ) : (
                <Shield size={18} className="text-sky-500" />
              )}
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-sm font-semibold">
                  {roleLabel(t, d.name, d.display_name)}
                  {d.system && (
                    <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-px text-[10px] font-medium text-muted-foreground ring-1 ring-inset ring-border">
                      <Lock size={9} />
                      {t("team.roles.systemBadge")}
                    </span>
                  )}
                </div>
                <code className="font-mono text-[10px] text-muted-foreground">
                  {d.name}
                </code>
              </div>
              {!d.system && (
                <div className="flex shrink-0 items-center gap-1">
                  <button
                    type="button"
                    onClick={() => setEditing(d)}
                    title={t("team.roles.edit")}
                    className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                  >
                    <Pencil size={13} />
                  </button>
                  <button
                    type="button"
                    onClick={() => setDeleting(d)}
                    title={t("team.roles.delete")}
                    className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              )}
            </div>
            {roleDesc(t, d.name, d.description) && (
              <p className="mt-2 text-xs text-muted-foreground">
                {roleDesc(t, d.name, d.description)}
              </p>
            )}
            <div className="mt-3 flex items-center gap-3 text-[11px] text-muted-foreground">
              <span className="inline-flex items-center gap-1">
                <KeyRound size={11} />
                {t("team.roles.permissionsCount", {
                  count: d.permissions.length,
                })}
              </span>
              <span>·</span>
              <span>
                {t("team.roles.modulesCount", {
                  count: new Set(d.permissions.map((p) => permResource(p.name)))
                    .size,
                })}
              </span>
            </div>
          </div>
        ))}
      </div>

      <div className="overflow-hidden rounded-xl border border-border bg-card">
        <div className="border-b border-border px-5 py-3">
          <h3 className="text-sm font-semibold">
            {t("team.roles.matrixTitle")}
          </h3>
          <p className="mt-0.5 text-xs text-muted-foreground">
            {t("team.roles.matrixSubtitle")}
          </p>
        </div>
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-5 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: matrixCols(details.length) }}
        >
          <div>{t("team.roles.permission")}</div>
          {details.map((d) => (
            <div key={d.name} className="text-center">
              {roleLabel(t, d.name, d.display_name)}
            </div>
          ))}
        </div>
        {byResource.map(({ resource, perms }) => (
          <div key={resource}>
            <div className="bg-muted/20 px-5 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
              {resourceLabel(t, resource)}
            </div>
            {perms.map((p) => (
              <div
                key={p.name}
                className="grid items-center gap-3 border-b border-border/50 px-5 py-2 text-xs last:border-b-0"
                style={{ gridTemplateColumns: matrixCols(details.length) }}
              >
                <div>
                  <code className="font-mono text-[11px]">{p.name}</code>
                  <div className="text-[11px] text-muted-foreground">
                    {p.description ||
                      actionLabel(t, p.name.split(".")[1] ?? "")}
                  </div>
                </div>
                {details.map((d) => (
                  <PermTick key={d.name} on={has(d.name, p.name)} />
                ))}
              </div>
            ))}
          </div>
        ))}
      </div>

      {editing && (
        <RoleEditor
          role={editing === "new" ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            onChanged();
          }}
        />
      )}
      {deleting && (
        <DeleteRoleDialog
          role={deleting}
          onClose={() => setDeleting(null)}
          onDeleted={() => {
            setDeleting(null);
            onChanged();
          }}
        />
      )}
    </div>
  );
}

export function matrixCols(n: number): string {
  return `1fr ${Array.from({ length: n }, () => "110px").join(" ")}`;
}

export function PermTick({ on }: { on: boolean }) {
  return (
    <div className="flex justify-center">
      {on ? (
        <Check size={14} className="text-emerald-500" strokeWidth={3} />
      ) : (
        <span className="text-muted-foreground/40">—</span>
      )}
    </div>
  );
}
