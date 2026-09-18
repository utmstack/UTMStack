import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { AlertCircle, Check, KeyRound, UserPlus, X } from "lucide-react";
import { Button } from "@/shared/components/ui/button";
import { Input } from "@/shared/components/ui/input";
import { cn } from "@/shared/lib/utils";
import { toast } from "sonner";
import { userError } from "../lib/team-utils";
import { usersHttpService } from "../services/team-http.service";
import type { Role } from "../types/team.types";
import { RolePicker } from "./role-picker";

type Mode = "invite" | "local";

function isPasswordValid(pw: string): boolean {
  return (
    pw.length >= 8 &&
    /[A-Z]/.test(pw) &&
    /[a-z]/.test(pw) &&
    /[^A-Za-z0-9]/.test(pw)
  );
}

export function InviteDialog({
  roles,
  onClose,
  onCreated,
}: {
  roles: Role[];
  onClose: () => void;
  onCreated: () => void;
}) {
  const { t } = useTranslation();
  const [mode, setMode] = useState<Mode>("invite");
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [roleNames, setRoleNames] = useState<string[]>(["ROLE_VIEWER"]);
  const [busy, setBusy] = useState(false);

  // Errors surface only after the user pauses typing (200ms), not mid-keystroke.
  const [idle, setIdle] = useState(true);
  useEffect(() => {
    setIdle(false);
    const handle = setTimeout(() => setIdle(true), 200);
    return () => clearTimeout(handle);
  }, [email, password]);

  const isLocal = mode === "local";
  const emailValid = /.+@.+\..+/.test(email.trim());
  const emailInvalid = idle && email.length > 0 && !emailValid;
  const pwdInvalid =
    idle && isLocal && password.length > 0 && !isPasswordValid(password);
  const valid = emailValid && (!isLocal || isPasswordValid(password));

  const toggleRole = (nm: string) =>
    setRoleNames((rs) =>
      rs.includes(nm) ? rs.filter((r) => r !== nm) : [...rs, nm],
    );

  const submit = async () => {
    if (!valid || busy) return;
    setBusy(true);
    try {
      await usersHttpService.create({
        email: email.trim(),
        name: name.trim() || undefined,
        // No lang_key: the user inherits the platform-default language until they
        // pick their own in their profile.
        role_names: roleNames,
        // Local users get a password set directly and are activated immediately;
        // invited users receive an email instead.
        ...(isLocal ? { password } : {}),
      });
      toast.success(
        isLocal
          ? t("team.toast.localUserCreated")
          : t("team.toast.invitationSent"),
      );
      onCreated();
    } catch (err) {
      toast.error(userError(err, t));
    } finally {
      setBusy(false);
    }
  };

  const pwdChecks: { ok: boolean; label: string }[] = [
    { ok: password.length >= 8, label: t("team.invite.passwordRules.length") },
    {
      ok: /[A-Z]/.test(password) && /[a-z]/.test(password),
      label: t("team.invite.passwordRules.case"),
    },
    {
      ok: /[^A-Za-z0-9]/.test(password),
      label: t("team.invite.passwordRules.special"),
    },
  ];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <UserPlus size={18} />
              {t("team.invite.title")}
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">
              {isLocal
                ? t("team.invite.localDescription")
                : t("team.invite.description")}
            </p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="space-y-4 px-6 py-5">
          <div className="grid grid-cols-2 gap-1 rounded-md border border-border bg-muted/40 p-1">
            <ModeBtn
              active={!isLocal}
              onClick={() => setMode("invite")}
              icon={UserPlus}
              label={t("team.invite.modeInvite")}
            />
            <ModeBtn
              active={isLocal}
              onClick={() => setMode("local")}
              icon={KeyRound}
              label={t("team.invite.modeLocal")}
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <DrawerField
              label={t("team.invite.email")}
              hint={emailInvalid ? undefined : t("team.invite.emailHint")}
            >
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="jsmith@company.com"
                className={emailInvalid ? "border-red-500" : undefined}
              />
              {emailInvalid && (
                <p className="flex items-center gap-1 text-[11px] text-red-500">
                  <AlertCircle size={11} />
                  {t("team.invite.emailInvalid")}
                </p>
              )}
            </DrawerField>
            <DrawerField label={t("team.invite.name")}>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Jane Smith"
              />
            </DrawerField>
          </div>

          {isLocal && (
            <div className="space-y-1.5">
              <label className="block text-xs font-medium text-foreground/80">
                {t("team.invite.passwordLabel")}
              </label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="new-password"
                placeholder={t("team.invite.passwordPlaceholder")}
                className={pwdInvalid ? "border-red-500" : undefined}
              />
              <div className="flex flex-wrap gap-x-3 gap-y-0.5">
                {pwdChecks.map((c) => {
                  const showRed = idle && password.length > 0 && !c.ok;
                  return (
                    <span
                      key={c.label}
                      className={cn(
                        "flex items-center gap-1 text-[11px]",
                        c.ok
                          ? "text-emerald-600 dark:text-emerald-400"
                          : showRed
                            ? "text-red-500"
                            : "text-muted-foreground",
                      )}
                    >
                      {c.ok ? (
                        <Check size={11} strokeWidth={3} />
                      ) : showRed ? (
                        <AlertCircle size={11} />
                      ) : (
                        <span className="inline-block h-1 w-1 rounded-full bg-current" />
                      )}
                      {c.label}
                    </span>
                  );
                })}
              </div>
              <p className="text-[11px] text-muted-foreground">
                {t("team.invite.localUserNote")}
              </p>
            </div>
          )}

          <div>
            <label className="mb-1.5 block text-xs font-medium text-foreground/80">
              {t("team.invite.roles")}
            </label>
            <RolePicker
              roles={roles}
              selected={roleNames}
              onToggle={toggleRole}
              compact
            />
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t("team.invite.cancel")}
          </Button>
          <Button
            size="sm"
            disabled={!valid || busy}
            onClick={() => void submit()}
          >
            <UserPlus size={13} className="mr-1.5" />
            {busy
              ? t("team.invite.sending")
              : isLocal
                ? t("team.invite.createLocal")
                : t("team.invite.send")}
          </Button>
        </footer>
      </div>
    </div>
  );
}

function ModeBtn({
  active,
  onClick,
  icon: Icon,
  label,
}: {
  active: boolean;
  onClick: () => void;
  icon: typeof UserPlus;
  label: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex items-center justify-center gap-1.5 rounded px-3 py-1.5 text-xs font-medium transition-colors",
        active
          ? "bg-card text-foreground shadow-sm"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      <Icon size={13} />
      {label}
    </button>
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
