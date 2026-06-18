package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Postgres SQLSTATE codes mapped to domain-level sentinels.
const (
	pgCodeUniqueViolation     = "23505"
	pgCodeForeignKeyViolation = "23503"
)

// mapWriteError translates raw driver errors from INSERT/UPSERT paths into the
// generic sentinels, so every service doesn't have to sniff PgError codes (which
// is what makes duplicate-key paths wrongly return 500 instead of 409).
func mapWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgCodeUniqueViolation:
			return ErrAlreadyExists
		case pgCodeForeignKeyViolation:
			return ErrForeignKeyViolation
		}
	}
	return err
}

// DB is the GORM/PostgreSQL implementation of Database. It wraps an existing
// *gorm.DB — PanelMigration creates the connection in its bootstrap, so use
// New to adapt it rather than opening a second pool.
type DB struct {
	conn *gorm.DB
}

var _ Database = (*DB)(nil)

// New wraps an existing GORM connection in the Database provider.
func New(conn *gorm.DB) *DB { return &DB{conn: conn} }

// GORM exposes the underlying connection for paths the interface doesn't cover.
// Use sparingly — prefer the Database methods.
func (db *DB) GORM() *gorm.DB { return db.conn }

func (db *DB) Create(ctx context.Context, entity any) error {
	return mapWriteError(db.conn.WithContext(ctx).Create(entity).Error)
}

func (db *DB) Update(ctx context.Context, entity any) error {
	return mapWriteError(db.conn.WithContext(ctx).Save(entity).Error)
}

func (db *DB) Delete(ctx context.Context, entity any) error {
	return db.conn.WithContext(ctx).Delete(entity).Error
}

func (db *DB) DeleteWhere(ctx context.Context, model any, opts ...QueryOption) error {
	return Apply(db.conn.WithContext(ctx), opts...).Delete(model).Error
}

func (db *DB) FindByID(ctx context.Context, dest any, id any) error {
	err := db.conn.WithContext(ctx).First(dest, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (db *DB) FindOne(ctx context.Context, dest any, opts ...QueryOption) error {
	err := Apply(db.conn.WithContext(ctx), opts...).First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (db *DB) FindAll(ctx context.Context, dest any, opts ...QueryOption) error {
	return Apply(db.conn.WithContext(ctx), opts...).Find(dest).Error
}

func (db *DB) Count(ctx context.Context, model any, opts ...QueryOption) (int64, error) {
	var count int64
	err := Apply(db.conn.WithContext(ctx).Model(model), opts...).Count(&count).Error
	return count, err
}

func (db *DB) Upsert(ctx context.Context, entity any, conflictColumns []string) error {
	columns := make([]clause.Column, len(conflictColumns))
	for i, col := range conflictColumns {
		columns[i] = clause.Column{Name: col}
	}
	return mapWriteError(db.conn.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   columns,
		UpdateAll: true,
	}).Create(entity).Error)
}

func (db *DB) Transaction(ctx context.Context, fn func(tx Database) error) error {
	return db.conn.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		return fn(&DB{conn: gormTx})
	})
}

func (db *DB) Raw(ctx context.Context, dest any, query string, args ...any) error {
	return db.conn.WithContext(ctx).Raw(query, args...).Scan(dest).Error
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) error {
	return db.conn.WithContext(ctx).Exec(query, args...).Error
}
