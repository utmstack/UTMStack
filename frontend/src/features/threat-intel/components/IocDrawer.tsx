import { useTiEntity } from '../hooks/use-ti-entity'
import type { EntityRelation } from '../domain/threat-intel.types'
import { useTiEntityRelations } from '../hooks/use-ti-entity-relations'
import type { RawRelationNode, RawRelations } from '../console/agent'
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
  // The relations endpoint returns a graph object ({nodes, edges}) with full
  // entity attributes, not a bare list. `EntityRelation = EntityAttributes` so
  // the drawer body can render value/reputation straight off each node.
  const rawRels = (relationsData?.kind === 'ok' ? relationsData.value : null) as unknown as RawRelations
  const relNodes: RawRelationNode[] = Array.isArray(rawRels) ? rawRels : rawRels?.nodes ?? []
  const relations: EntityRelation[] = relNodes.map((r) => ({ ...r, value: r.value ?? r.label ?? r.id }) as EntityRelation)

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
