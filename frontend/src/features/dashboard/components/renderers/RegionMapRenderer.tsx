import { useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Map as MapIcon } from 'lucide-react'
import L, { type LatLngExpression } from 'leaflet'
import 'leaflet/dist/leaflet.css'
import {
  MapContainer,
  Marker,
  Popup,
  TileLayer,
  useMap,
} from 'react-leaflet'

type Row = Record<string, unknown>

interface Point {
  name: string
  lat: number
  lon: number
  value?: number
}

const TILE_URL = 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
const TILE_ATTRIBUTION =
  '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'

const pinIcon = L.divIcon({
  className: 'utmstack-leaflet-pin',
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  html: '<span class="block h-3.5 w-3.5 rounded-full border-2 border-white bg-red-500 shadow-md"></span>',
})

export function RegionMapRenderer({ rows }: { rows: Row[] }) {
  const { t } = useTranslation()
  const points = useMemo(() => parsePoints(rows), [rows])

  if (points.length === 0) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-2 px-4 text-center text-xs text-muted-foreground">
        <MapIcon size={20} className="text-muted-foreground/60" />
        <span>{t('dashboards.widget.regionMapEmpty')}</span>
        <span className="max-w-xs text-[10px] text-muted-foreground/70">
          {t('dashboards.widget.regionMapShapeHint')}
        </span>
      </div>
    )
  }

  const center: LatLngExpression = [points[0].lat, points[0].lon]

  return (
    <MapContainer
      center={center}
      zoom={2}
      scrollWheelZoom
      className="h-full w-full overflow-hidden rounded-md"
      style={{ height: '100%', width: '100%' }}
    >
      <TileLayer url={TILE_URL} attribution={TILE_ATTRIBUTION} />
      <FitToPoints points={points} />
      {points.map((p, i) => (
        <Marker key={`${p.lat}-${p.lon}-${i}`} position={[p.lat, p.lon]} icon={pinIcon}>
          <Popup>
            <div className="text-xs">
              <div className="font-semibold">{p.name || t('dashboards.widget.regionMapUnnamed')}</div>
              {p.value != null && (
                <div className="text-muted-foreground">
                  {t('dashboards.widget.regionMapValue', { value: p.value })}
                </div>
              )}
              <div className="text-muted-foreground">
                {p.lat.toFixed(4)}, {p.lon.toFixed(4)}
              </div>
            </div>
          </Popup>
        </Marker>
      ))}
    </MapContainer>
  )
}

function FitToPoints({ points }: { points: Point[] }) {
  const map = useMap()
  useEffect(() => {
    if (points.length === 0) return
    if (points.length === 1) {
      map.setView([points[0].lat, points[0].lon], 5)
      return
    }
    const bounds = L.latLngBounds(points.map((p) => [p.lat, p.lon]))
    map.fitBounds(bounds, { padding: [24, 24], maxZoom: 10 })
  }, [map, points])
  return null
}

function parsePoints(rows: Row[]): Point[] {
  return rows
    .map((r) => parseRow(r))
    .filter((p): p is Point => p != null)
}

function parseRow(r: Row): Point | null {
  if (Array.isArray((r as { value?: unknown }).value)) {
    const arr = (r as { value: unknown[] }).value
    const lat = toNumber(arr[0])
    const lon = toNumber(arr[1])
    const value = arr[2] != null ? toNumber(arr[2]) : undefined
    const name = typeof (r as { name?: unknown }).name === 'string'
      ? ((r as { name: string }).name)
      : ''
    if (!isLatLon(lat, lon)) return null
    return { name, lat, lon, value: Number.isFinite(value as number) ? (value as number) : undefined }
  }

  const keys = Object.keys(r)
  const latKey = keys.find((k) => /^lat(itude)?$/i.test(k))
  const lonKey = keys.find((k) => /^(lon|lng|longitude)$/i.test(k))
  if (latKey == null || lonKey == null) return null
  const lat = toNumber(r[latKey])
  const lon = toNumber(r[lonKey])
  if (!isLatLon(lat, lon)) return null

  const nameKey = keys.find((k) => /^(name|label|host|ip|country|region|city)$/i.test(k))
  const valueKey = keys.find((k) => /^(value|count|y|total)$/i.test(k))
  const name = nameKey ? String(r[nameKey] ?? '') : ''
  const value = valueKey != null ? toNumber(r[valueKey]) : undefined
  return {
    name,
    lat,
    lon,
    value: Number.isFinite(value as number) ? (value as number) : undefined,
  }
}

function toNumber(value: unknown): number {
  if (typeof value === 'number') return value
  if (typeof value === 'string' && value.trim()) return Number(value)
  return Number.NaN
}

function isLatLon(lat: number, lon: number): boolean {
  return (
    Number.isFinite(lat) &&
    Number.isFinite(lon) &&
    lat >= -90 &&
    lat <= 90 &&
    lon >= -180 &&
    lon <= 180
  )
}
