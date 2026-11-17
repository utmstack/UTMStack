import { useState } from "react";
import { useTranslation } from "react-i18next";
import { UserX } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/shared/components/ui/button";
import { fullName, userError } from "../lib/team-utils";
import { usersHttpService } from "../services/team-http.service";
import type { UserBase } from "../types/team.types";

export function RemoveUserDialog({
  user,
  onClose,
  onRemoved,
}: {
  user: UserBase;
  onClose: () => void;
  onRemoved?: () => void;
}) {
  const { t } = useTranslation();
  const [busy, setBusy] = useState(false);

  const remove = async () => {
    setBusy(true);
    try {
      await usersHttpService.remove(user.id);
      toast.success(t("team.toast.userRemoved"));
      onRemoved?.();
      onClose();
    } catch (err) {
      toast.error(userError(err, t));
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
          <UserX size={17} className="text-red-500" />
          <h2 className="text-base font-semibold">
            {t("team.drawer.removeConfirmTitle",{name: fullName(user)})}
          </h2>
        </header>
        <div className="px-6 py-5 text-sm text-muted-foreground">
          {t("team.drawer.removeConfirmBody", { name: fullName(user) })}
        </div>
        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button variant="outline" size="sm" disabled={busy} onClick={onClose}>
            {t("team.drawer.cancel")}
          </Button>
          <Button variant="destructive" size="sm" disabled={busy} onClick={() => void remove()}>
            {t("team.drawer.remove")}
          </Button>
        </footer>
      </div>
    </div>
  );
}
