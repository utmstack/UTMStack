import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Code2, FlaskConical, LayoutList, Loader2, Lock, Trash2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import {
  DataProcessingHttpError,
  filtersHttpService,
} from '@/features/data-processing/services/data-processing-http.service'
import type { DataTypeOption, Filter } from '@/features/data-processing/types/data-processing.types'
import { regexPatternsHttpService, type RegexPattern } from '@/features/regex-patterns/services/regex-patterns-http.service'
import { TestPlaygroundModal } from '@/features/playground/components/TestPlaygroundModal'
import { displayName, emptyModel, parseFilter, sanitizeFileName, serializeFilter, type FilterModel } from '../lib/filter-model'
import { VisualFilterEditor } from './VisualFilterEditor'
import { PatternInsertButton } from './PatternInsertButton'
import { DataTypeMultiSelect } from './DataTypeMultiSelect'

interface Props {
  filter: Filter
  creating: boolean
  onClose: () => void
  onSaved: () => void
}

export function FilterFormDrawer({ filter, creating, onClose, onSaved }: Props) {
  const { t } = useTranslation()
  const readOnly = filter.system

  // For a brand-new filter, the user picks a friendly name — it's sanitized
  // into a flat (no subpath) yaml filename, which becomes the relPath sent on create.
  const [name, setName] = useState('')
  const newRelPath = useMemo(() => sanitizeFileName(name), [name])
  // A brand-new filter starts from a real (parseable) template — seeded with
  // the order the caller already computed (one past the current max) — so
  // the Visual tab is reachable immediately and that order actually survives
  // into the saved YAML, instead of only living in a placeholder hint.
  const [content, setContent] = useState(() =>
    creating ? serializeFilter(emptyModel([], filter.order || 100)) : filter.content,
  )
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [showTestModal, setShowTestModal] = useState(false)

  // Visual ↔ Code. Content (YAML string) stays canonical; the visual editor
  // serializes back into it on every change.
  const [mode, setMode] = useState<'visual' | 'code'>('code')
  const [model, setModel] = useState<FilterModel>({ dataTypes: [], order: filter.order || 100, steps: [] })
  const [dataTypeOptions, setDataTypeOptions] = useState<DataTypeOption[]>([])
  const [patternOptions, setPatternOptions] = useState<RegexPattern[]>([])
  const yamlTaRef = useRef<HTMLTextAreaElement | null>(null)

  // A pattern created on the fly is added to the catalog so it's immediately referenceable.
  const addPattern = (p: RegexPattern) =>
    setPatternOptions((prev) => (prev.some((x) => x.patternId === p.patternId) ? prev : [p, ...prev]))

  // Catalog of known dataTypes to pick from in the visual editor.
  useEffect(() => {
    filtersHttpService
      .dataTypeCatalog()
      .then((d) => setDataTypeOptions(d ?? []))
      .catch(() => setDataTypeOptions([]))
  }, [])

  // Catalog of regex patterns to reference from grok steps (`{{.name}}`).
  useEffect(() => {
    regexPatternsHttpService
      .list({ size: 200 })
      .then((r) => setPatternOptions(r.data ?? []))
      .catch(() => setPatternOptions([]))
  }, [])

  // Default to the visual editor when the content parses cleanly.
  useEffect(() => {
    const r = parseFilter(creating ? content : filter.content)
    if (r.ok) {
      setModel(r.model)
      setMode('visual')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter.relPath])

  const canVisual = useMemo(() => parseFilter(content).ok, [content])

  const toVisual = () => {
    const r = parseFilter(content)
    if (!r.ok) {
      toast.error(t('parsingFilters.visual.cantParse', { error: r.error }))
      return
    }
    setModel(r.model)
    setMode('visual')
  }

  const onModelChange = (m: FilterModel) => {
    setModel(m)
    setContent(serializeFilter(m))
  }

  const dirty = creating ? name.trim() !== '' || content.trim() !== '' : content !== filter.content

  const save = async () => {
    if (creating && !newRelPath) {
      toast.error(t('parsingFilters.toast.nameRequired'))
      return
    }
    if (!content.trim()) {
      toast.error(t('parsingFilters.toast.contentRequired'))
      return
    }
    setSaving(true)
    try {
      if (creating) await filtersHttpService.create({ relPath: newRelPath, content })
      else await filtersHttpService.update({ relPath: filter.relPath, content })
      toast.success(t('parsingFilters.toast.saved'))
      onSaved()
    } catch (err) {
      toast.error(err instanceof DataProcessingHttpError ? err.message : t('parsingFilters.toast.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const remove = async () => {
    setSaving(true)
    try {
      await filtersHttpService.remove(filter.relPath)
      toast.success(t('parsingFilters.toast.deleted'))
      onSaved()
    } catch {
      toast.error(t('parsingFilters.toast.deleteError'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="flex items-center gap-2 truncate text-lg font-semibold">
              {creating ? t('parsingFilters.editor.titleNew') : displayName(filter.relPath)}
              {readOnly && (
                <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                  <Lock size={9} /> {t('parsingFilters.system')}
                </span>
              )}
            </h2>
          </div>
          <button onClick={onClose} className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden p-6">
          {readOnly && (
            <div className="flex shrink-0 items-center gap-2 rounded-md bg-violet-500/10 px-3 py-2 text-xs text-violet-600 dark:text-violet-300">
              <Lock size={13} /> {t('parsingFilters.editor.readOnly')}
            </div>
          )}
          {creating && (
            <div className="shrink-0 space-y-1.5">
              <label className="block text-xs font-medium text-foreground/80">{t('parsingFilters.editor.name')}</label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t('parsingFilters.editor.namePlaceholder')} />
              <p className="text-[11px] text-muted-foreground">
                {newRelPath ? t('parsingFilters.editor.nameHint', { file: newRelPath }) : t('parsingFilters.editor.namePending')}
              </p>
            </div>
          )}

          <div className="flex min-h-0 flex-1 flex-col gap-1.5">
            <div className="flex shrink-0 items-center justify-between">
              <label className="block text-xs font-medium text-foreground/80">{t('parsingFilters.editor.content')}</label>
              <div className="inline-flex rounded-md border border-border p-0.5">
                <button
                  type="button"
                  onClick={toVisual}
                  disabled={!canVisual && mode === 'code'}
                  title={!canVisual ? t('parsingFilters.visual.codeOnly') : undefined}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors disabled:opacity-40',
                    mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <LayoutList size={13} /> {t('parsingFilters.visual.visualTab')}
                </button>
                <button
                  type="button"
                  onClick={() => setMode('code')}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors',
                    mode === 'code' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <Code2 size={13} /> {t('parsingFilters.visual.codeTab')}
                </button>
              </div>
            </div>
            {mode === 'visual' ? (
              <div className="min-h-0 flex-1 space-y-4 overflow-y-auto rounded-md border border-input bg-background/20 p-3">
                <div className="space-y-1.5">
                  <label className="block text-xs font-medium text-foreground/80">{t('parsingFilters.visual.dataTypes')}</label>
                  <DataTypeMultiSelect
                    values={model.dataTypes}
                    options={dataTypeOptions}
                    readOnly={readOnly}
                    onChange={(dataTypes) => onModelChange({ ...model, dataTypes })}
                  />
                </div>
                <div className="space-y-2">
                  <label className="block text-xs font-medium text-foreground/80">{t('parsingFilters.visual.steps')}</label>
                  <VisualFilterEditor
                    value={model.steps}
                    readOnly={readOnly}
                    patternOptions={patternOptions}
                    onPatternCreated={addPattern}
                    onChange={(steps) => onModelChange({ ...model, steps })}
                  />
                </div>
              </div>
            ) : (
              <div className="relative flex min-h-0 flex-1 flex-col">
                {!readOnly && (
                  <div className="absolute right-2 top-2 z-10">
                    <PatternInsertButton
                      taRef={yamlTaRef}
                      value={content}
                      onChange={setContent}
                      patternOptions={patternOptions}
                      onPatternCreated={addPattern}
                    />
                  </div>
                )}
                <YamlCodeEditor
                  value={content}
                  onChange={setContent}
                  readOnly={readOnly}
                  textareaRef={yamlTaRef}
                  placeholder={'pipeline:\n  - dataTypes: [ ... ]\n    order: 100\n    steps: [ ... ]'}
                />
              </div>
            )}
            <p className="shrink-0 text-[11px] text-muted-foreground">{t('parsingFilters.editor.contentHint')}</p>
          </div>
        </div>

        <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
          <div>
            {!creating && !readOnly &&
              (confirmDelete ? (
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="destructive" onClick={() => void remove()} disabled={saving}>
                    <Trash2 size={13} className="mr-1.5" /> {t('parsingFilters.editor.confirmDelete')}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)} disabled={saving}>
                    {t('parsingFilters.editor.cancel')}
                  </Button>
                </div>
              ) : (
                <button
                  onClick={() => setConfirmDelete(true)}
                  className="rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300"
                >
                  {t('parsingFilters.editor.delete')}
                </button>
              ))}
          </div>
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => setShowTestModal(true)}>
              <FlaskConical size={13} className="mr-1.5" /> {t('parsingFilters.editor.test')}
            </Button>
            {!readOnly && (
              <Button size="sm" disabled={!dirty || saving} onClick={() => void save()}>
                {saving ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null}
                {saving ? t('parsingFilters.editor.saving') : t('parsingFilters.editor.save')}
              </Button>
            )}
          </div>
        </footer>
      </div>

      {showTestModal && (
        <TestPlaygroundModal
          mode="filter"
          dataTypeOptions={model.dataTypes}
          draftContent={content}
          onClose={() => setShowTestModal(false)}
        />
      )}
    </div>
  )
}
