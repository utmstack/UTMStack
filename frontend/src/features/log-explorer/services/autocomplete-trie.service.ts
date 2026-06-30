export type SuggestionTag = 'sql' | 'field' | 'index'

export interface Suggestion {
  word: string
  tag: SuggestionTag
}

interface TrieNode {
  children: Map<string, TrieNode>
  word?: string
  tag?: SuggestionTag
}

const createNode = (): TrieNode => ({ children: new Map() })

export class AutocompleteTrie {
  private readonly root: TrieNode = createNode()

  insert(word: string, tag: SuggestionTag): void {
    if (!word) return
    let node = this.root
    const lower = word.toLowerCase()
    for (const ch of lower) {
      let next = node.children.get(ch)
      if (!next) {
        next = createNode()
        node.children.set(ch, next)
      }
      node = next
    }
    node.word = word
    node.tag = tag
  }

  clearTag(tag: SuggestionTag): void {
    this.stripTag(this.root, tag)
  }

  private stripTag(node: TrieNode, tag: SuggestionTag): boolean {
    if (node.tag === tag) {
      node.word = undefined
      node.tag = undefined
    }
    for (const [ch, child] of node.children) {
      const childEmpty = this.stripTag(child, tag)
      if (childEmpty) node.children.delete(ch)
    }
    return node.word === undefined && node.children.size === 0
  }

  suggest(prefix: string, limit = 20): Suggestion[] {
    if (!prefix) return []
    let node: TrieNode | undefined = this.root
    for (const ch of prefix.toLowerCase()) {
      node = node.children.get(ch)
      if (!node) return []
    }
    const buckets: Record<SuggestionTag, Suggestion[]> = { sql: [], field: [], index: [] }
    this.collect(node, buckets, limit)
    return [...buckets.sql, ...buckets.index, ...buckets.field].slice(0, limit)
  }

  private collect(
    node: TrieNode,
    buckets: Record<SuggestionTag, Suggestion[]>,
    limit: number,
  ): void {
    const size = buckets.sql.length + buckets.field.length + buckets.index.length
    if (size >= limit) return
    if (node.word && node.tag) {
      buckets[node.tag].push({ word: node.word, tag: node.tag })
    }
    for (const child of node.children.values()) {
      const cur = buckets.sql.length + buckets.field.length + buckets.index.length
      if (cur >= limit) return
      this.collect(child, buckets, limit)
    }
  }
}

export const createAutocompleteTrie = (): AutocompleteTrie => new AutocompleteTrie()
