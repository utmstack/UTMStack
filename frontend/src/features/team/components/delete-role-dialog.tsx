import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/shared/components/ui/button";
import { roleError, roleLabel } from "../lib/team-utils";
import { rolesHttpService } from "../services/team-http.service";
import type { RoleDetail } from "../types/team.types";

export function DeleteRoleDialog({
  role,
  onClose,
  onDeleted,
}: {
  role: RoleDetail;
  onClose: () => void;
  onDeleted: () => void;
}) {
  const { t } = useTranslation();
  const [busy, setBusy] = useState(false);

  const remove = async () => {
    setBusy(true);
    try {
      await rolesHttpService.remove(role.id);
      toast.success(t("team.toast.roleDeleted"));
      onDeleted();
    } catch (err) {
      toast.error(roleError(err, t));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={() => !busy && onClose()}
    >
      <div
        className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center gap-2 border-b border-border px-6 py-4">
          <Trash2 size={17} className="text-red-500" />
          <h2 className="text-base font-semibold">
            {t("team.roles.deleteConfirmTitle")}
          </h2>
        </header>
        <div className="px-6 py-5 text-sm text-muted-foreground">
          {t("team.roles.deleteConfirmBody", {
            name: roleLabel(t, role.name, role.display_name),
          })}
        </div>
        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button variant="outline" size="sm" disabled={busy} onClick={onClose}>
            {t("team.drawer.cancel")}
          </Button>
          <Button
            variant="destructive"
            size="sm"
            disabled={busy}
            onClick={() => void remove()}
          >
            {busy ? t("team.drawer.saving") : t("team.roles.delete")}
          </Button>
        </footer>
      </div>
    </div>
  );
}
