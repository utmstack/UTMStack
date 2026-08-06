import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Plus, RefreshCw, Search, TagIcon } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { useTaggingRulesList } from '../hooks/use-tagging-rules-list'
import { useTaggingRuleMutations } from '../hooks/use-tagging-rule-mutations'
import { useAlertTagCatalog } from '../hooks/use-alert-tag-catalog'
import { TaggingRulesTable } from '../components/tagging-rules-table'
import { TaggingRuleDrawer } from '../components/tagging-rule-drawer'
import { TaggingRulesEmptyCard } from '../components/tagging-rules-empty-card'
import type { TaggingRule, TaggingRuleListParams } from '../types/tagging-rule.types'

export function TaggingRulesPage() {
  const { t } = useTranslation()

  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize] = useState(50)
  const [tagFilter, setTagFilter] = useState<string | 'all'>('all')

  const [open, setOpen] = useState<TaggingRule | null>(null)
  const [creating, setCreating] = useState(false)
  const [pendingDelete, setPendingDelete] = useState<TaggingRule | null>(null)
  const [deleting, setDeleting] = useState(false)

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

  const submit = async (input: { name: string; description: string; conditions: any[]; tags: any[] }, id?: string) => {
    const ok = id != null
      ? await updateRule({ id, ...input })
      : await createRule(input)
    if (ok) {
      setCreating(false)
      setOpen(null)
    }
  }

  const remove = async (r: TaggingRule) => setPendingDelete(r)

  const confirmDelete = async () => {
    if (!pendingDelete) return
    setDeleting(true)
    try {
      const ok = await deleteRule(pendingDelete.id, pendingDelete.name)
      if (ok) setOpen(null)
    } finally {
      setDeleting(false)
      setPendingDelete(null)
    }
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <TagIcon size={14} strokeWidth={1.75} />
          <span><span className="font-medium text-foreground">{total}</span> {t('taggingRules.title').toLowerCase()}</span>
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
          value={tagFilter}
          onChange={(e) => {
            setTagFilter(e.target.value)
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
        <TaggingRulesEmptyCard>
          <AlertTriangle size={16} className="text-amber-500" />
          {t('taggingRules.loadError')}
          <button onClick={refresh} className="ml-2 text-primary hover:underline">
            {t('taggingRules.retry')}
          </button>
        </TaggingRulesEmptyCard>
      ) : loading && rules.length === 0 ? (
        <TaggingRulesEmptyCard>
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </TaggingRulesEmptyCard>
      ) : rules.length === 0 ? (
        <TaggingRulesEmptyCard>{t('taggingRules.empty')}</TaggingRulesEmptyCard>
      ) : (
        <div className="min-h-0 flex-1 overflow-y-auto">
          <TaggingRulesTable rules={rules} onOpen={setOpen} />
          <InfiniteScrollSentinel
            onReach={() => setPage((p) => p + 1)}
            hasMore={rules.length < total}
            loading={loading}
            endLabel={t('common.allLoaded', { count: total })}
          />
        </div>
      )}

      {open && (
        <TaggingRuleDrawer
          rule={open}
          tagCatalog={tagCatalog}
          onClose={() => setOpen(null)}
          onSubmit={(input, id) => submit(input, id ?? open.id)}
          onDelete={remove}
          onCreateTag={createTag}
        />
      )}
      {creating && (
        <TaggingRuleDrawer
          create
          tagCatalog={tagCatalog}
          onClose={() => setCreating(false)}
          onSubmit={(input) => submit(input)}
          onCreateTag={createTag}
        />
      )}

      <ConfirmDialog
        open={pendingDelete != null}
        title={t('taggingRules.deleteTitle') ?? 'Delete rule'}
        body={pendingDelete ? t('taggingRules.deleteConfirm', { name: pendingDelete.name }) : ''}
        confirmLabel={t('common.actions.delete') ?? undefined}
        danger
        busy={deleting}
        onClose={() => !deleting && setPendingDelete(null)}
        onConfirm={confirmDelete}
      />
    </div>
  )
}

