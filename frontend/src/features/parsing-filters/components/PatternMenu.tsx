import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Lock, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import {
  regexPatternsHttpService,
  RegexPatternsHttpError,
  type RegexPattern,
} from '@/features/regex-patterns/services/regex-patterns-http.service'

// A pattern name must be a valid Go-template field so `{{.name}}` resolves.
const PATTERN_ID_RE = /^[A-Za-z_]\w*$/

/**
 * The inner body of a regex-pattern picker: a searchable list of named patterns,
 * a "+ new pattern" inline create form, and selection. The parent owns the
 * positioned popover container; this renders only its contents so it can be
 * reused both from the visual grok field and the YAML content editor.
 *
 * `query` is controlled by the parent. Pass `onQueryChange` to render a search
 * box; omit it when the query is driven externally (e.g. an open `{{.foo` token).
 */
export function PatternMenu({
  patternOptions,
  query,
  onQueryChange,
  onPick,
  onPatternCreated,
}: {
  patternOptions: RegexPattern[]
  query: string
  onQueryChange?: (q: string) => void
  onPick: (name: string) => void
  onPatternCreated?: (p: RegexPattern) => void
}) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState<{ id: string; def: string; desc: string } | null>(null)
  const [saving, setSaving] = useState(false)

  const create = async () => {
    if (!draft) return
    const id = draft.id.trim()
    if (!PATTERN_ID_RE.test(id)) {
      toast.error(t('parsingFilters.visual.patternIdInvalid'))
      return
    }
    if (!draft.def.trim()) {
      toast.error(t('parsingFilters.visual.patternDefRequired'))
      return
    }
    setSaving(true)
    try {
      const created = await regexPatternsHttpService.create({
        patternId: id,
        patternDescription: draft.desc,
        patternDefinition: draft.def,
      })
      onPatternCreated?.(created)
      toast.success(t('parsingFilters.visual.patternCreated', { name: id }))
      onPick(id) // closes the popover and inserts `{{.id}}`
    } catch (err) {
      toast.error(err instanceof RegexPatternsHttpError ? err.message : t('parsingFilters.visual.patternCreateError'))
    } finally {
      setSaving(false)
    }
  }

  if (draft) {
    return (
      <div className="space-y-2 p-3">
        <div className="text-[11px] font-semibold text-foreground">{t('parsingFilters.visual.newPattern')}</div>
        <div className="space-y-1">
          <label className="block font-mono text-[10px] text-foreground/60">{t('regexPatterns.editor.patternId')}</label>
          <Input
            autoFocus
            value={draft.id}
            onChange={(e) => setDraft({ ...draft, id: e.target.value })}
            placeholder="my_custom_pattern"
            className="h-7 font-mono text-[12px]"
          />
        </div>
        <div className="space-y-1">
          <label className="block font-mono text-[10px] text-foreground/60">{t('regexPatterns.editor.definition')}</label>
          <textarea
            value={draft.def}
            onChange={(e) => setDraft({ ...draft, def: e.target.value })}
            spellCheck={false}
            rows={3}
            placeholder="(?P<name>\\d+)"
            className="w-full resize-none rounded-md border border-input bg-background/40 p-2 font-mono text-[11px] leading-relaxed focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
        </div>
        <div className="space-y-1">
          <label className="block font-mono text-[10px] text-foreground/60">{t('regexPatterns.editor.description')}</label>
          <Input
            value={draft.desc}
            onChange={(e) => setDraft({ ...draft, desc: e.target.value })}
            placeholder={t('regexPatterns.editor.descriptionPlaceholder')}
            className="h-7 text-[12px]"
          />
        </div>
        <div className="flex items-center justify-end gap-2 pt-1">
          <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => setDraft(null)} disabled={saving}>
            {t('regexPatterns.editor.cancel')}
          </Button>
          <Button type="button" size="sm" className="h-7" onClick={() => void create()} disabled={saving}>
            {saving && <Loader2 size={12} className="mr-1.5 animate-spin" />}
            {t('parsingFilters.visual.createInsert')}
          </Button>
        </div>
      </div>
    )
  }

  const filtered = patternOptions
    .filter((p) => p.patternId.toLowerCase().includes(query.toLowerCase()))
    .slice(0, 50)

  return (
    <>
      {onQueryChange && (
        <div className="border-b border-border p-1.5">
          <Input
            autoFocus
            value={query}
            onChange={(e) => onQueryChange(e.target.value)}
            placeholder={t('parsingFilters.visual.patternSearch')}
            className="h-7 text-[12px]"
          />
        </div>
      )}
      <div className="max-h-[220px] overflow-y-auto py-1">
        {filtered.length === 0 ? (
          <div className="px-3 py-4 text-center text-[11px] text-muted-foreground">
            {t('parsingFilters.visual.patternNone')}
          </div>
        ) : (
          filtered.map((p) => (
            <button
              key={p.patternId}
              type="button"
              onClick={() => onPick(p.patternId)}
              className="flex w-full flex-col gap-0.5 px-3 py-1.5 text-left hover:bg-muted"
            >
              <span className="flex items-center gap-1.5">
                <span className="font-mono text-[12px] text-foreground">{`{{.${p.patternId}}}`}</span>
                {p.systemOwner ? (
                  <span className="inline-flex items-center gap-0.5 rounded bg-violet-500/15 px-1 py-0 text-[9px] font-medium text-violet-500">
                    <Lock size={8} />
                    {t('parsingFilters.system')}
                  </span>
                ) : (
                  <span className="rounded bg-sky-500/15 px-1 py-0 text-[9px] font-medium text-sky-500">
                    {t('parsingFilters.user')}
                  </span>
                )}
              </span>
              {p.patternDefinition && (
                <span className="truncate font-mono text-[10px] text-muted-foreground">{p.patternDefinition}</span>
              )}
            </button>
          ))
        )}
      </div>
      {onPatternCreated && (
        <button
          type="button"
          onClick={() => setDraft({ id: query.trim(), def: '', desc: '' })}
          className="flex w-full items-center gap-1.5 border-t border-border px-3 py-2 text-left text-[12px] font-medium text-primary hover:bg-primary/5"
        >
          <Plus size={13} /> {t('parsingFilters.visual.newPattern')}
        </button>
      )}
    </>
  )
}
