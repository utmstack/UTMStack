package association

import (
	"sync"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/entities"
)

type AssociationBuilder struct {
	rules          []*AssociationRule
	entityRegistry *sync.Map
	mu             sync.RWMutex
}

func NewAssociationBuilder() *AssociationBuilder {
	builder := &AssociationBuilder{
		rules:          GetEnabledRules(),
		entityRegistry: &sync.Map{},
	}
	catcher.Info("association builder initialized", map[string]any{
		"total_rules": len(builder.rules),
	})
	return builder
}

func (b *AssociationBuilder) RegisterEntity(entity *entities.Entity, entityID, sourcePath string, ctx AssociationContext) {
	ref := &EntityReference{
		Entity:     entity,
		EntityID:   entityID,
		EntityType: entity.Type,
		SourcePath: sourcePath,
		Context:    ctx,
	}
	b.entityRegistry.Store(entityID, ref)
	catcher.Info("entity registered", map[string]any{
		"entity_id":   entityID,
		"entity_type": entity.Type,
		"source_path": sourcePath,
	})
}

func (b *AssociationBuilder) BuildAssociations() []*entities.Entity {
	contextGroups := b.groupByContext()
	for _, refs := range contextGroups {
		b.detectAssociationsInContext(refs)
	}
	result := make([]*entities.Entity, 0)
	b.entityRegistry.Range(func(key, value any) bool {
		if ref, ok := value.(*EntityReference); ok {
			if entity, ok := ref.Entity.(*entities.Entity); ok {
				result = append(result, entity)
			}
		}
		return true
	})
	catcher.Info("associations built", map[string]any{
		"total_entities":     len(result),
		"context_groups":     len(contextGroups),
		"total_associations": b.CountAssociations(result),
	})
	return result
}

func (b *AssociationBuilder) groupByContext() map[string][]*EntityReference {
	groups := make(map[string][]*EntityReference)
	b.entityRegistry.Range(func(key, value any) bool {
		if ref, ok := value.(*EntityReference); ok {
			contextKey := ref.Context.AlertID
			if contextKey != "" {
				groups[contextKey] = append(groups[contextKey], ref)
			}
		}
		return true
	})
	return groups
}

func (b *AssociationBuilder) detectAssociationsInContext(refs []*EntityReference) {
	for _, rule := range b.rules {
		if !rule.Enabled {
			continue
		}
		for i, sourceRef := range refs {
			if sourceRef.EntityType != rule.SourceType {
				continue
			}
			for j, targetRef := range refs {
				if i == j {
					continue
				}
				if targetRef.EntityType != rule.TargetType {
					continue
				}
				if b.shouldCreateAssociation(sourceRef, targetRef) {
					b.createAssociation(sourceRef, targetRef, rule)
				}
			}
		}
	}
}

func (b *AssociationBuilder) shouldCreateAssociation(source, target *EntityReference) bool {
	if source.Context.SameContext(target.Context) {
		return true
	}
	if source.Context.IsOriginToTarget(target.Context) {
		return true
	}
	if source.Context.IsTargetToOrigin(target.Context) {
		return true
	}
	return false
}

func (b *AssociationBuilder) createAssociation(source, target *EntityReference, rule *AssociationRule) {
	sourceEntity, ok := source.Entity.(*entities.Entity)
	if !ok {
		return
	}
	targetEntity, ok := target.Entity.(*entities.Entity)
	if !ok {
		return
	}
	associatedEntity := entities.EntityAssociation{
		Mode: string(rule.Mode),
		Entity: entities.Entity{
			Type:       targetEntity.Type,
			Attributes: targetEntity.Attributes,
		},
	}
	if sourceEntity.Associations == nil {
		sourceEntity.Associations = make([]entities.EntityAssociation, 0)
	}
	sourceEntity.Associations = append(sourceEntity.Associations, associatedEntity)
	catcher.Info("association created", map[string]any{
		"rule":        rule.Name,
		"source_type": sourceEntity.Type,
		"target_type": targetEntity.Type,
		"mode":        rule.Mode,
		"context":     source.Context.AlertID,
	})
}

func (b *AssociationBuilder) CountAssociations(entities []*entities.Entity) int {
	count := 0
	for _, entity := range entities {
		count += len(entity.Associations)
	}
	return count
}

func (b *AssociationBuilder) ClearRegistry() {
	b.entityRegistry = &sync.Map{}
	catcher.Info("entity registry cleared", nil)
}
