import type { TFunction } from 'i18next'
import { deriveStatus } from './status'
import { datasourcesHttpService as svc } from '../services/datasources-http.service'
import type { Datasource } from '../types/datasource.types'

const esc = (v: unknown) =>
  String(v ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

function fmtBytes(n?: number) {
  if (!n) return '—'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (n >= 1024 && i < u.length - 1) { n /= 1024; i++ }
  return `${n.toFixed(n >= 100 || i === 0 ? 0 : 1)} ${u[i]}`
}

/** Build the full report HTML (all sources, health, ingestion). */
async function buildReportHtml(t: TFunction, events24h: (name: string) => number) {
  const all: Datasource[] = []
  for (let page = 1; ; page++) {
    const res = await svc.list({ page, size: 500, sort: 'last_ping_at.desc' })
    all.push(...(res.items ?? []))
    if (all.length >= (res.total_items ?? 0) || page >= (res.total_pages ?? 0)) break
  }
  const [countRes, totalsRes] = await Promise.allSettled([svc.count(), svc.ingestionTotals()])
  const total = countRes.status === 'fulfilled' ? countRes.value.count : all.length
  const ingestion = totalsRes.status === 'fulfilled' ? totalsRes.value : null

  const offline = all.filter((s) => deriveStatus(s.lastPingAt) === 'offline').length
  const online = total - offline
  const byKind = (k: string) => all.filter((s) => s.sourceKind === k).length
  const fmtTime = (iso?: string) => (iso ? new Date(iso).toLocaleString() : '—')

  const rows = all
    .map((s) => {
      const st = deriveStatus(s.lastPingAt)
      const ev = events24h(s.name)
      return `<tr>
        <td>${esc(s.name)}</td>
        <td>${esc(s.dataType || s.sourceKind || '—')}</td>
        <td>${esc(s.sourceKind || '—')}</td>
        <td class="${st}">${t(`datasources.status.${st}`)}</td>
        <td class="num">${ev > 0 ? ev.toLocaleString() : '—'}</td>
        <td>${esc(s.ip || '—')}</td>
        <td>${fmtTime(s.lastPingAt)}</td>
      </tr>`
    })
    .join('\n')

  return `<!doctype html><html><head><meta charset="utf-8"><title>UTMStack Data Sources Report</title>
<style>
  * { box-sizing: border-box; }
  body { font-family: -apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1a1a2e; margin: 32px; font-size: 12px; }
  h1 { font-size: 20px; margin: 0; }
  .sub { color: #666; margin-top: 4px; }
  .stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin: 20px 0; }
  .stat { border: 1px solid #ddd; border-radius: 8px; padding: 10px 12px; }
  .stat .v { font-size: 18px; font-weight: 600; }
  .stat .l { color: #666; font-size: 11px; margin-top: 2px; }
  table { width: 100%; border-collapse: collapse; margin-top: 8px; }
  th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid #eee; white-space: nowrap; }
  th { background: #f5f5f7; font-size: 11px; text-transform: uppercase; letter-spacing: .03em; }
  td.num { text-align: right; font-variant-numeric: tabular-nums; }
  td.live { color: #059669; } td.idle { color: #d97706; } td.offline { color: #dc2626; }
  .foot { margin-top: 16px; color: #999; font-size: 10px; }
  .bar { position: sticky; top: 0; background: #fff; display: flex; gap: 8px; padding: 12px 0; margin: -32px -32px 0; padding: 12px 32px 0; border-bottom: 1px solid #eee; }
  .bar button { border: 1px solid #333; background: #333; color: #fff; padding: 6px 14px; border-radius: 6px; cursor: pointer; font-size: 12px; }
  @media print { .bar { display: none; } body { margin: 0; } tr { page-break-inside: avoid; } }
</style></head><body>
<div class="bar"><button onclick="window.print()">${t('datasources.report.print')}</button></div>
<h1>${t('datasources.report.title')}</h1>
<div class="sub">${new Date().toLocaleString()} &middot; ${t('datasources.report.generatedBy')}</div>
<div class="stats">
  <div class="stat"><div class="v">${total.toLocaleString()}</div><div class="l">${t('datasources.health.total')}</div></div>
  <div class="stat"><div class="v">${online.toLocaleString()}</div><div class="l">${t('datasources.health.online')}</div></div>
  <div class="stat"><div class="v">${offline.toLocaleString()}</div><div class="l">${t('datasources.health.offline')}</div></div>
  <div class="stat"><div class="v">${(ingestion?.total ?? 0).toLocaleString()}</div><div class="l">${t('datasources.report.eventsReceived')}</div></div>
  <div class="stat"><div class="v">${byKind('agent').toLocaleString()}</div><div class="l">${t('datasources.report.agents')}</div></div>
  <div class="stat"><div class="v">${byKind('puller').toLocaleString()}</div><div class="l">${t('datasources.report.integrations')}</div></div>
  <div class="stat"><div class="v">${byKind('direct').toLocaleString()}</div><div class="l">${t('datasources.report.direct')}</div></div>
  <div class="stat"><div class="v">${fmtBytes(ingestion?.totalBytes)}</div><div class="l">${t('datasources.report.volume')}</div></div>
</div>
<h2 style="font-size:14px;margin:16px 0 0">${t('datasources.report.list')}</h2>
<table>
<thead><tr>
  <th>${t('datasources.cols.source')}</th>
  <th>${t('datasources.cols.type')}</th>
  <th>Kind</th>
  <th>${t('datasources.cols.status')}</th>
  <th>${t('datasources.cols.events24h')}</th>
  <th>IP</th>
  <th>${t('datasources.cols.lastSeen')}</th>
</tr></thead>
<tbody>${rows || `<tr><td colspan="7">—</td></tr>`}</tbody>
</table>
<div class="foot">${all.length} / ${total} ${t('datasources.report.rowsNote')}</div>
</body></html>`
}

/** Open a report overlay (iframe + backdrop). "Print / Save as PDF" button inside;
 * auto-triggers the browser print dialog. No PDF lib — browser does it natively. */
export async function exportReportPdf(t: TFunction, events24h: (name: string) => number) {
  const backdrop = document.createElement('div')
  // Offscreen: the iframe is never shown — it only feeds the print dialog.
  backdrop.style.cssText = 'position:fixed; left:-99999px; top:0; width:0; height:0; overflow:hidden'
  const bar = document.createElement('div')
  bar.className = 'flex items-center justify-between gap-3 bg-card px-4 py-2 text-sm'
  const title = document.createElement('span')
  title.textContent = t('datasources.report.title')
  const closeBtn = document.createElement('button')
  closeBtn.className = 'rounded-md px-3 py-1 text-xs text-muted-foreground hover:bg-muted hover:text-foreground'
  closeBtn.textContent = t('datasources.report.close')
  const frame = document.createElement('iframe')
  frame.className = 'm-2 flex-1 rounded-lg border border-border bg-white'
  frame.style.cssText = 'width:900px; height:1200px'
  frame.title = 'Report'
  bar.append(title, closeBtn)
  backdrop.append(bar, frame)
  document.body.appendChild(backdrop)

  const remove = () => backdrop.remove()
  closeBtn.addEventListener('click', remove)

  const html = await buildReportHtml(t, events24h)
  frame.onload = () => {
    const win = frame.contentWindow
    // Close the overlay when the print dialog is dismissed.
    win?.addEventListener('afterprint', remove)
    try {
      win?.print()
    } catch {
      // same-origin srcdoc — print allowed; ignore
    }
  }
  frame.srcdoc = html
}
