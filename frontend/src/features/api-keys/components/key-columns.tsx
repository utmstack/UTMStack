import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { Globe2, KeyRound, Pencil, RefreshCw, Trash2 } from 'lucide-react'
import type { ApiKey } from '../types/api-key.types'
import { IconAction } from './IconAction'
import { StatusBadge } from './StatusBadge'

const TH = 'whitespace-nowrap px-3 py-2 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-3 align-middle'

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: '2-digit' })
}

interface KeyActions {
  onEdit: (key: ApiKey) => void
  onRotate: (key: ApiKey) => void
  onDelete: (key: ApiKey) => void
}

export function buildKeyColumns(t: TFunction, { onEdit, onRotate, onDelete }: KeyActions): ColumnDef<ApiKey>[] {
  const muted = `${TD} text-[11px] text-muted-foreground`
  return [
    {
      id: 'name',
      header: t('apiKeys.col.name'),
      size: 320,
      minSize: 120,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <div className="flex min-w-0 items-center gap-2">
          <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <KeyRound size={13} />
          </span>
          <span className="truncate font-medium">{row.original.name}</span>
        </div>
      ),
    },
    {
      id: 'allowedIps',
      header: t('apiKeys.col.allowedIps'),
      size: 260,
      minSize: 100,
      meta: { headerClassName: TH, cellClassName: muted, cellProps: (k) => ({ title: k.allowed_ip.join(', ') }) },
      cell: ({ row }) =>
        row.original.allowed_ip.length === 0 ? (
          <span className="inline-flex items-center gap-1">
            <Globe2 size={11} /> {t('apiKeys.anyIp')}
          </span>
        ) : (
          <span className="block truncate font-mono">{row.original.allowed_ip.join(', ')}</span>
        ),
    },
    {
      id: 'created',
      header: t('apiKeys.col.created'),
      size: 120,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: muted },
      cell: ({ row }) => formatDate(row.original.created_at),
    },
    {
      id: 'lastRotated',
      header: t('apiKeys.col.lastRotated'),
      size: 130,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: muted },
      cell: ({ row }) => formatDate(row.original.generated_at),
    },
    {
      id: 'expires',
      header: t('apiKeys.col.expires'),
      size: 120,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: muted },
      cell: ({ row }) =>
        row.original.expires_at ? (
          formatDate(row.original.expires_at)
        ) : (
          <span className="text-muted-foreground/60">{t('apiKeys.never')}</span>
        ),
    },
    {
      id: 'status',
      header: t('apiKeys.col.status'),
      size: 110,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <StatusBadge expiresAt={row.original.expires_at} />,
    },
    {
      id: 'actions',
      header: t('apiKeys.col.actions'),
      size: 128,
      minSize: 128,
      enableResizing: false,
      meta: { headerClassName: `${TH} text-right`, cellClassName: TD },
      cell: ({ row }) => (
        <div className="flex items-center justify-end gap-1">
          <IconAction label={t('apiKeys.action.rotate')} onClick={() => onRotate(row.original)}>
            <RefreshCw size={13} />
          </IconAction>
          <IconAction label={t('apiKeys.action.edit')} onClick={() => onEdit(row.original)}>
            <Pencil size={13} />
          </IconAction>
          <IconAction label={t('apiKeys.action.delete')} danger onClick={() => onDelete(row.original)}>
            <Trash2 size={13} />
          </IconAction>
        </div>
      ),
    },
  ]
}
