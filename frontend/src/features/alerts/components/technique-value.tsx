import { ExternalLink } from 'lucide-react'

// Renders the MITRE technique as a link to attack.mitre.org when it carries a
// technique id (e.g. "T1110 - Brute Force" or "T1059.001 - PowerShell").
export function TechniqueValue({ technique }: { technique?: string }) {
  if (!technique) return <span className="font-mono">—</span>
  const m = technique.match(/T\d{4}(?:\.\d{3})?/i)
  if (!m) return <span className="font-mono">{technique}</span>
  const id = m[0].toUpperCase()
  const path = id.replace('.', '/')
  return (
    <a
      href={`https://attack.mitre.org/techniques/${path}`}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center gap-1 font-mono text-primary hover:underline"
    >
      {technique}
      <ExternalLink size={11} className="shrink-0" />
    </a>
  )
}
