package association

type ContextType string

type AssociationContext struct {
	AlertID     string
	IncidentID  string
	SourceField string
}

type EntityReference struct {
	Entity     any
	EntityID   string
	EntityType string
	SourcePath string
	Context    AssociationContext
}

func (ctx *AssociationContext) IsOrigin() bool {
	return ctx.SourceField == "source"
}

func (ctx *AssociationContext) IsTarget() bool {
	return ctx.SourceField == "destination"
}

func (ctx *AssociationContext) SameContext(other AssociationContext) bool {
	if ctx.AlertID != "" && ctx.AlertID == other.AlertID {
		return true
	}
	return false
}

func (ctx *AssociationContext) IsOriginToTarget(other AssociationContext) bool {
	return ctx.IsOrigin() && other.IsTarget() && ctx.SameContext(other)
}

func (ctx *AssociationContext) IsTargetToOrigin(other AssociationContext) bool {
	return ctx.IsTarget() && other.IsOrigin() && ctx.SameContext(other)
}
