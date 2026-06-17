/**
 * App run mode. The SAME frontend bundle serves two products:
 *   - instance mode (default): a single UTMStack instance.
 *   - federation mode: the MSSP console that proxies many instances.
 *
 * Resolved at RUNTIME by the backend: at boot, mode-bootstrap.ts probes the FS
 * (GET /api/v1/mode) and sets `window.__UTM_FEDERATION__` before the app graph
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

/**
 * Base path of the Federation Service's own API. The FS shares the /api/v1
 * namespace with the instances it proxies: it handles its own subpaths
 * (/api/v1/auth, /api/v1/users, /api/v1/config, /api/v1/instances, /api/v1/mode)
 * and forwards every other /api/v1/* call to the selected instance.
 */
export const FS_API_URL = '/api/v1'
