/**
 * Lazy mermaid loader — the lib is ~heavy, so it's only pulled in the first time
 * an answer actually contains a diagram, never on the main bundle path.
 */
let loader: Promise<typeof import('mermaid')['default']> | null = null

function getMermaid() {
  if (!loader) loader = import('mermaid').then((m) => m.default)
  return loader
}

let counter = 0

/** Render mermaid source to an SVG string in the current (dark/light) theme. */
export async function renderMermaid(code: string, dark: boolean): Promise<string> {
  const mermaid = await getMermaid()
  mermaid.initialize({
    startOnLoad: false,
    theme: dark ? 'dark' : 'default',
    securityLevel: 'strict',
    fontFamily: 'inherit',
  })
  const { svg } = await mermaid.render(`socai-mmd-${++counter}`, code.trim())
  return svg
}
