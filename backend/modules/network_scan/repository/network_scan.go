package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"gorm.io/gorm"
)

type pgNetworkScanRepository struct {
	db *gorm.DB
}

func NewNetworkScanRepository(db *gorm.DB) connectors.NetworkScanRepository {
	return &pgNetworkScanRepository{db: db}
}

func (r *pgNetworkScanRepository) FindByID(ctx context.Context, id uint64) (*domain.UtmNetworkScan, error) {
	var e domain.UtmNetworkScan
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgNetworkScanRepository) FindByIDWithDetails(ctx context.Context, id uint64) (*domain.UtmNetworkScan, error) {
	var e domain.UtmNetworkScan
	err := r.db.WithContext(ctx).
		Preload("AssetType").
		Preload("AssetGroup").
		Preload("Ports").
		First(&e, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgNetworkScanRepository) FindByAssetName(ctx context.Context, name string) (*domain.UtmNetworkScan, error) {
	var e domain.UtmNetworkScan
	err := r.db.WithContext(ctx).Where("LOWER(asset_name) = LOWER(?)", name).First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgNetworkScanRepository) FindByNameOrIP(ctx context.Context, q string) (*domain.UtmNetworkScan, error) {
	var e domain.UtmNetworkScan
	err := r.db.WithContext(ctx).
		Where("LOWER(asset_name) = LOWER(?) OR asset_ip = ?", q, q).
		First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgNetworkScanRepository) FindByAssetIPsOrNames(
	ctx context.Context,
	ips, names []string,
) ([]domain.UtmNetworkScan, error) {
	if len(ips) == 0 && len(names) == 0 {
		return nil, nil
	}
	q := r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{})
	switch {
	case len(ips) > 0 && len(names) > 0:
		q = q.Where("asset_ip IN ? OR asset_name IN ?", ips, names)
	case len(ips) > 0:
		q = q.Where("asset_ip IN ?", ips)
	case len(names) > 0:
		q = q.Where("asset_name IN ?", names)
	}
	var items []domain.UtmNetworkScan
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgNetworkScanRepository) ListByCriteria(
	ctx context.Context,
	c domain.NetworkScanCriteria,
	p domain.Pagination,
) ([]domain.UtmNetworkScan, int64, error) {
	page, size := normalizePage(p)
	q := r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{})
	q = applyCriteria(q, c)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.UtmNetworkScan
	if err := q.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgNetworkScanRepository) CountByCriteria(ctx context.Context, c domain.NetworkScanCriteria) (int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{})
	q = applyCriteria(q, c)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func applyCriteria(q *gorm.DB, c domain.NetworkScanCriteria) *gorm.DB {
	if c.ID != nil {
		q = q.Where("id = ?", *c.ID)
	}
	if s := strings.TrimSpace(c.IP); s != "" {
		q = q.Where("asset_ip ILIKE ?", "%"+s+"%")
	}
	if s := strings.TrimSpace(c.MAC); s != "" {
		q = q.Where("asset_mac ILIKE ?", "%"+s+"%")
	}
	if s := strings.TrimSpace(c.OS); s != "" {
		q = q.Where("asset_os ILIKE ?", "%"+s+"%")
	}
	if s := strings.TrimSpace(c.Name); s != "" {
		q = q.Where("asset_name ILIKE ?", "%"+s+"%")
	}
	if c.Alive != nil {
		q = q.Where("asset_alive = ?", *c.Alive)
	}
	if s := strings.TrimSpace(c.Status); s != "" {
		q = q.Where("asset_status = ?", s)
	}
	if c.DiscoveredAt != nil {
		q = q.Where("discovered_at >= ?", *c.DiscoveredAt)
	}
	if c.ModifiedAt != nil {
		q = q.Where("modified_at >= ?", *c.ModifiedAt)
	}
	return q
}

