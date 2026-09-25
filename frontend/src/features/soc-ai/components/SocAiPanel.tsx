import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ArrowUp,
  Maximize2,
  Minimize2,
  Paperclip,
  Sparkles,
  Trash2,
  X,
} from "lucide-react";
import { cn } from "@/shared/lib/utils";
import { useSocAi } from "../SocAiProvider";
import { MessageRow } from "./MessageRow";

// Panel-visible scopes only — 'home' has its own inline transcript and never
// shows here, so it needs no title/empty-state copy in this map.
const SCOPE_TITLE_KEY: Record<
  "panel" | "dashboard-create" | "dashboard-edit" | "soar-edit",
  string
> = {
  panel: "socAi.chat.title",
  "dashboard-create": "socAi.chat.dashboardCreateTitle",
  "dashboard-edit": "socAi.chat.dashboardEditTitle",
  "soar-edit": "socAi.chat.soarEditTitle",
};

export function SocAiPanel() {
  const { t } = useTranslation();
  const {
    open,
    expanded,
    activeScope,
    messages,
    dashboardCreateMessages,
    dashboardEditMessages,
    soarEditMessages,
    dashboardEditTarget,
    soarEditTarget,
    focus,
    detachFocus,
    closePanel,
    toggleExpand,
    clear,
    submit,
  } = useSocAi();
  const [draft, setDraft] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);
  const taRef = useRef<HTMLTextAreaElement>(null);

  // Auto-grow the input, capped at 7 lines (max-h-[140px]) so long prompts
  // scroll internally instead of eating the message area.
  useEffect(() => {
    const el = taRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${el.scrollHeight}px`;
  }, [draft]);

  const activeMessages =
    activeScope === "dashboard-create"
      ? dashboardCreateMessages
      : activeScope === "dashboard-edit"
        ? dashboardEditMessages
        : activeScope === "soar-edit"
          ? soarEditMessages
          : messages;
  // 'home' never opens this panel (see the comment above), so it has no
  // entry here — fall back to the general panel title if it ever does.
  const titleKey =
    SCOPE_TITLE_KEY[activeScope as keyof typeof SCOPE_TITLE_KEY] ??
    SCOPE_TITLE_KEY.panel;

  // Stick to the bottom as messages stream in.
  useEffect(() => {
    if (scrollRef.current)
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [activeMessages]);

  // No queueing — block sending while the last message is still being answered.
  const last = activeMessages[activeMessages.length - 1];
  const isPending = last?.role === "ai" && !!last.pending;

  const send = () => {
    if (!draft.trim() || isPending) return;
    submit(draft, { scope: activeScope });
    setDraft("");
  };

  const width = expanded ? "w-[min(720px,95vw)]" : "w-[min(420px,95vw)]";

  // A column of the layout, not an overlay: it takes its width from the page
  // (see DashboardLayout), so everything beside it — including the drawers —
  // shrinks instead of being covered. The inner panel keeps its full width
  // while the column animates, so the chat doesn't reflow as it slides.
  return (
    <div
      inert={!open}
      className={cn(
        "shrink-0 overflow-hidden transition-[width] duration-200",
        open ? width : "w-0",
      )}
    >
      <aside
        role="dialog"
        aria-label="SOC Assistant"
        className={cn("flex h-full flex-col border-l border-border bg-background", width)}
      >
        <header className="flex shrink-0 items-center justify-between border-b border-border px-4 py-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2 text-[15px] font-semibold">
              <Sparkles size={18} className="text-primary" />
              <span>{t(titleKey)}</span>
            </div>
            {activeScope === "dashboard-edit" && dashboardEditTarget && (
              <p className="mt-0.5 truncate pl-[26px] text-xs text-muted-foreground">
                {t("socAi.chat.editingDashboard", {
                  name: dashboardEditTarget.name,
                })}
              </p>
            )}
            {activeScope === "soar-edit" && soarEditTarget && (
              <p className="mt-0.5 truncate pl-[26px] text-xs text-muted-foreground">
                {t("socAi.chat.editingFlow", { name: soarEditTarget.name })}
              </p>
            )}
          </div>
          <div className="flex items-center gap-0.5">
            <IconBtn
              label={expanded ? t("socAi.chat.collapse") : t("socAi.chat.expand")}
              onClick={toggleExpand}
            >
              {expanded ? <Minimize2 size={16} /> : <Maximize2 size={16} />}
            </IconBtn>
            <IconBtn
              label={t("socAi.chat.clear")}
              onClick={() => clear(activeScope)}
            >
              <Trash2 size={16} />
            </IconBtn>
            <IconBtn label={t("socAi.chat.close")} onClick={closePanel}>
              <X size={18} />
            </IconBtn>
          </div>
        </header>

        <div
          ref={scrollRef}
          className="flex flex-1 flex-col gap-4 overflow-y-auto px-4 py-5 text-[13.5px] leading-relaxed"
        >
          {activeMessages.length === 0 ? (
            <div className="m-auto max-w-[260px] text-center text-sm text-muted-foreground">
              <Sparkles size={22} className="mx-auto mb-3 text-primary/70" />
              <p>{t("socAi.chat.empty")}</p>
              <p className="mt-3 text-xs">
                {t("socAi.chat.tryPrefix")}{" "}
                <span className="text-foreground">
                  “{t("socAi.chat.suggestion")}”
                </span>
              </p>
            </div>
          ) : (
            activeMessages.map((m) => <MessageRow key={m.id} message={m} />)
          )}
        </div>

        <div className="shrink-0 border-t border-border bg-muted/30 p-3">
          {/* What the assistant is being told is open beside it. Only the general
              chat uses it; the dashboard threads carry their own context. */}
          {focus && activeScope === "panel" && (
            <div className="mb-2 flex items-center gap-1.5 rounded-md border border-primary/30 bg-primary/5 px-2 py-1 text-xs">
              <Paperclip size={12} className="shrink-0 text-primary" />
              <span className="shrink-0 text-muted-foreground">
                {t(`socAi.focus.${focus.kind}`)}
              </span>
              <span className="min-w-0 flex-1 truncate font-medium" title={focus.label}>
                {focus.label}
              </span>
              <button
                type="button"
                onClick={detachFocus}
                aria-label={t("socAi.focus.remove")}
                title={t("socAi.focus.remove")}
                className="flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
              >
                <X size={12} />
              </button>
            </div>
          )}
          <div className="relative rounded-lg border border-primary/40 bg-background p-2.5 pr-12 transition-colors focus-within:border-primary">
            <textarea
              ref={taRef}
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  send();
                }
              }}
              rows={1}
              placeholder={t("socAi.chat.placeholder")}
              className="max-h-[140px] w-full resize-none bg-transparent text-sm leading-5 text-foreground outline-none placeholder:text-muted-foreground"
            />
            <button
              type="button"
              onClick={send}
              disabled={!draft.trim() || isPending}
              aria-label={t("socAi.chat.send")}
              className="absolute bottom-2 right-2 flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground transition-opacity hover:opacity-90 active:scale-95 disabled:opacity-40"
            >
              <ArrowUp size={16} strokeWidth={2.5} />
            </button>
          </div>
        </div>
      </aside>
    </div>
  );
}

function IconBtn({
  label,
  onClick,
  children,
}: {
  label: string;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
    >
      {children}
    </button>
  );
}
