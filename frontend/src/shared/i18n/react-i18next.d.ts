import 'react-i18next'
import type en from './locales/en.json'

// Make t() keys type-safe and autocompletable, derived from the base (en) locale.
// A missing or misspelled key becomes a TypeScript error.
declare module 'react-i18next' {
  interface CustomTypeOptions {
    defaultNS: 'translation'
    resources: {
      translation: typeof en
    }
  }
}
