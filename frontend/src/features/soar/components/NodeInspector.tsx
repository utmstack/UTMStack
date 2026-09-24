import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Save, Trash2, Zap } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import type { FlowNode } from '../types/soar.types'
import { COMMAND_TEMPLATES, shellKindFor } from '../lib/command-templates'
import { mailRecipientError } from '../lib/mail-node-validity'
import { AgentPicker } from './AgentPicker'
import { ConditionalParamsEditor } from './ConditionalParamsEditor'
import { HttpParamsEditor } from './HttpParamsEditor'
import { IncidentParamsEditor } from './IncidentParamsEditor'
import { InsertFieldMenu } from './InsertFieldMenu'
import { MailParamsEditor } from './MailParamsEditor'
import { LLMParamsEditor } from './LLMParamsEditor'
import { NotifyParamsEditor } from './NotifyParamsEditor'

interface Props {
  nodeId: string
  node: FlowNode
  /** Full node map — used by InsertFieldMenu to enumerate enrichment
   *  ancestors reachable from this node. */
  nodes: Record<string, FlowNode>
  readOnly?: boolean
  onSave: (nextId: string, patch: Partial<FlowNode>) => boolean
  onDelete: () => void
  onClose: () => void
}

/** Right-side properties panel for the selected node: id, kind, executor,
 *  command/params (schema depends on executor), on_success/on_error left
 *  implicit (drawn on the canvas). */
