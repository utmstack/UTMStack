import 'react-i18next'
import type { EnTranslation } from './types'

// Make t() keys type-safe and autocompletable, derived from the base (en) locale.
// A missing or misspelled key becomes a TypeScript error.
declare module 'react-i18next' {
  interface CustomTypeOptions {
    defaultNS: 'translation'
    resources: {
      translation: EnTranslation
    }
  }
}
