/**
 * Parses a CEL `where` condition into a boolean tree so the read-only rule view
 * can render nested AND/OR groups (with parentheses and mixed operators) visually
 * — not just the flat single-operator lists the editor builder handles.
 *
 * Grammar (subset used by correlation rules):
 *   or   := and ( '||' and )*
 *   and  := unary ( '&&' unary )*
 *   unary:= '!'* primary
 *   prim := '(' or ')' | call
 *   call := ident '(' arg ( ',' arg )* ')'
 */

export interface CelCond {
  type: 'cond'
  fn: string
  field: string
  values: string[]
  negate: boolean
}
export interface CelGroup {
  type: 'and' | 'or'
  children: CelNode[]
  negate: boolean
}
export type CelNode = CelCond | CelGroup

type Tok =
  | { t: 'and' }
  | { t: 'or' }
  | { t: 'not' }
  | { t: 'lp' }
  | { t: 'rp' }
  | { t: 'call'; fn: string; args: string[] }

function tokenize(s: string): Tok[] | null {
  const toks: Tok[] = []
  let i = 0
  while (i < s.length) {
    const c = s[i]
    if (/\s/.test(c)) {
      i++
      continue
    }
    if (s.startsWith('&&', i)) {
      toks.push({ t: 'and' })
      i += 2
      continue
    }
    if (s.startsWith('||', i)) {
      toks.push({ t: 'or' })
      i += 2
      continue
    }
    if (c === '!') {
      toks.push({ t: 'not' })
      i++
      continue
    }
    if (c === '(') {
      toks.push({ t: 'lp' })
      i++
      continue
    }
    if (c === ')') {
      toks.push({ t: 'rp' })
      i++
      continue
    }
    const m = /^([A-Za-z_]\w*)\s*\(/.exec(s.slice(i))
    if (m) {
      const fn = m[1]
      i += m[0].length // consume `ident(`
      const args: string[] = []
      let closed = false
      while (i < s.length) {
        while (i < s.length && /\s/.test(s[i])) i++
        if (s[i] === ')') {
          i++
          closed = true
          break
        }
        if (s[i] === ',') {
          i++
          continue
        }
        if (s[i] === '"' || s[i] === "'") {
          const q = s[i]
          i++
          let val = ''
          while (i < s.length && s[i] !== q) {
            if (s[i] === '\\' && i + 1 < s.length) {
              val += s[i + 1]
              i += 2
              continue
            }
            val += s[i]
            i++
          }
          i++ // closing quote
          args.push(val)
        } else {
          let val = ''
          while (i < s.length && !/[,)\s]/.test(s[i])) {
            val += s[i]
            i++
          }
          args.push(val)
        }
      }
      if (!closed) return null // unterminated call
      toks.push({ t: 'call', fn, args })
      continue
    }
    return null // unexpected character
  }
  return toks
}

export function parseCelTree(cel: string): CelNode | null {
  if (!cel.trim()) return null
  const toks = tokenize(cel)
  if (!toks || toks.length === 0) return null

  let pos = 0
  const peek = () => toks[pos]
  const eat = () => toks[pos++]

  const parseOr = (): CelNode | null => {
    const first = parseAnd()
    if (!first) return null
    const children = [first]
    while (peek()?.t === 'or') {
      eat()
      const n = parseAnd()
      if (!n) return null
      children.push(n)
    }
    return children.length === 1 ? first : { type: 'or', children, negate: false }
  }
  const parseAnd = (): CelNode | null => {
    const first = parseUnary()
    if (!first) return null
    const children = [first]
    while (peek()?.t === 'and') {
      eat()
      const n = parseUnary()
      if (!n) return null
      children.push(n)
    }
    return children.length === 1 ? first : { type: 'and', children, negate: false }
  }
  const parseUnary = (): CelNode | null => {
    let neg = false
    while (peek()?.t === 'not') {
      eat()
      neg = !neg
    }
    const node = parsePrimary()
    if (!node) return null
    if (neg) node.negate = !node.negate
    return node
  }
  const parsePrimary = (): CelNode | null => {
    const tk = peek()
    if (!tk) return null
    if (tk.t === 'lp') {
      eat()
      const inner = parseOr()
      if (!inner || peek()?.t !== 'rp') return null
      eat()
      return inner
    }
    if (tk.t === 'call') {
      eat()
      return { type: 'cond', fn: tk.fn, field: tk.args[0] ?? '', values: tk.args.slice(1), negate: false }
    }
    return null
  }

  const result = parseOr()
  if (pos !== toks.length) return null // leftover tokens → not fully parsed
  return result
}