export function NodeInspector({ nodeId, node, nodes, readOnly, onSave, onDelete, onClose }: Props) {
  const { t } = useTranslation()
  const [localId, setLocalId] = useState(nodeId)
  const [draftNode, setDraftNode] = useState<FlowNode>(node)
  const [paramsText, setParamsText] = useState(() => (node.params ? JSON.stringify(node.params, null, 2) : ''))
  const [paramsError, setParamsError] = useState<string | null>(null)
  const commandRef = useRef<HTMLTextAreaElement>(null)
  const paramsRef = useRef<HTMLTextAreaElement>(null)
  const [templatesOpen, setTemplatesOpen] = useState(false)
  const [width, setWidth] = useState(384)
  const mailInvalid =
    draftNode.executor === 'mail' &&
    Boolean(mailRecipientError((draftNode.params as { to?: string; cc?: string } | undefined) ?? {}))

  const startResize = (e: React.MouseEvent) => {
    e.preventDefault()
    const startX = e.clientX
    const startW = width
    const onMove = (ev: MouseEvent) => {
      const next = startW + (startX - ev.clientX)
      setWidth(Math.max(280, Math.min(next, Math.min(900, window.innerWidth - 200))))
    }
    const onUp = () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  }

  useEffect(() => {
    setLocalId(nodeId)
    setDraftNode(node)
  }, [nodeId])
  useEffect(() => {
    setParamsText(node.params ? JSON.stringify(node.params, null, 2) : '')
    setParamsError(null)
  }, [node.params, nodeId])

  const commitId = () => {
    const trimmed = localId.trim()
    if (trimmed) setLocalId(trimmed)
    else setLocalId(nodeId)
  }

  const updateDraft = (patch: Partial<FlowNode>) => setDraftNode((current) => ({ ...current, ...patch }))

  const save = () => {
    if (mailInvalid) return
    const trimmedId = localId.trim()
    if (!trimmedId) {
      setLocalId(nodeId)
      return
    }
    if (onSave(trimmedId, draftNode)) onClose()
  }

  const commitParams = () => {
    if (!paramsText.trim()) {
      updateDraft({ params: undefined })
      setParamsError(null)
      return
    }
    try {
      const parsed = JSON.parse(paramsText)
      updateDraft({ params: parsed })
      setParamsError(null)
    } catch (e) {
      setParamsError(e instanceof Error ? e.message : t('soar.editor.canvas.invalidJson'))
    }
  }

  // Insert into the shell command textarea at the caret. Templates replace-all
  // when the field is empty and insert-at-caret otherwise, so users can build
  // a chain of them.
  const insertIntoCommand = (token: string, replaceAll: boolean) => {
    const cur = draftNode.command ?? ''
    if (replaceAll && !cur.trim()) {
      updateDraft({ command: token })
      requestAnimationFrame(() => {
        const el = commandRef.current
        if (el) {
          el.focus()
          el.setSelectionRange(token.length, token.length)
        }
      })
      return
    }
    const el = commandRef.current
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    updateDraft({ command: next })
    requestAnimationFrame(() => {
      const el2 = commandRef.current
      if (el2) {
        el2.focus()
        const pos = start + token.length
        el2.setSelectionRange(pos, pos)
      }
    })
  }

  // Insert into the non-shell params textarea at the caret. Works against the
  // local paramsText state (not committed yet) so users can keep composing.
  const insertIntoParams = (token: string) => {
    const el = paramsRef.current
    const cur = paramsText
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    setParamsText(next)
    requestAnimationFrame(() => {
      const el2 = paramsRef.current
      if (el2) {
        el2.focus()
        const pos = start + token.length
        el2.setSelectionRange(pos, pos)
      }
    })
  }

  const shellKind = shellKindFor(draftNode.platform ?? '', draftNode.shell ?? '')

  return (
    <aside
      className="relative flex shrink-0 flex-col border-l border-border bg-card h-[100%] overflow-y-auto"
      style={{ width }}
    >
      <div
        onMouseDown={startResize}
        className="absolute left-0 top-0 z-10 h-full w-1 cursor-col-resize hover:bg-primary/40"
        title={t('soar.editor.canvas.dragToResize')}
      />
      <div className="flex items-center justify-between border-b border-border py-2 pl-[30px] pr-3">
        <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">{t('soar.editor.canvas.node')}</div>
        {!readOnly && (
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={save}
              disabled={mailInvalid}
              className="rounded p-1 text-muted-foreground hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"
              title={t('soar.editor.save')}
            >
              <Save size={13} />
            </button>
            <button onClick={onDelete} className="rounded p-1 text-muted-foreground hover:text-red-500" title={t('soar.editor.canvas.deleteNode')}>
              <Trash2 size={13} />
            </button>
          </div>
        )}
      </div>
      <div className="flex-1 space-y-3 overflow-y-auto p-3 text-xs">
        <Field label={t('soar.editor.canvas.id')}>
          <Input
            value={localId}
            readOnly={readOnly}
            onChange={(e) => setLocalId(e.target.value)}
            onBlur={commitId}
            className="h-8 font-mono"
          />
        </Field>
        {draftNode.executor === 'shell' && (
          <>
            <Field label={t('soar.editor.canvas.commandLabel')}>
              {!readOnly && (
                <div className="mb-1 flex flex-wrap items-center gap-1.5">
                  <TemplatesPopover
                    open={templatesOpen}
                    onOpenChange={setTemplatesOpen}
                    onPick={(cmd) => insertIntoCommand(cmd, true)}
                    shellKind={shellKind}
                  />
                  <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={(token) => insertIntoCommand(token, false)} />
                </div>
              )}
              <textarea
                ref={commandRef}
                value={draftNode.command ?? ''}
                readOnly={readOnly}
                onChange={(e) => updateDraft({ command: e.target.value })}
                rows={4}
                className="w-full rounded-md border border-input bg-background px-2 py-1.5 font-mono text-[11px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                placeholder='usermod -s /sbin/nologin $(alert.target.user)'
              />
            </Field>
            <AgentPicker
              platform={draftNode.platform}
              agent={draftNode.agent}
              excludedAgents={draftNode.excludedAgents}
              shell={draftNode.shell}
              readOnly={readOnly}
              onChange={updateDraft}
            />
          </>
        )}

        {draftNode.executor === 'conditional' && (
          <Field label={t('soar.editor.conditions')}>
            <ConditionalParamsEditor
              nodeId={nodeId}
              nodes={nodes}
              params={draftNode.params}
              readOnly={readOnly}
              onChange={(next) => updateDraft({ params: next })}
            />
          </Field>
        )}

        {draftNode.executor === 'http' && (
          <HttpParamsEditor
            nodeId={nodeId}
            nodes={nodes}
            params={draftNode.params}
            readOnly={readOnly}
            onChange={(next) => updateDraft({ params: next })}
          />
        )}

        {draftNode.executor === 'incident' && (
          <IncidentParamsEditor
            nodeId={nodeId}
            nodes={nodes}
            params={draftNode.params}
            readOnly={readOnly}
            onChange={(next) => updateDraft({ params: next })}
          />
        )}

        {draftNode.executor === 'mail' && (
          <MailParamsEditor
            nodeId={nodeId}
            nodes={nodes}
            params={draftNode.params}
            readOnly={readOnly}
            onChange={(next) => updateDraft({ params: next })}
          />
        )}

        {draftNode.executor === 'notify' && (
          <NotifyParamsEditor
            nodeId={nodeId}
            nodes={nodes}
            params={draftNode.params}
            readOnly={readOnly}
            onChange={(next) => updateDraft({ params: next })}
          />
        )}

        {(draftNode.executor === 'llm_enrich' || draftNode.executor === 'llm_action') && (
          <LLMParamsEditor
            nodeId={nodeId}
            nodes={nodes}
            params={draftNode.params}
            readOnly={readOnly}
            executor={draftNode.executor}
            onChange={(next) => updateDraft({ params: next })}
          />
        )}

        {draftNode.executor !== 'shell' && draftNode.executor !== 'conditional' && draftNode.executor !== 'http' && draftNode.executor !== 'incident' && draftNode.executor !== 'mail' && draftNode.executor !== 'notify' && draftNode.executor !== 'llm_enrich' && draftNode.executor !== 'llm_action' && (
          <Field label={t('soar.editor.canvas.paramsJson')}>
            {!readOnly && (
              <div className="mb-1 flex flex-wrap items-center gap-1.5">
                <InsertFieldMenu nodes={nodes} currentNodeId={nodeId} onInsert={(token) => insertIntoParams(token)} />
              </div>
            )}
            <textarea
              ref={paramsRef}
              value={paramsText}
              readOnly={readOnly}
              onChange={(e) => setParamsText(e.target.value)}
              onBlur={commitParams}
              rows={8}
              className="w-full rounded-md border border-input bg-background px-2 py-1.5 font-mono text-[11px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              placeholder='{"url":"https://example.com"}'
            />
            {paramsError && <p className="mt-1 text-[10px] text-red-500">{paramsError}</p>}
          </Field>
        )}

        <div className="rounded-md bg-muted/40 p-2 text-[10px] text-muted-foreground">
          {t('soar.editor.canvas.handlesHint')}
        </div>
      </div>
      <div className="border-t border-border p-2">
        <div className="flex gap-2">
          <Button size="sm" variant="outline" className="flex-1" onClick={onClose}>
            {t('soar.editor.cancel')}
          </Button>
          {!readOnly && (
            <Button size="sm" className="flex-1" onClick={save} disabled={mailInvalid}>
              {t('soar.editor.save')}
            </Button>
          )}
        </div>
        <Button size="sm" variant="outline" className="mt-2 w-full" onClick={onDelete} disabled={readOnly}>
          {t('soar.editor.canvas.deleteNode')}
        </Button>
      </div>
    </aside>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{label}</label>
      {children}
    </div>
  )
}

function TemplatesPopover({
  open,
  onOpenChange,
  onPick,
  shellKind,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  onPick: (command: string) => void
  shellKind: ReturnType<typeof shellKindFor>
}) {
  const { t } = useTranslation()
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onOpenChange(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open, onOpenChange])
  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => onOpenChange(!open)}
        className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 text-[10px] hover:bg-muted"
      >
        <Zap size={11} /> {t('soar.editor.canvas.templatesLabel')} <ChevronDown size={10} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 max-h-64 w-64 overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-lg">
          {COMMAND_TEMPLATES.map((tpl) => (
            <button
              key={tpl.id}
              type="button"
              onClick={() => {
                onPick(tpl.command(shellKind))
                onOpenChange(false)
              }}
              className="flex w-full items-center px-3 py-1.5 text-left text-[11px] hover:bg-muted"
            >
              {tpl.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
