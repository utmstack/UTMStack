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

// Captured BEFORE init: i18next caches the detected language to localStorage on
// init, which would otherwise make a first visit look like the user already
// chose a language. This lets bootstrapPlatformLanguage tell "this browser
// already has a preference" apart from "first visit → apply the platform default".
const hadStoredLanguage =
  typeof localStorage !== 'undefined' && localStorage.getItem(LANGUAGE_STORAGE_KEY) != null

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

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

/**
 * Apply the platform-default UI language (from GET /branding/public) for browsers
 * that have no language preference yet. Precedence is: a per-user lang_key (applied
 * post-login by the auth context) → this platform default → browser → English.
 * A browser that already has a stored choice is left untouched.
 */
export async function bootstrapPlatformLanguage(): Promise<void> {
  if (hadStoredLanguage) return
  try {
    const res = await fetch(`${API_BASE}/branding/public`)
    if (!res.ok) return
    const data = (await res.json()) as { language?: string }
    const lang = data.language
    if (lang && resources[lang]) await i18n.changeLanguage(lang)
  } catch {
    // Backend not reachable yet (offline / boot): keep the detected language.
  }
}

export default i18n
