package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sourceRow struct {
	Source string
	Count  int64
}

// syncBatchSize bounds what a full sync holds in memory at once.
const syncBatchSize = 1000

type pgADUserRepository struct{ db *gorm.DB }

func NewADUserRepository(db *gorm.DB) connectors.ADUserRepository {
	return &pgADUserRepository{db: db}
}

func splitByIdentity(users []domain.ADUser) (windows, linuxKeyed, linuxByAccount []domain.ADUser) {
	for _, u := range users {
		switch u.Source {
		case "linux":
			if u.MachineID != nil && *u.MachineID != "" && u.UIDNumber != nil {
				linuxKeyed = append(linuxKeyed, u)
			} else {
				linuxByAccount = append(linuxByAccount, u)
			}
		default:
			if u.Source == "" {
				u.Source = "windows"
			}
			windows = append(windows, u)
		}
	}
	return windows, linuxKeyed, linuxByAccount
}

func (r *pgADUserRepository) Upsert(ctx context.Context, users []domain.ADUser) error {
	if len(users) == 0 {
		return nil
	}

	windowsBatch, linuxKeyed, linuxByAccount := splitByIdentity(users)

	if len(windowsBatch) > 0 {
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:     []clause.Column{{Name: "tenant_id"}, {Name: "sid"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "source = 'windows'"}}},
			DoUpdates: clause.AssignmentColumns([]string{
				"sam_account_name", "domain", "active",
				"account_created_at", "last_logon", "account_deleted_at", "last_seen",
			}),
		}).Create(&windowsBatch).Error; err != nil {
			return err
		}
	}

	if len(linuxKeyed) > 0 {
		// An account first seen without its uid is already a row; give it the
		// uid instead of inserting a second one beside it.
		for i := range linuxKeyed {
			if err := r.adoptUID(ctx, &linuxKeyed[i]); err != nil {
				return err
			}
		}
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:     []clause.Column{{Name: "tenant_id"}, {Name: "machine_id"}, {Name: "uid_number"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "source = 'linux'"}}},
			DoUpdates: clause.AssignmentColumns([]string{
				"username", "hostname", "active",
				"account_created_at", "last_logon", "account_deleted_at", "last_seen",
			}),
		}).Create(&linuxKeyed).Error; err != nil {
			return err
		}
	}

	for i := range linuxByAccount {
		if err := r.upsertByAccount(ctx, &linuxByAccount[i]); err != nil {
			return err
		}
	}

	return nil
}

// adoptUID stamps the uid (and machine) onto the row an account already has
// without one, so the keyed upsert that follows updates it.
func (r *pgADUserRepository) adoptUID(ctx context.Context, u *domain.ADUser) error {
	if u.Hostname == nil || u.Username == nil {
		return nil
	}
	return r.db.WithContext(ctx).Exec(`
		UPDATE ad_user SET uid_number = ?, machine_id = ?
		WHERE id = (
		    SELECT id FROM ad_user
		    WHERE tenant_id = ? AND source = 'linux' AND hostname = ? AND username = ? AND uid_number IS NULL
		    ORDER BY (machine_id IS NOT NULL) DESC, last_seen DESC NULLS LAST, id
		    LIMIT 1
		)
		AND NOT EXISTS (
		    SELECT 1 FROM ad_user x
		    WHERE x.tenant_id = ? AND x.source = 'linux' AND x.machine_id = ? AND x.uid_number = ?
		)`,
		*u.UIDNumber, *u.MachineID, u.TenantID, *u.Hostname, *u.Username,
		u.TenantID, *u.MachineID, *u.UIDNumber,
	).Error
}

// upsertByAccount records an observation of a Linux account by who it is —
// host and user name — because the observation does not carry everything the
// unique index needs. The account keeps one row however many times it is seen,
// and whether or not its machine id has been resolved in the meantime.
func (r *pgADUserRepository) upsertByAccount(ctx context.Context, u *domain.ADUser) error {
	if u.Hostname == nil || u.Username == nil {
		return nil
	}
	var existing domain.ADUser
	err := r.db.WithContext(ctx).Where(
		"tenant_id = ? AND source = 'linux' AND hostname = ? AND username = ?",
		u.TenantID, *u.Hostname, *u.Username,
	).Order("(uid_number IS NOT NULL) DESC, last_seen DESC NULLS LAST, id").First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(u).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]any{
		"active":             u.Active,
		"account_created_at": u.AccountCreatedAt,
		"last_logon":         u.LastLogon,
		"account_deleted_at": u.AccountDeletedAt,
		"last_seen":          u.LastSeen,
	}
	// Filling in what the row lacks is only safe while the row has no machine
	// id yet: with one, the (machine, uid) pair could already belong to another.
	if existing.MachineID == nil {
		if u.UIDNumber != nil && existing.UIDNumber == nil {
			updates["uid_number"] = *u.UIDNumber
		}
		if u.MachineID != nil && *u.MachineID != "" {
			updates["machine_id"] = *u.MachineID
		}
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(updates).Error
}

