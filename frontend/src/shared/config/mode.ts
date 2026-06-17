/**
 * App run mode. The SAME frontend bundle serves two products:
 *   - instance mode (default): a single UTMStack instance.
 *   - federation mode: the MSSP console that proxies many instances.
 *
 * Resolved at RUNTIME by the backend: at boot, mode-bootstrap.ts probes the FS
 * (GET /fs/config) and sets `window.__UTM_FEDERATION__` before the app graph
 * loads (the federation frontend isn't built in this repo, so it can't rely on a
 * build flag). In dev, `VITE_FEDERATION=true` (npm run dev:federation) still
 * forces it without a backend.
 */
declare global {
  interface Window {
    __UTM_FEDERATION__?: boolean
  }
}

export const IS_FEDERATION: boolean =
  (typeof window !== 'undefined' && window.__UTM_FEDERATION__ === true) ||
  import.meta.env.VITE_FEDERATION === 'true'

/** Base path of the Federation Service's own API (login, instances). */
export const FS_API_URL = '/fs'
