// Minimal ambient types for js-yaml 4.x (the bits we use). The package ships
// without bundled types and @types/js-yaml is not installed, so we declare the
// small surface the parsing-filters editor relies on.
declare module 'js-yaml' {
  export interface LoadOptions {
    filename?: string
    json?: boolean
  }
  export interface DumpOptions {
    indent?: number
    lineWidth?: number
    noRefs?: boolean
    sortKeys?: boolean
    quotingType?: '"' | "'"
    forceQuotes?: boolean
  }
  export function load(input: string, options?: LoadOptions): unknown
  export function dump(obj: unknown, options?: DumpOptions): string
  export class YAMLException extends Error {}
}
