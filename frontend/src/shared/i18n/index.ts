import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

/**
 * Auto-register every locale file in ./locales. Adding a new language is just
 * dropping a `<code>.json` file here (and, optionally, a label below) — no code
 * changes elsewhere.
 */
const localeModules = import.meta.glob('./locales/*.json', { eager: true }) as Record<
  string,
  { default: Record<string, unknown> }
>

const resources: Record<string, { translation: Record<string, unknown> }> = {}
for (const path in localeModules) {
  const code = path.match(/\/([a-z-]+)\.json$/)?.[1]
  if (code) resources[code] = { translation: localeModules[path].default }
}

/** Human labels for the language picker. Falls back to the code if unmapped. */
const LANGUAGE_LABELS: Record<string, string> = {
  en: 'English',
  es: 'Español',
  pt: 'Português',
  fr: 'Français',
  de: 'Deutsch',
  it: 'Italiano',
  ru: 'Русский',
}

export interface SupportedLanguage {
  code: string
  label: string
}

export const SUPPORTED_LANGUAGES: SupportedLanguage[] = Object.keys(resources)
  .sort()
  .map((code) => ({ code, label: LANGUAGE_LABELS[code] ?? code }))

export const FALLBACK_LANGUAGE = 'en'

/** Persisted key for the detected/chosen language (also synced to the user's lang_key). */
export const LANGUAGE_STORAGE_KEY = 'utmstack_lang'

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: FALLBACK_LANGUAGE,
    supportedLngs: Object.keys(resources),
    nonExplicitSupportedLngs: true, // 'es-ES' → 'es'
    interpolation: { escapeValue: false }, // React already escapes
    detection: {
      order: ['localStorage', 'navigator'],
      lookupLocalStorage: LANGUAGE_STORAGE_KEY,
      caches: ['localStorage'],
    },
  })

export default i18n
