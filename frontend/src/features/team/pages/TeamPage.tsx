import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  AlertTriangle,
  Loader2,
  Search,
  Shield,
  UserCheck,
  UserPlus,
} from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/components/ui/button";
import { Input } from "@/shared/components/ui/input";
import { InfiniteScrollSentinel } from "@/shared/components/ui/infinite-scroll";
import { ColumnResizeHandle } from "@/shared/components/ui/column-resize-handle";
import {
  colMins,
  useResizableColumns,
} from "@/shared/hooks/useResizableColumns";
import { MEMBER_COLS, PAGE_SIZE } from "../lib/team-utils";
import {
  rolesHttpService,
  usersHttpService,
} from "../services/team-http.service";
import type { PageInfo, Role, UserListItem } from "../types/team.types";
import { MemberDrawer } from "../components/member-drawer";
import { InviteDialog } from "../components/invite-dialog";
import { MemberRow } from "../components/member-row";
import { RolesView } from "../components/roles-view";

type View = "members" | "roles";

export function TeamPage() {
  const { t } = useTranslation();
  const [view, setView] = useState<View>("members");
  const [roles, setRoles] = useState<Role[]>([]);

  // Roles list is small and shared by the pickers + the roles tab.
  const loadRoles = useCallback(async () => {
    try {
      setRoles(await rolesHttpService.list());
    } catch {
      toast.error(t("team.toast.rolesLoadFailed"));
    }
  }, [t]);

  useEffect(() => {
    void loadRoles();
  }, [loadRoles]);

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <UserCheck size={14} strokeWidth={1.75} />
          <span className="font-medium text-foreground">{t("team.title")}</span>
        </div>
      </header>

      <div className="mt-3 flex items-center gap-1 border-b border-border">
        <TabBtn
          active={view === "members"}
          onClick={() => setView("members")}
          icon={UserCheck}
        >
          {t("team.tabs.members")}
        </TabBtn>
        <TabBtn
          active={view === "roles"}
          onClick={() => setView("roles")}
          icon={Shield}
        >
          {t("team.tabs.roles")}
        </TabBtn>
      </div>

      {view === "members" ? (
        <MembersView roles={roles} />
      ) : (
        <RolesView roles={roles} onChanged={() => void loadRoles()} />
      )}
    </div>
  );
}

function TabBtn({
  active,
  onClick,
  icon: Icon,
  children,
}: {
  active: boolean;
  onClick: () => void;
  icon: typeof UserCheck;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "relative flex items-center gap-1.5 px-3 py-2 text-sm transition-colors",
        active
          ? "text-foreground"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      <Icon size={14} strokeWidth={1.75} />
      {children}
      {active && (
        <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />
      )}
    </button>
  );
}

function MembersView({ roles }: { roles: Role[] }) {
  const { t } = useTranslation();
  const [users, setUsers] = useState<UserListItem[] | null>(null);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debounced, setDebounced] = useState("");

  const [openId, setOpenId] = useState<string | null>(null);
  const [inviteOpen, setInviteOpen] = useState(false);
  const memberHeaders = [
    t("team.members.colUser"),
    t("team.members.colRoles"),
    t("team.members.col2fa"),
    t("team.members.colStatus"),
    "",
  ];
  const memberMins = [...colMins(memberHeaders.slice(0, -1)), 40];
  const { template: memberCols, startDrag } = useResizableColumns(MEMBER_COLS, {
    min: memberMins,
    storageKey: "team-members-table-columns",
  });

  // Debounce the search box, and reset to page 1 when the query changes.
  useEffect(() => {
    const handle = setTimeout(() => {
      setDebounced(search.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(handle);
  }, [search]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const resp = await usersHttpService.list({
        page,
        page_size: PAGE_SIZE,
        search: debounced,
      });
      setUsers((prev) =>
        page === 1 ? resp.data : [...(prev ?? []), ...resp.data],
      );
      setPageInfo(resp.page_info);
    } catch {
      setError(true);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, [page, debounced]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <>
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("team.members.searchPlaceholder")}
            className="h-9 pl-9"
          />
        </div>
        <Button size="sm" onClick={() => setInviteOpen(true)}>
          <UserPlus size={14} className="mr-1.5" />
          {t("team.members.invite")}
        </Button>
      </div>

      <div className="mt-3 overflow-x-auto overflow-y-hidden rounded-xl border border-border bg-card">
        <div
          className="grid w-max min-w-full items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: memberCols }}
        >
          {memberHeaders.map((header, index, headers) => (
            <div
              key={index}
              data-resizable-col
              className="relative min-w-0 pr-2 last:pr-0"
            >
              {header}
              {index < headers.length - 1 && (
                <ColumnResizeHandle onMouseDown={startDrag(index)} />
              )}
            </div>
          ))}
        </div>

        {loading && (!users || users.length === 0) && (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t("team.members.loading")}
          </div>
        )}

        {!loading && error && (
          <div className="flex flex-col items-center gap-3 px-6 py-12 text-sm">
            <span className="inline-flex items-center gap-2 text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t("team.members.loadFailed")}
            </span>
            <Button variant="outline" size="sm" onClick={() => void load()}>
              {t("team.members.retry")}
            </Button>
          </div>
        )}

        {!loading && !error && users && users.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            {debounced
              ? t("team.members.noSearchMatch")
              : t("team.members.none")}
          </div>
        )}

        {users &&
          users.length > 0 &&
          users.map((u) => (
            <MemberRow
              key={u.id}
              user={u}
              tableCols={memberCols}
              onOpen={() => setOpenId(u.id)}
            />
          ))}
      </div>

      {users && users.length > 0 && (
        <InfiniteScrollSentinel
          onReach={() => setPage((p) => p + 1)}
          hasMore={users.length < (pageInfo?.total_items ?? 0)}
          loading={loading}
          endLabel={t("common.allLoaded", {
            count: pageInfo?.total_items ?? 0,
          })}
        />
      )}

      {openId != null && (
        <MemberDrawer
          userId={openId}
          roles={roles}
          onClose={() => setOpenId(null)}
          onChanged={() => void load()}
        />
      )}
      {inviteOpen && (
        <InviteDialog
          roles={roles}
          onClose={() => setInviteOpen(false)}
          onCreated={() => {
            setInviteOpen(false);
            setPage(1);
            void load();
          }}
        />
      )}
    </>
  );
}
