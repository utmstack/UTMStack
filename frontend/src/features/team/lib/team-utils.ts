import type { TFunction } from "i18next";
import { SUPPORTED_LANGUAGES } from "@/shared/i18n";
import { TeamHttpError } from "../services/team-http.service";
import type { UserBase } from "../types/team.types";

export const PAGE_SIZE = 20;

/* Roles & permissions arrive from the backend as English DB strings. Translate the
 * stable identifiers (role name, permission resource/action) with a fallback to the
 * backend value so any custom/unknown role still renders something sensible. */
export function roleLabel(
  t: TFunction,
  name: string,
  fallback?: string,
): string {
  return t(`team.roleNames.${name}`, { defaultValue: fallback || name });
}
export function roleDesc(
  t: TFunction,
  name: string,
  fallback?: string,
): string {
  return t(`team.roleDescriptions.${name}`, { defaultValue: fallback || "" });
}
export function resourceLabel(t: TFunction, resource: string): string {
  return t(`team.resources.${resource}`, { defaultValue: resource });
}
export function actionLabel(t: TFunction, action: string): string {
  return t(`team.actions.${action}`, { defaultValue: action });
}

/* Permissions are flat "resource.action" strings; the two halves are what the
 * matrix groups and labels by. */
export function permResource(name: string): string {
  return name.split(".")[0] ?? name;
}
export function permAction(name: string): string {
  return name.split(".")[1] ?? "";
}

export function langLabel(code: string): string {
  return SUPPORTED_LANGUAGES.find((l) => l.code === code)?.label ?? code;
}

// An account can exist before anyone has typed a display name into it — every
// federated one starts that way — so the address is what is always there.
export function fullName(u: UserBase): string {
  return u.name?.trim() || u.email;
}

export function initials(u: UserBase): string {
  return (
    fullName(u)
      .split(/[\s.@_-]+/)
      .map((p) => p[0])
      .filter(Boolean)
      .slice(0, 2)
      .join("")
      .toUpperCase() || "U"
  );
}

export function relativeTime(iso: string, t: TFunction): string {
  if (!iso) return "—";
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.round(diff / 60_000);
  if (m < 1) return t("team.relative.justNow");
  if (m < 60) return t("team.relative.minutesAgo", { count: m });
  const h = Math.round(m / 60);
  if (h < 24) return t("team.relative.hoursAgo", { count: h });
  const d = Math.round(h / 24);
  if (d < 30) return t("team.relative.daysAgo", { count: d });
  return new Date(iso).toLocaleDateString();
}

export function roleError(err: unknown, t: TFunction): string {
  if (err instanceof TeamHttpError) {
    if (err.status === 409) return t("team.toast.roleNameInUse");
    // The backend refuses to touch a seeded role whatever the request says.
    if (err.status === 403) return t("team.toast.roleImmutable");
    if (err.status === 400)
      return err.message || t("team.toast.invalidRequest");
  }
  return err instanceof Error ? err.message : t("team.toast.operationFailed");
}

export function userError(err: unknown, t: TFunction): string {
  if (err instanceof TeamHttpError) {
    if (err.status === 409) return t("team.toast.emailInUse");
    if (err.status === 400)
      return err.message || t("team.toast.invalidRequest");
    if (err.status === 403) return t("team.toast.noPermission");
    if (err.status === 404) return t("team.toast.userNotFound");
  }
  return err instanceof Error ? err.message : t("team.toast.operationFailed");
}
