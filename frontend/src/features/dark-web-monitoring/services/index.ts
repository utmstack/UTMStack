import { darkWebHttpService } from './dark-web-http.service'
import { darkWebMockService } from './dark-web-mock.service'

/**
 * Explicit feature switch for the dark web data source. The API endpoints are
 * not implemented on the backend yet (the proposed contract lives in
 * `dark-web-http.service.ts`), so the default is `true` = mock. This is an
 * intentional, visible flag: do NOT rely on the env override in production,
 * and flip `false` (or delete the mock service) once the endpoints are live.
 * `VITE_DARK_WEB_USE_MOCK=false` is kept only as a per-environment override.
 */
const USE_MOCK: boolean = import.meta.env.VITE_DARK_WEB_USE_MOCK !== 'false'

export const darkWebMonitoringService = USE_MOCK ? darkWebMockService : darkWebHttpService
