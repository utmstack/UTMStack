import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Loader2, ShieldOff, UserCheck, UserX, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/shared/components/ui/button";
import {
  fullName,
  langLabel,
  relativeTime,
  userError,
} from "../lib/team-utils";
import { usersHttpService } from "../services/team-http.service";
import type { Role, UserDetail } from "../types/team.types";
import { Avatar, StatusBadge } from "./avatar";
import { RolePicker } from "./role-picker";

export function MemberDrawer({
  userId,
  roles,
  onClose,
  onChanged,
}: {
  userId: string;
  roles: Role[];
  onClose: () => void;
  onChanged: () => void;
}) {
  const { t } = useTranslation();
  const [user, setUser] = useState<UserDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [roleNames, setRoleNames] = useState<string[]>([]);
  const [confirmResetTfa, setConfirmResetTfa] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    usersHttpService
      .get(userId)
      .then((u) => {
        if (cancelled) return;
        setUser(u);
        setRoleNames((u.roles ?? []).map((r) => r.name));
      })
      .catch(() => {
        if (!cancelled) {
          toast.error(t("team.toast.userLoadFailed"));
          onClose();
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [userId, onClose, t]);

  // Pending = invited but never accepted. It can't be reactivated from here —
  // the user has to finish setting their password first.
  const isPending = user?.status === "pending";

  // Admins only manage roles + account status. Profile details (name, email,
  // language) are edited by the user themselves on their own Profile page.
  const baseRoles = (user?.roles ?? []).map((r) => r.name);
  const dirty =
    roleNames.slice().sort().join(",") !== baseRoles.slice().sort().join(",");

  const toggleRole = (nm: string) =>
    setRoleNames((rs) =>
      rs.includes(nm) ? rs.filter((r) => r !== nm) : [...rs, nm],
    );

  const save = async () => {
    if (!user || !dirty) return;
    setBusy(true);
    try {
      await usersHttpService.assignRoles(user.id, roleNames);
      toast.success(t("team.toast.rolesUpdated"));
      onChanged();
      onClose();
    } catch (err) {
      toast.error(userError(err, t));
    } finally {
      setBusy(false);
    }
  };

  const resetTfa = async () => {
    if (!user) return;
    setBusy(true);
    try {
      await usersHttpService.resetTfa(user.id);
      toast.success(t("team.toast.tfaReset"));
      setConfirmResetTfa(false);
      onChanged();
      onClose();
    } catch (err) {
      toast.error(userError(err, t));
    } finally {
      setBusy(false);
    }
  };

  const setActivated = async (activate: boolean) => {
    if (!user) return;
    setBusy(true);
    try {
      if (activate) {
        await usersHttpService.update(user.id, { status: "active" });
        toast.success(t("team.toast.userReactivated"));
      } else {
        await usersHttpService.deactivate(user.id);
        toast.success(t("team.toast.userDeactivated"));
      }
      onChanged();
      onClose();
    } catch (err) {
      toast.error(userError(err, t));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[640px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {loading || !user ? (
          <div className="flex flex-1 items-center justify-center text-muted-foreground">
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            {t("team.drawer.loading")}
          </div>
        ) : (
          <>
            <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-5">
              <div className="flex min-w-0 items-center gap-3">
                <Avatar user={user} size={48} />
                <div className="min-w-0">
                  <h2 className="truncate text-lg font-semibold">
                    {fullName(user)}
                  </h2>
                  <div className="truncate text-xs text-muted-foreground">
                    <span className="font-mono">{user.email}</span>
                    {user.created_at
                      ? ` · ${t("team.drawer.created", { time: relativeTime(user.created_at, t) })}`
                      : ""}
                  </div>
                </div>
              </div>
              <button
                onClick={onClose}
                className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                <X size={16} />
              </button>
            </header>

            <div className="flex-1 space-y-6 overflow-y-auto p-6">
              <DrawerSection
                title={t("team.drawer.profileTitle")}
                subtitle={t("team.drawer.profileSubtitle")}
              >
                <dl className="grid grid-cols-[110px_1fr] gap-y-2.5 text-xs">
                  <InfoRow k={t("team.drawer.name")}>
                    {user.name?.trim() || "—"}
                  </InfoRow>
                  <InfoRow k={t("team.drawer.email")}>
                    <span className="font-mono">{user.email}</span>
                  </InfoRow>
                  <InfoRow k={t("team.drawer.language")}>
                    {user.lang_key
                      ? langLabel(user.lang_key)
                      : t("team.langDefault")}
                  </InfoRow>
                </dl>
              </DrawerSection>

              <DrawerSection title={t("team.drawer.accountTitle")}>
                <ul className="space-y-1 text-[11px] text-muted-foreground">
                  <li className="flex items-center gap-2">
                    {t("team.drawer.status")}:{" "}
                    <StatusBadge status={user.status} />
                  </li>
                  <li className="flex items-center gap-2">
                    <span>
                      {t("team.drawer.tfa")}:{" "}
                      {user.tfa_enabled ? (
                        <span className="text-emerald-600 dark:text-emerald-400">
                          {t("team.drawer.tfaEnabled")}
                        </span>
                      ) : (
                        t("team.drawer.tfaNotConfigured")
                      )}
                    </span>
                    {user.tfa_enabled && (
                      <button
                        type="button"
                        disabled={busy}
                        onClick={() => setConfirmResetTfa(true)}
                        className="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium text-red-500 hover:bg-red-500/10 disabled:opacity-50"
                        title={t("team.drawer.resetTfaHint")}
                      >
                        <ShieldOff size={11} />
                        {t("team.drawer.resetTfa")}
                      </button>
                    )}
                  </li>
                  {isPending && (
                    <li className="text-amber-600 dark:text-amber-300">
                      {t("team.drawer.pendingNote")}
                    </li>
                  )}
                  {user.federated && <li>{t("team.drawer.federatedNote")}</li>}
                  {user.updated_at && (
                    <li>
                      {t("team.drawer.lastModified", {
                        time: relativeTime(user.updated_at, t),
                      })}
                    </li>
                  )}
                </ul>
              </DrawerSection>

              <DrawerSection
                title={t("team.drawer.rolesTitle")}
                subtitle={t("team.drawer.rolesSubtitle")}
              >
                <RolePicker
                  roles={roles}
                  selected={roleNames}
                  onToggle={toggleRole}
                />
                {roleNames.length === 0 && (
                  <p className="mt-2 text-[11px] text-amber-600 dark:text-amber-300">
                    {t("team.drawer.noRoleWarning")}
                  </p>
                )}
              </DrawerSection>
            </div>

            <footer className="flex items-center justify-between gap-3 border-t border-border px-6 py-3">
              {user.status === "active" ? (
                <Button
                  variant="outline"
                  size="sm"
                  className="text-red-500 hover:bg-red-500/10"
                  disabled={busy}
                  onClick={() => void setActivated(false)}
                >
                  <UserX size={13} className="mr-1.5" />
                  {t("team.drawer.deactivate")}
                </Button>
              ) : isPending ? (
                <span className="text-[11px] text-muted-foreground">
                  {t("team.drawer.waitingInvite")}
                </span>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={busy}
                  onClick={() => void setActivated(true)}
                >
                  <UserCheck size={13} className="mr-1.5" />
                  {t("team.drawer.reactivate")}
                </Button>
              )}
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={onClose}>
                  {t("team.drawer.cancel")}
                </Button>
                <Button
                  size="sm"
                  disabled={!dirty || busy}
                  onClick={() => void save()}
                >
                  {busy ? t("team.drawer.saving") : t("team.drawer.saveRoles")}
                </Button>
              </div>
            </footer>

            {confirmResetTfa && (
              <div
                className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
                onClick={() => !busy && setConfirmResetTfa(false)}
              >
                <div
                  className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
                  onClick={(e) => e.stopPropagation()}
                >
                  <header className="flex items-center gap-2 border-b border-border px-6 py-4">
                    <ShieldOff size={17} className="text-red-500" />
                    <h2 className="text-base font-semibold">
                      {t("team.drawer.resetTfaConfirmTitle")}
                    </h2>
                  </header>
                  <div className="px-6 py-5 text-sm text-muted-foreground">
                    {t("team.drawer.resetTfaConfirmBody", {
                      name: fullName(user),
                    })}
                  </div>
                  <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={busy}
                      onClick={() => setConfirmResetTfa(false)}
                    >
                      {t("team.drawer.cancel")}
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      disabled={busy}
                      onClick={() => void resetTfa()}
                    >
                      {busy
                        ? t("team.drawer.saving")
                        : t("team.drawer.resetTfaConfirm")}
                    </Button>
                  </footer>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

export function DrawerSection({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
}) {
  return (
    <section>
      <div className="mb-3">
        <h3 className="text-sm font-semibold">{title}</h3>
        {subtitle && (
          <p className="mt-0.5 text-[11px] text-muted-foreground">{subtitle}</p>
        )}
      </div>
      <div className="space-y-4">{children}</div>
    </section>
  );
}

export function InfoRow({
  k,
  children,
}: {
  k: string;
  children: React.ReactNode;
}) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  );
}
