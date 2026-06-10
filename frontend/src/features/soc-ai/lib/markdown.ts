/**
 * Dependency-free Markdown → HTML for assistant replies. Ported from GitDocAI's
 * publish-web/public/md.js so SOC-AI renders the same rich blocks the assistant
 * emits: callouts, GFM tables, code fences (```mermaid → a diagram node the
 * <MarkdownMessage> post-processes), inline images, cards.
 *
 * HTML is escaped first so raw model HTML never gets through; then a safe subset
 * of Markdown is applied. Underscore emphasis is intentionally unsupported — it
 * would mangle snake_case identifiers common in technical/SIEM content.
 */

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

const linkUrl = (url: string) => (/^(https?:|mailto:|\/)/i.test(url) ? url : '#')
const imgUrl = (url: string) => (/^(https?:|\/)/i.test(url) ? url : '')

const CALLOUT_ALIAS: Record<string, string> = {
  tip: 'tip',
  note: 'note',
  info: 'info',
  important: 'info',
  warning: 'warning',
  caution: 'warning',
  danger: 'danger',
  success: 'success',
}

function calloutIcon(type: string): string {
  const a =
    '<svg class="socai-callout__icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">'
  switch (type) {
    case 'warning':
    case 'danger':
      return (
        a +
        '<path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>'
      )
    case 'tip':
      return (
        a +
        '<path d="M9 18h6"/><path d="M10 22h4"/><path d="M15.09 14c.18-.98.65-1.74 1.41-2.5A4.65 4.65 0 0 0 18 8 6 6 0 0 0 6 8c0 1 .23 2.23 1.5 3.5A4.61 4.61 0 0 1 8.91 14"/></svg>'
      )
    case 'success':
      return a + '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>'
    default:
      return a + '<circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>'
  }
}

function renderInline(s: string): string {
  const codeSpans: string[] = []
  s = s.replace(/`([^`]+)`/g, (_m, c: string) => {
    codeSpans.push(c)
    return `@@CODE${codeSpans.length - 1}@@`
  })
  s = s.replace(/!\[([^\]]*)\]\(([^)\s]+)\)/g, (_m, alt: string, url: string) => {
    const safe = imgUrl(url)
    return safe ? `<img class="socai-img" src="${safe}" alt="${alt}" loading="lazy">` : ''
  })
  s = s.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_m, label: string, url: string) => {
    const safe = linkUrl(url)
    return `<a href="${safe}" target="_blank" rel="noopener noreferrer">${label}</a>`
  })
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  s = s.replace(/\*([^*\n]+)\*/g, '<em>$1</em>')
  s = s.replace(/@@CODE(\d+)@@/g, (_m, i: string) => `<code>${codeSpans[Number(i)]}</code>`)
  return s
}

const splitRow = (l: string) =>
  l
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((c) => c.trim())

function renderBlocks(lines: string[]): string {
  const out: string[] = []
  let listType: 'ul' | 'ol' | null = null
  const closeList = () => {
    if (listType) {
      out.push(`</${listType}>`)
      listType = null
    }
  }
  let i = 0
  while (i < lines.length) {
    const line = lines[i]

    // Fenced code (```lang); mermaid becomes a diagram node.
    const fence = line.match(/^```(\w*)/)
    if (fence) {
      closeList()
      const lang = (fence[1] || '').toLowerCase()
      const code: string[] = []
      i++
      while (i < lines.length && !/^```\s*$/.test(lines[i])) {
        code.push(lines[i])
        i++
      }
      i++
      out.push(
        lang === 'mermaid'
          ? `<pre class="socai-mermaid">${code.join('\n')}</pre>`
          : `<pre><code>${code.join('\n')}</code></pre>`,
      )
      continue
    }

    // Blockquotes / GitHub-style callouts. `>` was escaped to `&gt;`.
    if (/^\s*&gt;/.test(line)) {
      closeList()
      const quoted: string[] = []
      while (i < lines.length && /^\s*&gt;/.test(lines[i])) {
        quoted.push(lines[i].replace(/^\s*&gt;\s*/, ''))
        i++
      }
      const firstIdx = quoted.findIndex((q) => q.trim().length > 0)
      const probe = firstIdx >= 0 ? quoted[firstIdx] : ''
      const alert = probe.match(/^\[!(\w+)\]:?\s*$/i)
      if (alert) {
        const type = CALLOUT_ALIAS[alert[1].toLowerCase()] || 'note'
        out.push(
          `<div class="socai-callout socai-callout--${type}">${calloutIcon(type)}<div class="socai-callout__body">${renderBlocks(quoted.slice(firstIdx + 1))}</div></div>`,
        )
      } else {
        out.push(`<blockquote>${renderBlocks(quoted)}</blockquote>`)
      }
      continue
    }

    // GFM table — header row followed by a |---|---| separator.
    if (line.includes('|') && i + 1 < lines.length && /^[\s|:-]+$/.test(lines[i + 1]) && lines[i + 1].includes('-')) {
      closeList()
      const header = splitRow(line)
      i += 2
      const rows: string[][] = []
      while (i < lines.length && lines[i].includes('|') && lines[i].trim() !== '') {
        rows.push(splitRow(lines[i]))
        i++
      }
      let html = '<div class="socai-table-wrap"><table class="socai-table"><thead><tr>'
      for (const h of header) html += `<th>${renderInline(h)}</th>`
      html += '</tr></thead><tbody>'
      for (const r of rows) {
        html += '<tr>'
        for (let c = 0; c < header.length; c++) html += `<td>${renderInline(r[c] ?? '')}</td>`
        html += '</tr>'
      }
      html += '</tbody></table></div>'
      out.push(html)
      continue
    }

    const h = line.match(/^(#{1,6})\s+(.*)$/)
    if (h) {
      closeList()
      const lvl = Math.min(h[1].length, 4)
      out.push(`<h${lvl}>${renderInline(h[2])}</h${lvl}>`)
      i++
      continue
    }
    if (/^(---|\*\*\*|___)\s*$/.test(line)) {
      closeList()
      out.push('<hr>')
      i++
      continue
    }
    const ul = line.match(/^\s*[-*+]\s+(.*)$/)
    if (ul) {
      if (listType !== 'ul') {
        closeList()
        out.push('<ul>')
        listType = 'ul'
      }
      out.push(`<li>${renderInline(ul[1])}</li>`)
      i++
      continue
    }
    const ol = line.match(/^\s*\d+\.\s+(.*)$/)
    if (ol) {
      if (listType !== 'ol') {
        closeList()
        out.push('<ol>')
        listType = 'ol'
      }
      out.push(`<li>${renderInline(ol[1])}</li>`)
      i++
      continue
    }
    if (line.trim() === '') {
      closeList()
      i++
      continue
    }
    closeList()
    const para = [line]
    i++
    while (
      i < lines.length &&
      lines[i].trim() !== '' &&
      !/^```/.test(lines[i]) &&
      !/^\s*&gt;/.test(lines[i]) &&
      !/^#{1,6}\s/.test(lines[i]) &&
      !/^\s*[-*+]\s/.test(lines[i]) &&
      !/^\s*\d+\.\s/.test(lines[i]) &&
      !/^(---|\*\*\*|___)\s*$/.test(lines[i])
    ) {
      para.push(lines[i])
      i++
    }
    out.push(`<p>${renderInline(para.join('<br>'))}</p>`)
  }
  closeList()
  return out.join('\n')
}

export function renderMarkdown(text: string): string {
  return renderBlocks(escapeHtml(text).split('\n'))
}