// SearchByFilters implements the rich /search-by-filters endpoint. Mirrors the Java
// UtmNetworkScanRepository.searchByFilters JPQL — assembled progressively with GORM
// instead of one giant native string.
func (r *pgNetworkScanRepository) SearchByFilters(
	ctx context.Context,
	f domain.NetworkScanFilter,
	p domain.Pagination,
) ([]domain.UtmNetworkScan, int64, error) {
	page, size := normalizePage(p)
	q := r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{}).
		Preload("AssetType").
		Preload("AssetGroup")

	// Asset IP / MAC / Name combined free-text search.
	if s := strings.TrimSpace(f.AssetIPMacName); s != "" {
		like := "%" + s + "%"
		q = q.Where("(asset_ip ILIKE ? OR asset_mac ILIKE ? OR asset_name ILIKE ?)", like, like, like)
	}
	if len(f.OS) > 0 {
		q = q.Where("asset_os IN ?", f.OS)
	}
	if len(f.OSPlatform) > 0 {
		q = q.Where("asset_os_platform IN ?", f.OSPlatform)
	}
	if len(f.Alive) > 0 {
		q = q.Where("asset_alive IN ?", f.Alive)
	}
	if len(f.Status) > 0 {
		q = q.Where("asset_status IN ?", f.Status)
	}
	if len(f.Type) > 0 {
		q = q.Where("asset_type_id IN ?", f.Type)
	}
	if len(f.Alias) > 0 {
		q = q.Where("asset_alias IN ?", f.Alias)
	}
	if len(f.Probe) > 0 {
		q = q.Where("server_name IN ?", f.Probe)
	}
	if len(f.Groups) > 0 {
		q = q.Where("group_id IN ?", f.Groups)
	}
	if f.DiscoveredInitDate != nil {
		q = q.Where("discovered_at >= ?", *f.DiscoveredInitDate)
	}
	if f.DiscoveredEndDate != nil {
		q = q.Where("discovered_at <= ?", *f.DiscoveredEndDate)
	}
	if s := strings.TrimSpace(f.RegisteredMode); s != "" {
		q = q.Where("registered_mode = ?", s)
	}
	if len(f.Agent) > 0 {
		q = q.Where("is_agent IN ?", f.Agent)
	}

	// open_ports filter: EXISTS in utm_ports
	if len(f.OpenPorts) > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM utm_ports p WHERE p.scan_id = utm_network_scan.id AND p.port IN ?)", f.OpenPorts)
	}
	// dataTypes filter: EXISTS in utm_data_input_status (asset_name == source)
	if len(f.DataTypes) > 0 {
		q = q.Where(
			"EXISTS (SELECT 1 FROM utm_data_input_status s WHERE s.source = utm_network_scan.asset_name AND s.data_type IN ?)",
			f.DataTypes,
		)
	}

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UtmNetworkScan
	if err := q.Order("modified_at DESC NULLS LAST, id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgNetworkScanRepository) SearchPropertyValues(
	ctx context.Context,
	prop domain.Property,
	value string,
	forGroups bool,
	p domain.Pagination,
) ([]string, error) {
	page, size := normalizePage(p)
	from := prop.FromTable
	join := ""
	if prop.JoinTable != "" {
		join = " INNER JOIN " + prop.JoinTable
	}
	where := "1=1"
	args := []any{}
	if s := strings.TrimSpace(value); s != "" {
		where = "CAST(" + prop.FromTable + "." + prop.Name + " AS TEXT) ILIKE ?"
		args = append(args, "%"+s+"%")
	}
	groupBy := ""
	if forGroups {
		// `forGroups=true` requests aggregation-friendly distinct values across groups.
		// Java behavior: same distinct list; preserve by adding a GROUP BY clause.
		groupBy = " GROUP BY " + prop.FromTable + "." + prop.Name
	} else {
		groupBy = " GROUP BY " + prop.FromTable + "." + prop.Name
	}
	query := fmt.Sprintf(
		"SELECT DISTINCT %s.%s FROM %s%s WHERE %s%s ORDER BY 1 OFFSET ? LIMIT ?",
		prop.FromTable, prop.Name, from, join, where, groupBy,
	)
	args = append(args, (page-1)*size, size)

	var results []*string
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(results))
	for _, s := range results {
		if s != nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *pgNetworkScanRepository) CountByStatus(ctx context.Context, status domain.AssetStatus) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{}).
		Where("asset_status = ?", string(status)).
		Count(&total).Error
	return total, err
}

