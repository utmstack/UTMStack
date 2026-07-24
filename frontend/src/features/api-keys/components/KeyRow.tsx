import { useTranslation } from 'react-i18next'
import { Globe2, KeyRound, Pencil, RefreshCw, Trash2 } from 'lucide-react'
import type { ApiKey } from '../types/api-key.types'
import { IconAction } from './IconAction'
import { StatusBadge } from './StatusBadge'

export const COLS = '1.4fr 1.2fr 110px 110px 110px 100px 110px'

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: '2-digit' })
}

export function KeyRow({
  apiKey: k,
  onEdit,
  onRotate,
  onDelete,
}: {
  apiKey: ApiKey
  onEdit: () => void
  onRotate: () => void
  onDelete: () => void
}) {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-3 border-b border-border px-4 py-3 text-xs last:border-b-0 hover:bg-muted/30"
      style={{ gridTemplateColumns: COLS }}
    >
      <div className="flex min-w-0 items-center gap-2">
        <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
          <KeyRound size={13} />
        </span>
        <span className="truncate font-medium">{k.name}</span>
      </div>
      <div className="min-w-0 truncate text-[11px] text-muted-foreground">
        {k.allowed_ip.length === 0 ? (
          <span className="inline-flex items-center gap-1">
            <Globe2 size={11} /> {t('apiKeys.anyIp')}
          </span>
        ) : (
          <span className="font-mono" title={k.allowed_ip.join(', ')}>
            {k.allowed_ip.join(', ')}
          </span>
        )}
      </div>
      <div className="text-[11px] text-muted-foreground">{formatDate(k.created_at)}</div>
      <div className="text-[11px] text-muted-foreground">{formatDate(k.generated_at)}</div>
      <div className="text-[11px] text-muted-foreground">
        {k.expires_at ? formatDate(k.expires_at) : <span className="text-muted-foreground/60">{t('apiKeys.never')}</span>}
      </div>
      <div>
        <StatusBadge expiresAt={k.expires_at} />
      </div>
      <div className="flex items-center justify-end gap-1">
        <IconAction label={t('apiKeys.action.rotate')} onClick={onRotate}>
          <RefreshCw size={13} />
        </IconAction>
        <IconAction label={t('apiKeys.action.edit')} onClick={onEdit}>
          <Pencil size={13} />
        </IconAction>
        <IconAction label={t('apiKeys.action.delete')} danger onClick={onDelete}>
          <Trash2 size={13} />
        </IconAction>
      </div>
    </div>
  )
}
