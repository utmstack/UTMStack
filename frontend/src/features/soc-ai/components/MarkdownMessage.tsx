import { useEffect, useMemo, useRef } from 'react'
import { useThemeContext } from '@/app/providers'
import { renderMarkdown } from '../lib/markdown'
import { renderMermaid } from '../lib/mermaid'

/**
 * Renders an assistant reply: Markdown → HTML (callouts, tables, code, …), then
 * post-processes any `pre.socai-mermaid` nodes into real SVG diagrams. Mirrors
 * GitDocAI's "render markdown, hydrate the diagram islands" approach.
 */
export function MarkdownMessage({ text }: { text: string }) {
  const ref = useRef<HTMLDivElement>(null)
  const { theme } = useThemeContext()
  const html = useMemo(() => renderMarkdown(text), [text])

  useEffect(() => {
    const root = ref.current
    if (!root) return
    let cancelled = false
    const blocks = Array.from(root.querySelectorAll<HTMLElement>('pre.socai-mermaid'))
    for (const pre of blocks) {
      const code = pre.textContent || ''
      void renderMermaid(code, theme === 'dark')
        .then((svg) => {
          if (cancelled) return
          const wrap = document.createElement('div')
          wrap.className = 'socai-diagram'
          wrap.innerHTML = svg
          pre.replaceWith(wrap)
        })
        .catch(() => {
          pre.classList.add('socai-mermaid--error')
        })
    }
    return () => {
      cancelled = true
    }
  }, [html, theme])

  return <div ref={ref} className="socai-md" dangerouslySetInnerHTML={{ __html: html }} />
}
