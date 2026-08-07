// Notes written by the AI SOC agent start with this marker and pack labelled
// sections separated by " | ". We parse them into a structured assessment.
export const AI_NOTE_MARKER = '[AI SOC Agent]'

// A stored notes value sometimes arrives JSON-encoded (wrapped in double quotes)
// from the pipeline that persisted the AI assessment. Unwrap it so the marker
// sits at the start and the analyst's textarea doesn't show a stray leading `"`.
export function unquoteNotes(s?: string): string {
  if (!s) return ''
  const v = s
  if (v.length >= 2 && v[0] === '"' && v[v.length - 1] === '"') {
    try {
      const parsed = JSON.parse(v)
      if (typeof parsed === 'string') return parsed
    } catch {
      return v.slice(1, -1)
    }
  }
  return v
}

// Robust detection: the marker may appear with leading whitespace/quotes or a
// slightly different case, so we look for it anywhere, case-insensitively.
export function aiMarkerIndex(s?: string): number {
  const v = unquoteNotes(s)
  return v ? v.toUpperCase().indexOf(AI_NOTE_MARKER.toUpperCase()) : -1
}

export function isAiNote(s?: string): boolean {
  return aiMarkerIndex(s) >= 0
}

// A stored `notes` value may contain the analyst's own text followed by the AI
// SOC block. We keep them separable so the analyst can keep adding notes while
// the AI assessment (rendered read-only) is preserved untouched.
export function aiBlock(notes?: string): string {
  const v = unquoteNotes(notes)
  const idx = aiMarkerIndex(notes)
  return idx >= 0 ? v.slice(idx) : ''
}

export function userNotePart(notes?: string): string {
  const v = unquoteNotes(notes)
  if (!v) return ''
  const idx = aiMarkerIndex(notes)
  return (idx >= 0 ? v.slice(0, idx) : v).trim()
}

export function combineUserNote(userText: string, original?: string): string {
  const ai = aiBlock(original)
  const u = userText.trim()
  if (!ai) return u
  return u ? `${u}\n\n${ai}` : ai
}

export interface AiNote {
  headline: string
  score?: number
  risk?: string
  sections: { label: string; value: string }[]
}

// The agent writes free text into every section, so a pipe inside one of them
// is indistinguishable from the separator by position alone. What tells them
// apart is the label: a real section opens with one, a continuation does not.
// Splitting on the bare "|" turned "Action: block the IP | then escalate" into
// two entries, the second of them unlabelled and orphaned.
const SECTION_LABEL = /^[A-Za-z][A-Za-z ]{0,30}:\s/

function splitSections(body: string): string[] {
  const out: string[] = []
  for (const fragment of body.split(/\s\|\s/)) {
    const part = fragment.trim()
    if (!part) continue
    // The first fragment is the headline, which carries no label of its own.
    if (out.length > 0 && !SECTION_LABEL.test(part)) {
      out[out.length - 1] += ' | ' + part
      continue
    }
    out.push(part)
  }
  return out
}

export function parseAiNote(notes?: string): AiNote | null {
  const v = unquoteNotes(notes)
  const idx = aiMarkerIndex(notes)
  if (idx < 0 || !v) return null
  const body = v.slice(idx + AI_NOTE_MARKER.length)
  const parts = splitSections(body)
  const headline = parts[0] ?? ''
  const sections = parts.slice(1).map((p) => {
    const i = p.indexOf(':')
    return i > 0 ? { label: p.slice(0, i).trim(), value: p.slice(i + 1).trim() } : { label: '', value: p }
  })
  const scoreM = headline.match(/Score:\s*([\d.]+)/i)
  const riskM = headline.match(/(low|medium|high)\s*risk/i)
  return {
    headline,
    score: scoreM ? Number(scoreM[1]) : undefined,
    risk: riskM ? riskM[1].toLowerCase() : undefined,
    sections,
  }
}
