import type { ReactNode } from 'react'

function inlineFmt(text: string): ReactNode[] {
  const out: ReactNode[] = []
  const re = /\*\*([^*]+)\*\*|`([^`]+)`/g
  let last = 0
  let key = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(text))) {
    if (m.index > last) out.push(text.slice(last, m.index))
    if (m[1] != null) out.push(<strong key={key++} className="font-semibold text-foreground">{m[1]}</strong>)
    else if (m[2] != null) out.push(<code key={key++} className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">{m[2]}</code>)
    last = m.index + m[0].length
  }
  if (last < text.length) out.push(text.slice(last))
  return out
}

export function RuleDescription({ text }: { text: string }) {
  const lines = text.split('\n')
  const blocks: ReactNode[] = []
  let i = 0
  let key = 0
  const isOl = (l: string) => /^\d+[.)]\s+/.test(l.trim())
  const isUl = (l: string) => /^[-*]\s+/.test(l.trim())

  while (i < lines.length) {
    if (lines[i].trim() === '') {
      i++
      continue
    }
    if (isOl(lines[i])) {
      const items: string[] = []
      while (i < lines.length && isOl(lines[i])) items.push(lines[i++].trim().replace(/^\d+[.)]\s+/, ''))
      blocks.push(<ol key={key++} className="ml-5 list-decimal space-y-1">{items.map((it, j) => <li key={j}>{inlineFmt(it)}</li>)}</ol>)
      continue
    }
    if (isUl(lines[i])) {
      const items: string[] = []
      while (i < lines.length && isUl(lines[i])) items.push(lines[i++].trim().replace(/^[-*]\s+/, ''))
      blocks.push(<ul key={key++} className="ml-5 list-disc space-y-1">{items.map((it, j) => <li key={j}>{inlineFmt(it)}</li>)}</ul>)
      continue
    }
    const para: string[] = []
    while (i < lines.length && lines[i].trim() !== '' && !isOl(lines[i]) && !isUl(lines[i])) para.push(lines[i++].trim())
    blocks.push(<p key={key++}>{inlineFmt(para.join(' '))}</p>)
  }
  return <div className="space-y-2 text-xs leading-relaxed text-muted-foreground">{blocks}</div>
}