const foldProvisionalSQL = `
WITH folded AS (
    DELETE FROM ad_user p
    USING ad_user r
    WHERE p.tenant_id = ? AND p.source = 'linux' AND p.hostname = ? AND p.machine_id IS NULL
      AND r.tenant_id = p.tenant_id AND r.source = 'linux' AND r.machine_id = ? AND r.username = p.username
    RETURNING r.id AS keep_id, p.last_seen, p.last_logon, p.account_created_at
)
UPDATE ad_user k SET
    last_seen = GREATEST(k.last_seen, f.last_seen),
    last_logon = GREATEST(k.last_logon, f.last_logon),
    account_created_at = LEAST(k.account_created_at, f.account_created_at)
FROM (
    SELECT keep_id, MAX(last_seen) AS last_seen, MAX(last_logon) AS last_logon, MIN(account_created_at) AS account_created_at
    FROM folded GROUP BY keep_id
) f
WHERE k.id = f.keep_id`

func (r *pgADUserRepository) ResolveLinuxIdentity(ctx context.Context, tenantID, hostname, machineID string) (int64, error) {
	if tenantID == "" || hostname == "" || machineID == "" {
		return 0, nil
	}
	var resolved int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(foldProvisionalSQL, tenantID, hostname, machineID).Error; err != nil {
			return err
		}
		result := tx.Exec(`
			UPDATE ad_user SET machine_id = ?
			WHERE tenant_id = ?
			  AND source = 'linux'
			  AND hostname = ?
			  AND machine_id IS NULL
			  AND NOT EXISTS (
			      SELECT 1 FROM ad_user r2
			      WHERE r2.tenant_id = ad_user.tenant_id
			        AND r2.source = 'linux'
			        AND r2.machine_id = ?
			        AND r2.uid_number = ad_user.uid_number
			  )`, machineID, tenantID, hostname, machineID)
		resolved = result.RowsAffected
		return result.Error
	})
	return resolved, err
}

func applyStatus(q *gorm.DB, status string) *gorm.DB {
	switch status {
	case "active":
		return q.Where("active = true AND account_deleted_at IS NULL")
	case "disabled":
		return q.Where("active = false AND account_deleted_at IS NULL")
	case "deleted":
		return q.Where("account_deleted_at IS NOT NULL")
	case "stale":
		return q.Where("active = true AND account_deleted_at IS NULL AND last_logon IS NOT NULL AND last_logon < NOW() - INTERVAL '30 days'")
	case "service":
		return q.Where("account_deleted_at IS NULL AND sam_account_name ILIKE 'svc%' AND source = 'windows'")
	}
	return q
}

func (r *pgADUserRepository) List(ctx context.Context, f dto.ADUserFilter) ([]domain.ADUser, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.ADUser{})
	if f.Source != "" {
		q = q.Where("source = ?", f.Source)
	}
	if f.Status != "" {
		// Status is a richer, derived filter; when present it supersedes the raw
		// `active` flag so the two can't contradict each other.
		q = applyStatus(q, f.Status)
	} else if f.Active != nil {
		q = q.Where("active = ?", *f.Active)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("sam_account_name ILIKE ? OR sid ILIKE ? OR username ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "sam_account_name ASC"
	if f.Sort == "recent" {
		order = "last_seen DESC NULLS LAST"
	}
	var items []domain.ADUser
	if err := q.Order(order).Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgADUserRepository) Stats(ctx context.Context) (*dto.ADUserStats, error) {
	base := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&domain.ADUser{})
	}

	s := dto.ADUserStats{ByDomain: []dto.DomainCount{}}
	if err := base().Count(&s.Total).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "active").Count(&s.Active).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "disabled").Count(&s.Disabled).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "deleted").Count(&s.Deleted).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "stale").Count(&s.Stale).Error; err != nil {
		return nil, err
	}
	if err := applyStatus(base(), "service").Count(&s.Service).Error; err != nil {
		return nil, err
	}
	if err := base().Where("last_seen IS NOT NULL AND last_seen > NOW() - INTERVAL '24 hours'").Count(&s.Seen24h).Error; err != nil {
		return nil, err
	}

	var srcRows []sourceRow
	if err := base().Select("source, COUNT(*) AS count").Group("source").Scan(&srcRows).Error; err != nil {
		return nil, err
	}
	for _, row := range srcRows {
		switch row.Source {
		case "windows":
			s.BySource.Windows = row.Count
		case "linux":
			s.BySource.Linux = row.Count
		}
	}

	if err := base().Select("domain, COUNT(*) AS count").Group("domain").Order("count DESC").Limit(8).Scan(&s.ByDomain).Error; err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *pgADUserRepository) Each(ctx context.Context, source string, fn func(domain.ADUser) error) error {
	q := r.db.WithContext(ctx).Model(&domain.ADUser{})
	if source != "" {
		q = q.Where("source = ?", source)
	}

	var batch []domain.ADUser
	return q.Order("id").FindInBatches(&batch, syncBatchSize, func(*gorm.DB, int) error {
		for i := range batch {
			if err := fn(batch[i]); err != nil {
				return err
			}
		}
		return nil
	}).Error
}
