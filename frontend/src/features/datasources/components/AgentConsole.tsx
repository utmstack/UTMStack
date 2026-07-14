import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { KeyRound, Loader2, TerminalSquare, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { runAgentCommand } from '../services/agent-console'
import { VariablesManager } from './VariablesManager'
import { ShellCommandInput } from './ShellCommandInput'

const clamp = (n: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, n))

/** Windows agents let the analyst choose between cmd/powershell; the rest use bash. */
function isWindowsOs(osPlatform?: string): boolean {
  return (osPlatform ?? '').toLowerCase().includes('win')
}

type WinShell = 'cmd' | 'powershell'

interface Line {
  kind: 'cmd' | 'out' | 'err'
  text: string
}

/**
 * Interactive console (live response) for an agent datasource. Each command opens
 * a fresh WebSocket to the backend agent command stream and renders the streamed
 * output, terminal-style.
 */
export function AgentConsole({
  agentId,
  hostname,
  osPlatform,
  offline,
  onClose,
  variant = 'docked',
}: {
  agentId: string
  hostname: string
  osPlatform?: string
  offline?: boolean
  onClose: () => void
  /** 'docked' (default): fixed panel pinned to the bottom of the viewport, like an
   * IDE terminal — used by Data Sources. 'inline': fills its parent container
   * instead, for pages that dedicate a panel to the console (e.g. SOAR). */
  variant?: 'docked' | 'inline'
}) {
  const { t } = useTranslation()
  const embedded = variant === 'inline'
  // Windows → analyst picks cmd/powershell (default cmd); Linux/other → bash fixed.
  const windows = isWindowsOs(osPlatform)
  const [winShell, setWinShell] = useState<WinShell>('cmd')
  const shell = windows ? winShell : 'bash'
  const [lines, setLines] = useState<Line[]>([])
  const [input, setInput] = useState('')
  const [running, setRunning] = useState(false)
  const [history, setHistory] = useState<string[]>([])
  const [histIdx, setHistIdx] = useState(-1)
  const [varsOpen, setVarsOpen] = useState(false)
  const [height, setHeight] = useState(() => clamp(Math.round(window.innerHeight * 0.42), 200, window.innerHeight - 80))
  const cancelRef = useRef<(() => void) | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Drag the top edge to resize the docked panel, like an IDE terminal.
  const startResize = (e: React.MouseEvent) => {
    e.preventDefault()
    const onMove = (ev: MouseEvent) => setHeight(clamp(window.innerHeight - ev.clientY, 160, window.innerHeight - 60))
    const onUp = () => {
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
    }
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
  }

  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [lines, running])

  useEffect(() => {
    inputRef.current?.focus()
    return () => cancelRef.current?.()
  }, [])

  // Re-focus the inline prompt when a command finishes (the input remounts).
  useEffect(() => {
    if (!running) inputRef.current?.focus()
  }, [running])

  const append = (kind: Line['kind'], text: string) => setLines((l) => [...l, { kind, text }])

  // Insert a `$[variables.NAME]` reference at the caret in the command input.
  const insertToken = (variableName: string) => {
    const token = `$[variables.${variableName}]`
    const el = inputRef.current
    if (el) {
      const start = el.selectionStart ?? input.length
      const end = el.selectionEnd ?? input.length
      setInput(input.slice(0, start) + token + input.slice(end))
      requestAnimationFrame(() => {
        el.focus()
        const pos = start + token.length
        el.setSelectionRange(pos, pos)
      })
    } else {
      setInput((v) => v + token)
    }
    setVarsOpen(false)
  }

  const run = () => {
    const command = input.trim()
    if (!command || running || offline) return
    append('cmd', `${shell}> ${command}`)
    setHistory((h) => [...h, command])
    setHistIdx(-1)
    setInput('')
    setRunning(true)
    cancelRef.current = runAgentCommand(
      agentId,
      { command, shell },
      {
        onOutput: (d) =>{
          append('out', d);
          setRunning(false);
        },
        onError: (m) => append('err', m),
        onDone: () => {
          setRunning(false)
          cancelRef.current = null
          inputRef.current?.focus()
        },
      },
    )
  }

  const onKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      run()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (!history.length) return
      const idx = histIdx < 0 ? history.length - 1 : Math.max(0, histIdx - 1)
      setHistIdx(idx)
      setInput(history[idx])
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (histIdx < 0) return
      const idx = histIdx + 1
      if (idx >= history.length) {
        setHistIdx(-1)
        setInput('')
      } else {
        setHistIdx(idx)
        setInput(history[idx])
      }
    }
  }

  return (
    <>
    <div
      className={cn(
        'flex flex-col overflow-hidden shadow-2xl',
        embedded
          ? 'h-full w-full rounded-xl border border-white/10 bg-[#0b0d11]'
          : 'fixed inset-x-0 bottom-0 z-[60] border-t border-white/10 bg-[#0b0d11]',
      )}
      style={embedded ? undefined : { height }}
    >
      {/* Drag handle to resize, like an IDE terminal — docked mode only. */}
      {!embedded && (
        <div onMouseDown={startResize} className="group absolute inset-x-0 top-0 z-10 h-2 cursor-row-resize">
          <div className="mx-auto mt-0.5 h-0.5 w-10 rounded-full bg-zinc-700 transition-colors group-hover:bg-primary" />
        </div>
      )}

      <header className="flex items-center justify-between border-b border-white/10 px-4 py-2">
        <div className="flex min-w-0 items-center gap-2 text-sm font-medium text-zinc-200">
          <TerminalSquare size={16} className="shrink-0 text-emerald-400" />
          <span className="truncate">{hostname}</span>
          {windows ? (
            <select
              value={winShell}
              onChange={(e) => setWinShell(e.target.value as WinShell)}
              title={t('datasources.console.shellSelect')}
              className="shrink-0 rounded border border-white/10 bg-white/5 px-1.5 py-0.5 text-[11px] text-zinc-300 outline-none focus:border-primary"
            >
              <option value="cmd">cmd</option>
              <option value="powershell">powershell</option>
            </select>
          ) : (
            <span className="shrink-0 rounded bg-white/10 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-zinc-400">{shell}</span>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <button
            onClick={() => setVarsOpen(true)}
            className="flex h-7 items-center gap-1.5 rounded px-2 text-xs text-zinc-400 hover:bg-white/10 hover:text-zinc-100"
          >
            <KeyRound size={13} /> {t('datasources.console.vars.button')}
          </button>
          <button
            onClick={onClose}
            aria-label={t('datasources.console.close')}
            className="flex h-7 w-7 items-center justify-center rounded text-zinc-400 hover:bg-white/10 hover:text-zinc-100"
          >
            <X size={16} />
          </button>
        </div>
      </header>

        {offline && (
          <div className="border-b border-amber-500/20 bg-amber-500/10 px-4 py-2 text-[11px] text-amber-300">
            {t('datasources.console.offline')}
          </div>
        )}

        <div
          ref={scrollRef}
          onClick={() => inputRef.current?.focus()}
          className="flex-1 cursor-text overflow-y-auto px-4 py-3 font-mono text-[12px] leading-relaxed text-zinc-300"
        >
          {lines.length === 0 && <p className="mb-1 text-zinc-500">{t('datasources.console.hint')}</p>}
          {lines.map((l, i) => (
            <pre
              key={i}
              className={cn(
                'whitespace-pre-wrap [overflow-wrap:anywhere]',
                l.kind === 'cmd' && 'mt-2 text-emerald-400 first:mt-0',
                l.kind === 'err' && 'text-red-400',
              )}
            >
              {l.text}
            </pre>
          ))}
          {running ? (
            <div className="mt-1 flex items-center gap-2 text-zinc-500">
              <Loader2 size={12} className="animate-spin" /> {t('datasources.console.running')}
            </div>
          ) : (
            // The active prompt + input live inline at the bottom of the output,
            // like a real terminal — not in a separate bar. Offline → disabled.
            <div className={cn('mt-1 flex items-center gap-2', offline && 'opacity-50')}>
              <span className="shrink-0 text-emerald-400">{shell}&gt;</span>
              <ShellCommandInput
                ref={inputRef}
                value={input}
                shell={shell}
                onChange={setInput}
                onKeyDown={onKey}
                disabled={offline}
                placeholder={offline ? t('datasources.console.disabled') : t('datasources.console.placeholder')}
              />
            </div>
          )}
        </div>
    </div>
    {varsOpen && <VariablesManager onClose={() => setVarsOpen(false)} onInsert={insertToken} />}
    </>
  )
}
