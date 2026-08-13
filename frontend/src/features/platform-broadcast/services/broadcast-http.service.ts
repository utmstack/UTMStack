import { createApiClient } from '@/shared/lib/api-client'

/**
 * One typed service for every `/platform/**\/bulk/**` endpoint.
 *
 * Every bulk endpoint takes the same selector and returns the same shape. The
 * per-endpoint helpers below only differ in the resource payload they merge
 * into the request body — the caller assembles that payload from the same
 * form state the single-tenant version uses.
 *
 * Backend contract lives in backend/bulk.md.
 */

export interface BulkSelector {
  tenantIds: string[]
  allTenants: boolean
}

export interface BulkFailure {
  tenantId: string
  error: string
}

export interface BulkResult {
  succeeded: string[]
  failed: BulkFailure[]
}

const api = createApiClient()

function withSelector<T extends object>(selector: BulkSelector, resource: T) {
  return { selector, ...resource }
}

// Go's nil slices marshal to JSON null; normalize so consumers can safely read `.length`.
function normalize(r: BulkResult | null | undefined): BulkResult {
  return { succeeded: r?.succeeded ?? [], failed: r?.failed ?? [] }
}

async function post<Resource extends object>(
  path: string,
  selector: BulkSelector,
  resource: Resource,
): Promise<BulkResult> {
  return normalize(await api.post<BulkResult>(path, withSelector(selector, resource)))
}

/** Send `resource` to every selected tenant against `path` (a full `/platform/**\/bulk/**` route). */
export function broadcast<Resource extends object>(
  path: string,
  selector: BulkSelector,
  resource: Resource,
): Promise<BulkResult> {
  return post(path, selector, resource)
}

/**
 * Path constants keep the caller sites honest — mistyping one is a compile
 * error rather than a 404 discovered in production.
 */
export const BULK_PATHS = {
  pipelines: {
    create: '/platform/eventprocessing/pipelines/bulk/create',
    update: '/platform/eventprocessing/pipelines/bulk/update',
    delete: '/platform/eventprocessing/pipelines/bulk/delete',
    activate: '/platform/eventprocessing/pipelines/bulk/activate',
  },
  correlationRules: {
    create: '/platform/eventprocessing/correlation-rule/bulk/create',
    update: '/platform/eventprocessing/correlation-rule/bulk/update',
    delete: '/platform/eventprocessing/correlation-rule/bulk/delete',
    activate: '/platform/eventprocessing/correlation-rule/bulk/activate',
  },
  soarRules: {
    create: '/platform/soar/rules/bulk/create',
    update: '/platform/soar/rules/bulk/update',
    delete: '/platform/soar/rules/bulk/delete',
    enable: '/platform/soar/rules/bulk/enable',
  },
  compliance: {
    frameworkCreate: '/platform/compliance/frameworks/bulk/create',
    frameworkUpdate: '/platform/compliance/frameworks/bulk/update',
    frameworkDelete: '/platform/compliance/frameworks/bulk/delete',
    controlCreate: '/platform/compliance/controls/bulk/create',
    controlUpdate: '/platform/compliance/controls/bulk/update',
    controlDelete: '/platform/compliance/controls/bulk/delete',
  },
  smtp: {
    update: '/platform/config/smtp/bulk/update',
    test: '/platform/config/smtp/bulk/test',
  },
  idp: {
    create: '/platform/identity-providers/bulk/create',
    update: '/platform/identity-providers/bulk/update',
    delete: '/platform/identity-providers/bulk/delete',
  },
  branding: {
    update: '/platform/branding/bulk/update',
    uploadAsset: (slot: string) => `/platform/branding/bulk/upload-asset/${slot}`,
  },
} as const

/**
 * Upload a single asset and broadcast the resulting URL — the branding
 * endpoint is multipart because the file is uploaded once and every tenant
 * points at the same object.
 */
export async function broadcastBrandingAsset(
  slot: string,
  file: File,
  selector: BulkSelector,
): Promise<BulkResult> {
  const form = new FormData()
  form.append('file', file)
  form.append('selector', JSON.stringify(selector))
  return normalize(await api.post<BulkResult>(BULK_PATHS.branding.uploadAsset(slot), form))
}
