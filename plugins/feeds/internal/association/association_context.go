package association

type AssociationContext struct {
	AlertID     string
	EventID     string
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
	return ctx.SourceField == "origin"
}

func (ctx *AssociationContext) IsTarget() bool {
	return ctx.SourceField == "target"
}

func (ctx *AssociationContext) SameAlert(other AssociationContext) bool {
	return ctx.AlertID != "" && ctx.AlertID == other.AlertID
}

func (ctx *AssociationContext) SameEvent(other AssociationContext) bool {
	return ctx.EventID != "" && ctx.EventID == other.EventID
}

func (ctx *AssociationContext) IsOriginToTarget(other AssociationContext) bool {
	return ctx.SameEvent(other) && ctx.IsOrigin() && other.IsTarget()
}

func (ctx *AssociationContext) CrossEventAssociation(other AssociationContext) bool {
	return ctx.SameAlert(other) && ctx.EventID != "" && other.EventID != "" && ctx.EventID != other.EventID
}
