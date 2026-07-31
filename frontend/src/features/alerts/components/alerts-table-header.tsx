import { useTranslation } from 'react-i18next'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'

export function AlertsTableHeader({ allChecked, onTogglePage }: { allChecked: boolean; onTogglePage: () => void }) {
  const { t } = useTranslation()
  return (
    <thead className="sticky top-0 z-10 bg-muted/90 text-[10px] uppercase tracking-wider text-muted-foreground">
      <tr className="border-b border-border">
        <th className={`${TH} w-[6px] p-0`} aria-hidden />
        <th className={`${TH} w-px`}>
          <button onClick={onTogglePage} className="flex h-4 w-4 items-center justify-center rounded border border-input">
            {allChecked && <span className="h-2 w-2 rounded-sm bg-primary" />}
          </button>
        </th>
        <th className={`${TH} text-center`} colSpan={2}>{t('alerts.table.actions')}</th>
        <th className={TH}>{t('alerts.table.alert')}</th>
        <th className={TH}>{t('alerts.table.status')}</th>
        <th className={TH}>{t('alerts.table.technique')}</th>
        <th className={TH}>{t('alerts.table.sourceAdversary')}</th>
        <th className={`${TH} text-center`}>{t('alerts.table.severity')}</th>
        <th className={`${TH} text-center`}>{t('alerts.table.echoes')}</th>
        <th className={`${TH} text-center`}>{t('alerts.table.time')}</th>
      </tr>
    </thead>
  )
}

export const ALERTS_TABLE_COLUMN_COUNT = 11
