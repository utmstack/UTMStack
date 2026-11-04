import { useMemo } from 'react'
import type { JSX } from 'react'


type Inline = string | JSX.Element

const INLINE_PATTERN = /(\*\*[^*]+\*\*|`[^`]+`|\*[^*]+\*|_[^_]+_)/g

function renderInline(text: string, keyPrefix: string): Inline[] {
  const parts = text.split(INLINE_PATTERN).filter(Boolean)
  return parts.map((part, i) => {
    const key = `${keyPrefix}-${i}`
    if (part.startsWith('**') && part.endsWith('**') && part.length > 4) {
      return (
        <strong key={key} className="font-semibold text-foreground">
          {part.slice(2, -2)}
        </strong>
      )
    }
    if (part.startsWith('`') && part.endsWith('`') && part.length > 2) {
      return (
        <code
          key={key}
          className="rounded-sm bg-muted/70 px-1 py-0.5 font-mono text-[0.85em] text-primary"
        >
          {part.slice(1, -1)}
        </code>
      )
    }
    if (
      ((part.startsWith('*') && part.endsWith('*')) ||
        (part.startsWith('_') && part.endsWith('_'))) &&
      part.length > 2
    ) {
      return (
        <em key={key} className="italic">
          {part.slice(1, -1)}
        </em>
      )
    }
    return part
  })
}

export type Align = 'start' | 'center' | 'end'

interface Block {
  type: 'p' | 'h' | 'ul' | 'ol' | 'table' | 'code'
  level?: number
  lines: string[]
  head?: string[]
  rows?: string[][]
  align?: Align[]
  lang?: string
}

/** ```` ```sh ```` or a bare ```` ``` ````. The language is optional. */
const FENCE_OPEN = /^\s*```\s*([\w+#.-]*)\s*$/
const FENCE_CLOSE = /^\s*```\s*$/

/**
 * Splits one table row into cells.
 *
 * Aware of backticks, because a threat-intel table is full of inline code and
 * a regex or a pattern can legitimately contain a pipe — `a|b` must stay one
 * cell rather than silently becoming two and shifting every column after it.
 * Escaped `\|` is honoured for the same reason.
 */
export function splitRow(line: string): string[] {
  const body = line.trim().replace(/^\|/, '').replace(/\|$/, '')
  const cells: string[] = []
  let cell = ''
  let inCode = false

  for (let i = 0; i < body.length; i++) {
    const ch = body[i]
    if (ch === '\\' && body[i + 1] === '|') {
      cell += '|'
      i++
      continue
    }
    if (ch === '`') inCode = !inCode
    if (ch === '|' && !inCode) {
      cells.push(cell.trim())
      cell = ''
      continue
    }
    cell += ch
  }
  cells.push(cell.trim())
  return cells
}

/** `---`, `:---`, `---:` or `:---:` in every cell — the row under the header. */
export function isDelimiterRow(line: string): boolean {
  if (!line.includes('-')) return false
  const cells = splitRow(line)
  return cells.length > 0 && cells.every((c) => /^:?-+:?$/.test(c))
}

function alignOf(cell: string): Align {
  const left = cell.startsWith(':')
  const right = cell.endsWith(':')
  if (left && right) return 'center'
  if (right) return 'end'
  return 'start'
}

