import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

const localeLoaders = import.meta.glob('./locales/**/*.json', {
  query: '?raw',
  import: 'default',
}) as Record<string, () => Promise<string>>

const supported = new Set<string>()
const byLanguage: Record<string, string[]> = {}
for (const path in localeLoaders) {
  const match = path.match(/\/locales\/([a-z-]+)(?:\/(\w+))?\.json$/)
  if (!match) continue
  supported.add(match[1])
  ;(byLanguage[match[1]] ??= []).push(path)
}

/** Load one language's domain files and merge them into a single bundle. */
async function loadLanguage(code: string): Promise<Record<string, unknown>> {
  const translation: Record<string, unknown> = {}
  for (const path of byLanguage[code] ?? []) {
    let json: Record<string, unknown>
    try {
      json = JSON.parse(await localeLoaders[path]())
    } catch {
      console.error('i18n: bad JSON in locale file', path)
      continue
    }
    const match = path.match(/\/locales\/[a-z-]+\/(\w+)\.json$/)
    if (match) translation[match[1]] = json
    else Object.assign(translation, json)
  }
  return translation
}

/** Minimal i18next backend: v26 core calls read() per namespace (loadLanguageResources is v19-era, never invoked). */
const cache = new Map<string, Promise<Record<string, unknown>>>()
const lazyBackend = {
  type: 'backend' as const,
  init() {},

  // v26 core calls read(lng, ns) once per hierarchy level and consumes the returned Promise (2-arg form).
  read(lngs: string | string[], _ns: string): Promise<Record<string, unknown>> {
    const lng = Array.isArray(lngs) ? lngs[0] : lngs
    if (!supported.has(lng)) return Promise.resolve({})
    let p = cache.get(lng)
    if (!p) {
      p = loadLanguage(lng).catch((err) => {
        cache.delete(lng)
        console.error('i18n: failed to load language', err)
        return {} as Record<string, unknown>
      })
      cache.set(lng, p)
    }
    return p
  },
  create(_lngs: string[], _ns: string, cb: (err: null) => void) {
    cb(null)
  },
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

export const SUPPORTED_LANGUAGES: SupportedLanguage[] = [...supported]
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
  .use(lazyBackend)
  .use(initReactI18next)
  .init({
    // No upfront `resources`: the backend loads the active (and fallback) language.
    fallbackLng: FALLBACK_LANGUAGE,
    supportedLngs: [...supported],
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
    if (lang && supported.has(lang)) await i18n.changeLanguage(lang)
  } catch {
    // Backend not reachable yet (offline / boot): keep the detected language.
  }
}

export default i18n