func (r *pgNetworkScanRepository) Save(ctx context.Context, scan *domain.UtmNetworkScan) error {
	return r.db.WithContext(ctx).Save(scan).Error
}

func (r *pgNetworkScanRepository) SaveAll(ctx context.Context, scans []domain.UtmNetworkScan) error {
	if len(scans) == 0 {
		return nil
	}
	// Save() honors PK and runs an upsert-by-key; for batch we just split insert vs update by ID.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range scans {
			if err := tx.Save(&scans[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *pgNetworkScanRepository) UpdateType(ctx context.Context, assetIDs []uint64, typeID *uint64) error {
	if len(assetIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{}).
		Where("id IN ?", assetIDs).
		Update("asset_type_id", typeID).Error
}

func (r *pgNetworkScanRepository) UpdateGroup(ctx context.Context, assetIDs []uint64, groupID *uint64) error {
	if len(assetIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{}).
		Where("id IN ?", assetIDs).
		Update("group_id", groupID).Error
}

func (r *pgNetworkScanRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmNetworkScan{}, id).Error
}

func (r *pgNetworkScanRepository) FindAllByUpdateLevelIn(
	ctx context.Context,
	levels []domain.UpdateLevel,
) ([]domain.UtmNetworkScan, error) {
	var items []domain.UtmNetworkScan
	q := r.db.WithContext(ctx).Model(&domain.UtmNetworkScan{})
	if len(levels) == 0 {
		if err := q.Where("update_level IS NULL").Find(&items).Error; err != nil {
			return nil, err
		}
		return items, nil
	}
	str := make([]string, 0, len(levels))
	for _, l := range levels {
		str = append(str, string(l))
	}
	if err := q.Where("update_level IS NULL OR update_level IN ?", str).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgNetworkScanRepository) FindAllAssetGroupMappings(ctx context.Context) ([]domain.AssetGroupMapping, error) {
	type row struct {
		AssetName string  `gorm:"column:asset_name"`
		GroupID   *uint64 `gorm:"column:id"`
		GroupName *string `gorm:"column:group_name"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT n.asset_name, g.id, g.group_name
		FROM utm_network_scan n
		LEFT JOIN utm_asset_group g ON n.group_id = g.id
		WHERE n.asset_name IS NOT NULL
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.AssetGroupMapping, 0, len(rows))
	for _, x := range rows {
		m := domain.AssetGroupMapping{AssetName: x.AssetName}
		if x.GroupID != nil {
			m.GroupID = *x.GroupID
		}
		if x.GroupName != nil {
			m.GroupName = *x.GroupName
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *pgNetworkScanRepository) FindAgentsOSPlatforms(ctx context.Context) ([]string, error) {
	var rows []*string
	if err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT asset_os_platform
		FROM utm_network_scan
		WHERE is_agent = TRUE AND asset_os_platform IS NOT NULL
		ORDER BY 1
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, s := range rows {
		if s != nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *pgNetworkScanRepository) FindAgentNamesByPlatform(ctx context.Context, platform string) ([]string, error) {
	var rows []*string
	if err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT asset_name
		FROM utm_network_scan
		WHERE is_agent = TRUE AND asset_os_platform = ?
		ORDER BY 1
	`, platform).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, s := range rows {
		if s != nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *pgNetworkScanRepository) FindAllAgents(ctx context.Context) ([]domain.UtmNetworkScan, error) {
	var items []domain.UtmNetworkScan
	if err := r.db.WithContext(ctx).Where("is_agent = TRUE").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
