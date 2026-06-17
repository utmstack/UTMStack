/**
 * Resolves the app run mode BEFORE the app graph loads, so the synchronous
 * `IS_FEDERATION` constant (mode.ts) reads the right value everywhere.
 *
 * The same frontend artifact serves both products and is NOT rebuilt per mode,
 * so the backend decides: the Federation Service answers GET /api/v1/mode with
 * {federation:true}; an instance backend has no such route (404 / SPA-fallback
 * HTML) → instance mode. Result is stashed on window.__UTM_FEDERATION__.
 *
 * IMPORTANT: this module must not import mode.ts (that would evaluate
 * IS_FEDERATION too early). main.tsx awaits this, then dynamically imports the app.
 */
export async function resolveFederationMode(): Promise<void> {
  // Dev/build override wins and stays synchronous (npm run dev:federation).
  if (import.meta.env.VITE_FEDERATION === 'true') {
    window.__UTM_FEDERATION__ = true
    return
  }
  try {
    const ctrl = new AbortController()
    const timer = setTimeout(() => ctrl.abort(), 4000)
    const res = await fetch('/api/v1/mode', {
      signal: ctrl.signal,
      headers: { Accept: 'application/json' },
    })
    clearTimeout(timer)
    // Only the FS backend serves JSON here; an instance returns 404 (or the SPA
    // fallback HTML), both of which mean instance mode.
    if (res.ok && (res.headers.get('content-type') ?? '').includes('application/json')) {
      const cfg = (await res.json()) as { federation?: boolean }
      window.__UTM_FEDERATION__ = cfg.federation === true
      return
    }
  } catch {
    // Network error / timeout → fall back to instance mode.
  }
  window.__UTM_FEDERATION__ = false
}
