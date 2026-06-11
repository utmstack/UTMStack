import { forwardRef, useRef, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/shared/lib/utils'

interface Props {
  value: string
  shell: string
  disabled?: boolean
  placeholder?: string
  onChange: (value: string) => void
  onKeyDown?: (e: KeyboardEvent<HTMLInputElement>) => void
}

/**
 * Single-line command input with shell-aware syntax highlighting. A colored
 * layer renders behind a transparent <input> (which keeps caret, selection and
 * key handling), and the two scroll horizontally in sync.
 */
export const ShellCommandInput = forwardRef<HTMLInputElement, Props>(function ShellCommandInput(
  { value, shell, disabled, placeholder, onChange, onKeyDown },
  ref,
) {
  const overlayRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement | null>(null)

  const setRefs = (el: HTMLInputElement | null) => {
    inputRef.current = el
    if (typeof ref === 'function') ref(el)
    else if (ref) ref.current = el
  }

  const sync = () => {
    if (overlayRef.current && inputRef.current) {
      overlayRef.current.scrollLeft = inputRef.current.scrollLeft
    }
  }

  const shared = 'whitespace-pre font-mono text-[12px] leading-relaxed'
  const nodes = highlightCommand(value, shell)

  return (
    <div className="relative min-w-0 flex-1">
      <div ref={overlayRef} aria-hidden className={cn(shared, 'pointer-events-none overflow-hidden text-zinc-100')}>
        {nodes.length ? nodes : ' '}
      </div>
      <input
        ref={setRefs}
        value={value}
        onChange={(e) => {
          onChange(e.target.value)
          requestAnimationFrame(sync)
        }}
        onKeyDown={onKeyDown}
        onScroll={sync}
        disabled={disabled}
        placeholder={placeholder}
        spellCheck={false}
        autoComplete="off"
        className={cn(
          shared,
          'absolute inset-0 w-full bg-transparent text-transparent caret-zinc-100 outline-none placeholder:text-zinc-600 disabled:cursor-not-allowed',
        )}
      />
    </div>
  )
})

/* ── Shell-aware highlighting ─────────────────────────────────────────────── */

const COL = {
  cmd: 'text-emerald-300', // command / cmdlet
  flag: 'text-sky-300', // -flag / --long / /flag
  vari: 'text-amber-300', // $var ${..} %VAR%
  utm: 'text-fuchsia-300', // $[variables.NAME] (UTMStack token)
  str: 'text-orange-300', // "string" 'string'
  op: 'text-zinc-500', // | & ; < > pipes/redirects
  comment: 'text-zinc-500 italic',
}

interface Tok {
  text: string
  cls?: string
}

function highlightCommand(cmd: string, shell: string): ReactNode[] {
  const isPwsh = shell === 'powershell'
  const isCmd = shell === 'cmd'
  const toks: Tok[] = []
  let i = 0
  let expectCmd = true // first word of the line / after a pipe is a command
  const push = (text: string, cls?: string) => toks.push({ text, cls })

  while (i < cmd.length) {
    const rest = cmd.slice(i)
    let m: RegExpExecArray | null

    if ((m = /^\s+/.exec(rest))) {
      push(m[0])
      i += m[0].length
      continue
    }
    // UTMStack variable token — highlight in every shell.
    if ((m = /^\$\[[^\]]*\]/.exec(rest))) {
      push(m[0], COL.utm)
      i += m[0].length
      expectCmd = false
      continue
    }
    // Comment (# in bash/powershell).
    if (!isCmd && (m = /^#.*/.exec(rest))) {
      push(m[0], COL.comment)
      i += m[0].length
      continue
    }
    // Quoted strings (allow unterminated while typing).
    if ((m = /^"(?:[^"\\]|\\.)*"?/.exec(rest)) || (m = /^'[^']*'?/.exec(rest))) {
      push(m[0], COL.str)
      i += m[0].length
      expectCmd = false
      continue
    }
    // Variables.
    if (isCmd) {
      if ((m = /^%[^%]*%/.exec(rest))) {
        push(m[0], COL.vari)
        i += m[0].length
        expectCmd = false
        continue
      }
    } else if ((m = /^\$\{[^}]*\}/.exec(rest)) || (m = /^\$[A-Za-z_][\w:]*/.exec(rest))) {
      push(m[0], COL.vari)
      i += m[0].length
      expectCmd = false
      continue
    }
    // Operators / pipes / redirects → next token is a fresh command.
    if ((m = /^(\|\||&&|>>|<<|[|;&<>])/.exec(rest))) {
      push(m[0], COL.op)
      i += m[0].length
      expectCmd = true
      continue
    }
    // Flags.
    if (isCmd) {
      if ((m = /^\/[A-Za-z?][\w]*/.exec(rest))) {
        push(m[0], COL.flag)
        i += m[0].length
        continue
      }
    } else if ((m = /^-{1,2}[A-Za-z][\w-]*/.exec(rest))) {
      push(m[0], COL.flag)
      i += m[0].length
      continue
    }
    // Word (command, cmdlet, or argument).
    if ((m = /^[^\s|;&<>"']+/.exec(rest))) {
      const word = m[0]
      let cls: string | undefined
      if (expectCmd) {
        cls = COL.cmd
        expectCmd = false
      } else if (isPwsh && /^[A-Za-z]+-[A-Za-z]+$/.test(word)) {
        cls = COL.cmd // Verb-Noun cmdlet anywhere in the line
      }
      push(word, cls)
      i += word.length
      continue
    }
    push(rest[0])
    i += 1
  }

  return toks.map((tk, j) => (tk.cls ? <span key={j} className={tk.cls}>{tk.text}</span> : <span key={j}>{tk.text}</span>))
}
