package connectors

import "context"

type IndexPatternToggler interface {
	SetActiveByModule(ctx context.Context, moduleName string, active bool) error
}

type LogstashFilterToggler interface {
	SetActiveByModule(ctx context.Context, moduleName string, active bool) error
}

type MenuToggler interface {
	SetActiveByModule(ctx context.Context, moduleName string, active bool) error
}
