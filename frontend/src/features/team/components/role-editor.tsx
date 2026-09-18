import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Check, Loader2, Shield, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/shared/components/ui/button";
import { Input } from "@/shared/components/ui/input";
import { cn } from "@/shared/lib/utils";
import { permResource, resourceLabel, roleError } from "../lib/team-utils";
import { rolesHttpService } from "../services/team-http.service";
import type { Permission, RoleDetail } from "../types/team.types";

export function RoleEditor({
  role,
  onClose,
  onSaved,
}: {
  role: RoleDetail | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useTranslation();
  const [catalog, setCatalog] = useState<Permission[] | null>(null);
  const [name, setName] = useState(role?.name ?? "ROLE_");
  const [displayName, setDisplayName] = useState(role?.display_name ?? "");
  const [description, setDescription] = useState(role?.description ?? "");
  const [selected, setSelected] = useState<string[]>(
    (role?.permissions ?? []).map((p) => p.name),
  );
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let cancelled = false;
    rolesHttpService
      .listPermissions()
      .then((p) => {
        if (!cancelled) setCatalog(p);
      })
      .catch(() => {
        if (!cancelled) toast.error(t("team.toast.permissionsLoadFailed"));
      });
    return () => {
      cancelled = true;
    };
  }, [t]);

  const byResource = useMemo(() => {
    const map = new Map<string, Permission[]>();
    catalog?.forEach((p) => {
      const resource = permResource(p.name);
      if (!map.has(resource)) map.set(resource, []);
      map.get(resource)!.push(p);
    });
    return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]));
  }, [catalog]);

  const valid = name.trim().length >= 2 && name.trim().length <= 50;

  const toggle = (perm: string) =>
    setSelected((s) =>
      s.includes(perm) ? s.filter((p) => p !== perm) : [...s, perm],
    );

  const toggleResource = (perms: Permission[]) => {
    const names = perms.map((p) => p.name);
    const allOn = names.every((n) => selected.includes(n));
    setSelected((s) =>
      allOn
        ? s.filter((p) => !names.includes(p))
        : [...new Set([...s, ...names])],
    );
  };

  const submit = async () => {
    if (!valid || busy) return;
    setBusy(true);
    try {
      const body = {
        name: name.trim(),
        display_name: displayName.trim() || undefined,
        description: description.trim() || undefined,
        permissions: selected,
      };
      if (role) {
        await rolesHttpService.update(role.id, body);
        toast.success(t("team.toast.roleUpdated"));
      } else {
        await rolesHttpService.create(body);
        toast.success(t("team.toast.roleCreated"));
      }
      onSaved();
    } catch (err) {
      toast.error(roleError(err, t));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex max-h-[85vh] w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <Shield size={18} className="text-sky-500" />
              {role ? t("team.roles.editTitle") : t("team.roles.createTitle")}
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">
              {t("team.roles.editorSubtitle")}
            </p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto px-6 py-5">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <DrawerField
              label={t("team.roles.nameLabel")}
              hint={t("team.roles.nameHint")}
            >
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="font-mono"
                placeholder="ROLE_AUDITOR"
              />
            </DrawerField>
            <DrawerField label={t("team.roles.displayNameLabel")}>
              <Input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="Auditor"
              />
            </DrawerField>
          </div>
          <DrawerField label={t("team.roles.descriptionLabel")}>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </DrawerField>

          <div>
            <div className="mb-2 flex items-center justify-between">
              <label className="text-xs font-medium text-foreground/80">
                {t("team.roles.permissionsLabel")}
              </label>
              <span className="text-[11px] text-muted-foreground">
                {t("team.roles.selectedCount", { count: selected.length })}
              </span>
            </div>
            {catalog === null ? (
              <div className="flex items-center gap-2 py-6 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                {t("team.roles.loadingPermissions")}
              </div>
            ) : (
              <div className="space-y-3">
                {byResource.map(([resource, perms]) => (
                  <div
                    key={resource}
                    className="rounded-md border border-border"
                  >
                    <button
                      type="button"
                      onClick={() => toggleResource(perms)}
                      className="flex w-full items-center justify-between bg-muted/30 px-3 py-1.5 text-left text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/60"
                    >
                      {resourceLabel(t, resource)}
                      <span className="text-[10px] normal-case tracking-normal">
                        {t("team.roles.toggleAll")}
                      </span>
                    </button>
                    <div className="grid grid-cols-1 gap-px sm:grid-cols-2">
                      {perms.map((p) => {
                        const on = selected.includes(p.name);
                        return (
                          <button
                            key={p.name}
                            type="button"
                            onClick={() => toggle(p.name)}
                            className={cn(
                              "flex items-start gap-2 px-3 py-2 text-left text-xs transition-colors",
                              on ? "bg-primary/5" : "hover:bg-muted/40",
                            )}
                          >
                            <span
                              className={cn(
                                "mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded border",
                                on
                                  ? "border-primary bg-primary text-primary-foreground"
                                  : "border-input",
                              )}
                            >
                              {on && <Check size={11} strokeWidth={3} />}
                            </span>
                            <span className="min-w-0">
                              <code className="font-mono text-[11px]">
                                {p.name}
                              </code>
                              {p.description && (
                                <span className="block text-[11px] text-muted-foreground">
                                  {p.description}
                                </span>
                              )}
                            </span>
                          </button>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t("team.drawer.cancel")}
          </Button>
          <Button
            size="sm"
            disabled={!valid || busy}
            onClick={() => void submit()}
          >
            {busy ? t("team.drawer.saving") : t("team.roles.save")}
          </Button>
        </footer>
      </div>
    </div>
  );
}

function DrawerField({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">
        {label}
      </label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  );
}
