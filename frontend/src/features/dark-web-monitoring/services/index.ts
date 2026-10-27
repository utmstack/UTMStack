import { darkWebHttpService } from './dark-web-http.service'
import { darkWebMockService } from './dark-web-mock.service'

/**
 * Single entry point for dark web monitoring data. The API endpoints are not
 * implemented on the backend yet, so this resolves to the mock; flip this
 * flag to `false` (or drop the `VITE_DARK_WEB_USE_MOCK` override) once they
 * land.
 */
const USE_MOCK = import.meta.env.VITE_DARK_WEB_USE_MOCK !== 'false'

export const darkWebMonitoringService = USE_MOCK ? darkWebMockService : darkWebHttpService