export function parseBlocks(source: string): Block[] {
  const blocks: Block[] = []
  let current: Block | null = null

  const flush = () => {
    if (current) blocks.push(current)
    current = null
  }

  const all = source.split('\n')
  for (let index = 0; index < all.length; index++) {
    const rawLine = all[index]
    const line = rawLine.trimEnd()

    // Fenced code is checked before anything else: a shell command beginning
    // with `-` or a pattern containing `|` would otherwise be reformatted.
    // An unterminated fence runs to the end of the input — the answer is
    // streamed a token at a time, so a half-arrived fence must already
    // render as code.
    const opening = FENCE_OPEN.exec(line)
    if (opening) {
      flush()
      const body: string[] = []
      let cursor = index + 1
      while (cursor < all.length && !FENCE_CLOSE.test(all[cursor])) {
        body.push(all[cursor])
        cursor++
      }
      blocks.push({ type: 'code', lines: body, lang: opening[1] || undefined })
      index = cursor
      continue
    }

    if (!line.trim()) {
      flush()
      continue
    }

    // A table is a header row, a delimiter row, then body rows.
    const next = all[index + 1]?.trimEnd() ?? ''
    if (line.includes('|') && isDelimiterRow(next)) {
      const head = splitRow(line)
      const align = splitRow(next).map(alignOf)
      if (align.length === head.length) {
        flush()
        const rows: string[][] = []
        let cursor = index + 2
        while (cursor < all.length && all[cursor].includes('|') && all[cursor].trim()) {
          const cells = splitRow(all[cursor])
          while (cells.length < head.length) cells.push('')
          rows.push(cells.slice(0, head.length))
          cursor++
        }
        blocks.push({ type: 'table', lines: [], head, rows, align })
        index = cursor - 1
        continue
      }
    }

    const heading = /^(#{1,4})\s+(.*)$/.exec(line)
    if (heading) {
      flush()
      blocks.push({ type: 'h', level: heading[1].length, lines: [heading[2]] })
      continue
    }

    const bullet = /^\s*[-*+]\s+(.*)$/.exec(line)
    if (bullet) {
      if (current?.type !== 'ul') {
        flush()
        current = { type: 'ul', lines: [] }
      }
      current.lines.push(bullet[1])
      continue
    }

    const numbered = /^\s*\d+[.)]\s+(.*)$/.exec(line)
    if (numbered) {
      if (current?.type !== 'ol') {
        flush()
        current = { type: 'ol', lines: [] }
      }
      current.lines.push(numbered[1])
      continue
    }

    if (current?.type !== 'p') {
      flush()
      current = { type: 'p', lines: [] }
    }
    current.lines.push(line.trim())
  }

  flush()
  return blocks
}

export function Markdown({ children }: { children: string }) {
  const blocks = useMemo(() => parseBlocks(children), [children])

  return (
    <div className="space-y-2 text-sm leading-relaxed text-foreground/90">
      {blocks.map((block, i) => {
        const key = `b${i}`
        switch (block.type) {
          case 'h':
            return (
              <div
                key={key}
                className="pt-1 font-mono text-[11px] uppercase tracking-widest text-muted-foreground"
              >
                {renderInline(block.lines[0], key)}
              </div>
            )
          case 'ul':
            return (
              <ul key={key} className="space-y-1 ps-1">
                {block.lines.map((line, j) => (
                  <li key={`${key}-${j}`} className="flex gap-2">
                    <span aria-hidden className="mt-2 h-1 w-1 shrink-0 rounded-full bg-primary/60" />
                    <span className="min-w-0">{renderInline(line, `${key}-${j}`)}</span>
                  </li>
                ))}
              </ul>
            )
          case 'ol':
            return (
              <ol key={key} className="space-y-1 ps-1">
                {block.lines.map((line, j) => (
                  <li key={`${key}-${j}`} className="flex gap-2">
                    <span className="shrink-0 font-mono text-[11px] text-primary/60">{j + 1}.</span>
                    <span className="min-w-0">{renderInline(line, `${key}-${j}`)}</span>
                  </li>
                ))}
              </ol>
            )
          case 'code':
            return (
              // Scrolls inside itself; `renderInline` deliberately not applied —
              // asterisks and backticks in code are characters, not formatting.
              <pre
                key={key}
                className="overflow-x-auto rounded-sm border border-border/50 bg-surface/40 p-2.5"
              >
                <code className="font-mono text-xs leading-relaxed text-foreground/90">
                  {block.lines.join('\n')}
                </code>
              </pre>
            )

          case 'table': {
            const align = block.align ?? []
            const cellAlign = (c: number) =>
              align[c] === 'center' ? 'text-center' : align[c] === 'end' ? 'text-end' : 'text-start'
            return (
              <div key={key} className="overflow-x-auto">
                <table className="w-full border-collapse text-xs">
                  <thead>
                    <tr className="border-b border-border/50">
                      {(block.head ?? []).map((cell, c) => (
                        <th
                          key={`${key}-h${c}`}
                          scope="col"
                          className={`px-2 py-1.5 font-mono text-[10px] font-normal uppercase tracking-widest text-muted-foreground ${cellAlign(c)}`}
                        >
                          {renderInline(cell, `${key}-h${c}`)}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {(block.rows ?? []).map((row, r) => (
                      <tr key={`${key}-r${r}`} className="border-b border-border/20 last:border-0">
                        {row.map((cell, c) => (
                          <td
                            key={`${key}-r${r}c${c}`}
                            className={`px-2 py-1.5 align-top text-foreground/90 ${cellAlign(c)}`}
                          >
                            {renderInline(cell, `${key}-r${r}c${c}`)}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )
          }
          default:
            return <p key={key}>{renderInline(block.lines.join(' '), key)}</p>
        }
      })}
    </div>
  )
}
