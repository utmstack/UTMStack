package tenancy

import (
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const tenantFieldName = "TenantID"

var ErrNoTenant = errors.New("tenancy: operation on a tenant-scoped model without a tenant in context")

type SystemScoped interface {
	SystemFlagColumn() string
}

type allTenantsKey struct{}

func WithAllTenants(ctx context.Context) context.Context {
	return context.WithValue(ctx, allTenantsKey{}, true)
}

func spansAllTenants(ctx context.Context) bool {
	v, _ := ctx.Value(allTenantsKey{}).(bool)
	return v
}

func Register(db *gorm.DB, enabled func() bool) error {
	if enabled == nil {
		return errors.New("tenancy: Register needs an enabled predicate")
	}
	multiTenant = enabled

	c := db.Callback()

	if err := c.Query().Before("gorm:query").Register("tenancy:query", restrictRead); err != nil {
		return err
	}
	if err := c.Update().Before("gorm:update").Register("tenancy:update", restrictWrite); err != nil {
		return err
	}
	if err := c.Delete().Before("gorm:delete").Register("tenancy:delete", restrictWrite); err != nil {
		return err
	}
	if err := c.Row().Before("gorm:row").Register("tenancy:row", restrictRead); err != nil {
		return err
	}

	return c.Create().Before("gorm:create").Register("tenancy:create", stamp)
}

var multiTenant = func() bool { return false }

func tenantField(db *gorm.DB) *schema.Field {
	if !multiTenant() {
		return nil
	}
	if db.Statement == nil || db.Statement.Schema == nil {
		return nil
	}
	return db.Statement.Schema.LookUpField(tenantFieldName)
}

func tenantFor(db *gorm.DB) (string, bool) {
	ctx := db.Statement.Context
	if ctx == nil {
		db.AddError(ErrNoTenant)
		return "", false
	}

	if spansAllTenants(ctx) {
		return "", false
	}

	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		db.AddError(ErrNoTenant)
		return "", false
	}

	return tenant, true
}

func restrictRead(db *gorm.DB) { restrict(db, true) }

func restrictWrite(db *gorm.DB) { restrict(db, false) }

func restrict(db *gorm.DB, includeSystem bool) {
	field := tenantField(db)
	if field == nil {
		return
	}

	tenant, ok := tenantFor(db)
	if !ok {
		return
	}

	owned := clause.Expression(clause.Eq{
		Column: clause.Column{Table: db.Statement.Table, Name: field.DBName},
		Value:  tenant,
	})

	if col := systemColumn(db); includeSystem && col != "" {
		owned = clause.Or(owned, clause.Eq{
			Column: clause.Column{Table: db.Statement.Table, Name: col},
			Value:  true,
		})
	}

	db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{owned}})
}

func systemColumn(db *gorm.DB) string {
	if db.Statement.Schema == nil || db.Statement.Schema.ModelType == nil {
		return ""
	}
	model, ok := reflect.New(db.Statement.Schema.ModelType).Interface().(SystemScoped)
	if !ok {
		return ""
	}
	return model.SystemFlagColumn()
}

func stamp(db *gorm.DB) {
	field := tenantField(db)
	if field == nil {
		return
	}

	tenant, ok := tenantFor(db)
	if !ok {
		return
	}

	value := db.Statement.ReflectValue
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for i := range value.Len() {
			if err := field.Set(db.Statement.Context, value.Index(i), tenant); err != nil {
				db.AddError(err)
				return
			}
		}
	case reflect.Struct:
		if err := field.Set(db.Statement.Context, value, tenant); err != nil {
			db.AddError(err)
		}
	}
}
