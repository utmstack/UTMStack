import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { KeyRound, Loader2, TerminalSquare, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { runAgentCommand } from '../services/agent-console'
import { VariablesManager } from './VariablesManager'
import { ShellCommandInput } from './ShellCommandInput'
import { useAuth } from '@/features/auth'
import { authHttpService } from '@/features/auth/services/auth-http.service'

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
 * output, terminal-style. The command line also enforces a second factor:
 * before each command the prompt asks for the password and proves it via
 * `POST /auth/empty_login` — a token with no claims, unaffected by 2FA. That
 * token rides the WebSocket as `password_token` while the real session JWT
 * stays the `token`. Typed characters render as '*', never in clear.
 */
export function AgentConsole({
  agentId,
  hostname,
  osPlatform,
  offline,
  noRemoteControl,
  onClose,
  variant = 'docked',
}: {
  agentId: string
  hostname: string
  osPlatform?: string
  offline?: boolean
  /** The agent was installed refusing commands. It enforces that itself, so
   *  this only spares the analyst typing something that will come back denied. */
  noRemoteControl?: boolean
  onClose: () => void
  /** 'docked' (default): fixed panel pinned to the bottom of the viewport, like an
   * IDE terminal — used by Data Sources. 'inline': fills its parent container
   * instead, for pages that dedicate a panel to the console (e.g. SOAR). */
  variant?: 'docked' | 'inline'
}) {
  const { t } = useTranslation()
  const { user } = useAuth()
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
  const [authMode, setAuthMode] = useState<'off' | 'password'>('password')
  const [authBusy, setAuthBusy] = useState(false)
  /** Claimless password token from `empty_login`, captured once and reused for
   *  every subsequent command while the console is open. */
  const passwordTokenRef = useRef<string | null>(null)
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

  // Focus the inline prompt — on open and after an authentication round trip
  // (the input remounts between the two modes).
  useEffect(() => {
    inputRef.current?.focus()
    return () => cancelRef.current?.()
  }, [authMode])

  // Re-focus the inline prompt when a command finishes (the input remounts).
  useEffect(() => {
    if (!running) inputRef.current?.focus()
  }, [running])

  const append = (kind: Line['kind'], text: string) => setLines((l) => [...l, { kind, text }])

  // Fire a command with the password token captured at auth time.
  const fireCommand = (command: string) => {
    append('cmd', `${shell}> ${command}`)
    setHistory((h) => [...h, command])
    setHistIdx(-1)
    setRunning(true)
    cancelRef.current = runAgentCommand(
      agentId,
      { command, shell },
      {
        onOutput: (d) => {
          append('out', d)
          setRunning(false)
        },
        onError: (m) => append('err', m),
        onDone: () => {
          setRunning(false)
          cancelRef.current = null
          inputRef.current?.focus()
        },
      },
      passwordTokenRef.current ?? undefined,
    )
  }

  // Prove the password via empty_login (no claims, no 2FA) ONCE; the token is
  // stored and reused for every command. During this mode 'input' holds the secret.
  const submitAuth = async () => {
    const secret = input
    if (!secret.trim() || authBusy) return
    setAuthBusy(true)
    try {
      const res = await authHttpService.emptyLogin({ login: user?.email ?? '', password: secret })
      passwordTokenRef.current = res.token
      setAuthMode('off')
      setInput('')
    } catch (err) {
      append('err', err instanceof Error ? err.message : t('datasources.console.password.authError'))
      setInput('')
    } finally {
      setAuthBusy(false)
    }
  }

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
    // While the prompt asks for the credential, Enter submits it, not a command.
    if (authMode !== 'off') {
      void submitAuth()
      return
    }
    const command = input.trim()
    if (!command || running || offline || noRemoteControl) return
    setInput('')
    fireCommand(command)
  }

  const onKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      run()
    } else if (authMode === 'off') {
      if (e.key === 'ArrowUp') {
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
  }

  const authing = authMode !== 'off'

  return (
    <>
    <div
      className={cn(
        'flex flex-col overflow-hidden shadow-2xl bg-card text-card-foreground',
        embedded
          ? 'h-full w-full rounded-xl border border-border'
          : 'fixed inset-x-0 bottom-0 z-[60] border-t border-border',
      )}
      style={embedded ? undefined : { height }}
    >
      {/* Drag handle to resize, like an IDE terminal — docked mode only. */}
      {!embedded && (
        <div onMouseDown={startResize} className="group absolute inset-x-0 top-0 z-10 h-2 cursor-row-resize">
          <div className="mx-auto mt-0.5 h-0.5 w-10 rounded-full bg-muted-foreground/60 transition-colors group-hover:bg-primary" />
        </div>
      )}

      <header className="flex items-center justify-between border-b border-border px-4 py-2">
        <div className="flex min-w-0 items-center gap-2 text-sm font-medium text-foreground">
          <TerminalSquare size={16} className="shrink-0 text-emerald-500" />
          <span className="truncate">{hostname}</span>
          {windows ? (
            <select
              value={winShell}
              onChange={(e) => setWinShell(e.target.value as WinShell)}
              title={t('datasources.console.shellSelect')}
              className="h-7 shrink-0 rounded-md border border-border bg-background px-1.5 text-[11px] text-foreground transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            >
              <option value="cmd">cmd</option>
              <option value="powershell">powershell</option>
            </select>
          ) : (
            <span className="shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-muted-foreground">{shell}</span>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <button
            onClick={() => setVarsOpen(true)}
            className="flex h-7 items-center gap-1.5 rounded px-2 text-xs text-muted-foreground hover:bg-accent hover:text-foreground"
          >
            <KeyRound size={13} /> {t('datasources.console.vars.button')}
          </button>
          <button
            onClick={onClose}
            aria-label={t('datasources.console.close')}
            className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground"
          >
            <X size={16} />
          </button>
        </div>
      </header>

        {noRemoteControl ? (
          <div className="border-b border-red-500/20 bg-red-500/10 px-4 py-2 text-[11px] text-red-300">
            {t('datasources.console.noRemoteControl')}
          </div>
        ) : (
          offline && (
            <div className="border-b border-amber-500/20 bg-amber-500/10 px-4 py-2 text-[11px] text-amber-300">
              {t('datasources.console.offline')}
            </div>
          )
        )}

        <div
          ref={scrollRef}
          onClick={() => inputRef.current?.focus()}
          className="flex-1 cursor-text overflow-y-auto px-4 py-3 font-mono text-[12px] leading-relaxed text-foreground/80"
        >
          {lines.length === 0 && !authing && <p className="mb-1 text-muted-foreground">{t('datasources.console.hint')}</p>}
          {authing && lines.length === 0 && (
            <p className="mb-1 text-muted-foreground">{t('datasources.console.password.authHint')}</p>
          )}
          {lines.map((l, i) => (
            <pre
              key={i}
              className={cn(
                'whitespace-pre-wrap [overflow-wrap:anywhere]',
                l.kind === 'cmd' && 'mt-2 text-emerald-500 first:mt-0',
                l.kind === 'err' && 'text-destructive',
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
            // like a real terminal — not in a separate bar.
            authing ? (
              <div className="mt-1 flex items-center gap-2">
                <span className="shrink-0 text-amber-400">pwd&gt;</span>
                <div className="relative min-w-0 flex-1">
                  <div aria-hidden className="whitespace-pre overflow-hidden font-mono text-[12px] leading-relaxed text-zinc-100">
                    {input.length ? '•'.repeat(input.length) : ' '}
                  </div>
                  <input
                    ref={inputRef}
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={onKey}
                    spellCheck={false}
                    autoComplete="off"
                    aria-label={t('datasources.console.password.title')}
                    className="absolute inset-0 w-full bg-transparent font-mono text-[12px] leading-relaxed text-transparent caret-zinc-100 outline-none"
                  />
                </div>
                {authBusy && <Loader2 size={12} className="shrink-0 animate-spin text-zinc-500" />}
                <span className="shrink-0 text-[11px] text-muted-foreground">{t('datasources.console.password.title')}</span>
              </div>
            ) : (
              <div className={cn('mt-1 flex items-center gap-2', (offline || noRemoteControl) && 'opacity-50')}>
                <span className="shrink-0 text-emerald-400">{shell}&gt;</span>
                <ShellCommandInput
                  ref={inputRef}
                  value={input}
                  shell={shell}
                  onChange={setInput}
                  onKeyDown={onKey}
                  disabled={offline || noRemoteControl}
                  placeholder={
                    noRemoteControl
                      ? t('datasources.console.noRemoteControlShort')
                      : offline
                        ? t('datasources.console.disabled')
                        : t('datasources.console.placeholder')
                  }
                />
              </div>
            )
          )}
        </div>
    </div>
    {varsOpen && <VariablesManager onClose={() => setVarsOpen(false)} onInsert={insertToken} />}
    </>
  )
}
