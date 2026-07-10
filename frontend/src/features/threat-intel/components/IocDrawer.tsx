import { useTiEntity } from '../hooks/use-ti-entity'
import { useTiEntityRelations } from '../hooks/use-ti-entity-relations'
import { IocDrawerBody } from './IocDrawerBody'
import { IocDrawerLoading } from './IocDrawerLoading'

interface IocDrawerProps {
  id: string | null
  onClose: () => void
}

export function IocDrawer({ id, onClose }: IocDrawerProps) {
  const { data: entityData, isLoading: entityLoading } = useTiEntity(id ?? '')
  const { data: relationsData } = useTiEntityRelations(id ?? '')

  if (!id) return null
  if (entityData?.kind === 'not-configured') return null

  const detail = entityData?.kind === 'ok' ? entityData.value : null
  const relations = relationsData?.kind === 'ok' ? relationsData.value : []

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {entityLoading || !detail ? (
          <IocDrawerLoading onClose={onClose} />
        ) : (
          <IocDrawerBody detail={detail} relations={relations} onClose={onClose} />
        )}
      </div>
    </div>
  )
}
