import { useEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Plus, RefreshCw, Search, TagIcon } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { useTaggingRulesList } from '../hooks/use-tagging-rules-list'
import { useTaggingRuleMutations } from '../hooks/use-tagging-rule-mutations'
import { useAlertTagCatalog } from '../hooks/use-alert-tag-catalog'
import { TaggingRulesTable } from '../components/tagging-rules-table'
import { TaggingRuleDrawer } from '../components/tagging-rule-drawer'
import { TAGGING_RULES_PAGE_SIZE } from '../lib/tagging-rule-meta'
import { taggingRulesHttpService } from '../services/tagging-rules-http.service'
import type { AlertTag, TaggingRule, TaggingRuleListParams } from '../types/tagging-rule.types'

export function TaggingRulesPage() {
  const { t } = useTranslation()

  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(TAGGING_RULES_PAGE_SIZE)
  const [tagFilter, setTagFilter] = useState<number | 'all'>('all')

  const [open, setOpen] = useState<TaggingRule | null>(null)
  const [openInEdit, setOpenInEdit] = useState(false)
  const [creating, setCreating] = useState(false)
  const [creatingWith, setCreatingWith] = useState<AlertTag[]>([])

  // Deep-link from the alert tag editor: if a rule already references this tag,
  // open the first match in edit mode; otherwise open the create drawer with
  // the tag pre-selected. Clear the router state on read so a refresh doesn't
  // re-seed over the user's edits.
  const location = useLocation()
  const navigate = useNavigate()
  const seededRef = useRef(false)
  useEffect(() => {
    const state = location.state as { createWithTag?: AlertTag } | null
    if (!state?.createWithTag || seededRef.current) return
    seededRef.current = true
    const tag = state.createWithTag
    navigate(location.pathname, { replace: true })
    void (async () => {
      try {
        const { data } = await taggingRulesHttpService.list({
          page: 1,
          size: 1,
          tagIds: [tag.id],
          ruleDeleted: false,
        })
        if (data && data.length > 0) {
          setOpenInEdit(true)
          setOpen(data[0])
          return
        }
      } catch {
        /* fall through to create */
      }
      setCreatingWith([tag])
      setCreating(true)
    })()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Debounce the search box.
  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const params = useMemo<TaggingRuleListParams>(
    () => ({
      page: page + 1, // backend is 1-based
      size: pageSize,
      name: debounced || undefined,
      tagIds: tagFilter === 'all' ? undefined : [tagFilter],
      ruleDeleted: false,
    }),
    [page, pageSize, debounced, tagFilter]
  )

  const { rules, total, loading, error, refresh } = useTaggingRulesList(params)
  const { tagCatalog, createTag } = useAlertTagCatalog()
  const { createRule, updateRule, deleteRule } = useTaggingRuleMutations(refresh)

  const submit = async (input: { name: string; description: string; conditions: any[]; tags: any[] }, id?: number) => {
    const ok = id != null
      ? await updateRule({ id, ...input })
      : await createRule(input)
    if (ok) {
      setCreating(false)
      setCreatingWith([])
      setOpen(null)
      setOpenInEdit(false)
    }
  }

  const remove = async (r: TaggingRule) => {
    if (!confirm(t('taggingRules.deleteConfirm', { name: r.name }))) return
    const ok = await deleteRule(r.id, r.name)
    if (ok) {
      setOpen(null)
      setOpenInEdit(false)
    }
  }

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 py-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-semibold">
            <TagIcon size={18} strokeWidth={1.75} />
            {t('taggingRules.title')}
            <span className="text-sm font-normal text-muted-foreground">· {total}</span>
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">{t('taggingRules.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" onClick={() => setCreating(true)}>
            <Plus size={14} className="mr-1.5" /> {t('taggingRules.new')}
          </Button>
          <button
            onClick={refresh}
            title={t('taggingRules.refresh')}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
          </button>
        </div>
      </header>

      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[220px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={t('taggingRules.toolbar.search')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9 pl-9"
          />
        </div>
        <select
          value={String(tagFilter)}
          onChange={(e) => {
            setTagFilter(e.target.value === 'all' ? 'all' : Number(e.target.value))
            setPage(0)
          }}
          className="h-9 rounded-md border border-border bg-background px-2 text-sm"
        >
          <option value="all">{t('taggingRules.toolbar.allTags')}</option>
          {tagCatalog.map((tg) => (
            <option key={tg.id} value={tg.id}>
              {tg.tagName}
            </option>
          ))}
        </select>
      </div>

      {error ? (
        <Center>
          <AlertTriangle size={16} className="text-amber-500" />
          {t('taggingRules.loadError')}
          <button onClick={refresh} className="ml-2 text-primary hover:underline">
            {t('taggingRules.retry')}
          </button>
        </Center>
      ) : loading && rules.length === 0 ? (
        <Center>
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </Center>
      ) : rules.length === 0 ? (
        <Center>{t('taggingRules.empty')}</Center>
      ) : (
        <>
          <TaggingRulesTable rules={rules} onOpen={setOpen} />
          <div className="mt-3 shrink-0">
            <Pagination
              page={page}
              pageSize={pageSize}
              total={total}
              onPageChange={setPage}
              onPageSizeChange={(s) => {
                setPageSize(s)
                setPage(0)
              }}
            />
          </div>
        </>
      )}

      {open && (
        <TaggingRuleDrawer
          rule={open}
          startInEdit={openInEdit}
          tagCatalog={tagCatalog}
          onClose={() => {
            setOpen(null)
            setOpenInEdit(false)
          }}
          onSubmit={(input, id) => submit(input, id ?? open.id)}
          onDelete={remove}
          onCreateTag={createTag}
        />
      )}
      {creating && (
        <TaggingRuleDrawer
          create
          initialTags={creatingWith}
          tagCatalog={tagCatalog}
          onClose={() => {
            setCreating(false)
            setCreatingWith([])
          }}
          onSubmit={(input) => submit(input)}
          onCreateTag={createTag}
        />
      )}
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return (
    <div className="mt-4 flex flex-1 items-center justify-center gap-2 rounded-xl border border-border bg-card text-sm text-muted-foreground">
      {children}
    </div>
  )
}
