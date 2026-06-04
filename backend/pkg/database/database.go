package database

import (
	"context"
	"errors"
)

// Sentinel errors the provider maps raw driver errors into, so callers can
// errors.Is against them instead of sniffing PgError codes / parsing strings.
var (
	ErrNotFound            = errors.New("record not found")
	ErrAlreadyExists       = errors.New("record already exists")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Database is the data-access port. The GORM/Postgres provider (DB) implements
// it; services can depend on this interface to swap in a fake for tests.
type Database interface {
	Create(ctx context.Context, entity any) error
	Update(ctx context.Context, entity any) error
	Delete(ctx context.Context, entity any) error
	DeleteWhere(ctx context.Context, model any, opts ...QueryOption) error

	FindByID(ctx context.Context, dest any, id any) error
	FindOne(ctx context.Context, dest any, opts ...QueryOption) error
	FindAll(ctx context.Context, dest any, opts ...QueryOption) error
	Count(ctx context.Context, model any, opts ...QueryOption) (int64, error)

	Upsert(ctx context.Context, entity any, conflictColumns []string) error
	Transaction(ctx context.Context, fn func(tx Database) error) error

	Raw(ctx context.Context, dest any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) error
}
