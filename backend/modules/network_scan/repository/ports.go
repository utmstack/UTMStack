package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"gorm.io/gorm"
)

type pgPortsRepository struct {
	db *gorm.DB
}

func NewPortsRepository(db *gorm.DB) connectors.PortsRepository {
	return &pgPortsRepository{db: db}
}

func (r *pgPortsRepository) Save(ctx context.Context, p *domain.UtmPorts) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *pgPortsRepository) SaveAll(ctx context.Context, ports []domain.UtmPorts) error {
	if len(ports) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&ports).Error
}

func (r *pgPortsRepository) FindByID(ctx context.Context, id uint64) (*domain.UtmPorts, error) {
	var p domain.UtmPorts
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *pgPortsRepository) ListByCriteria(
	ctx context.Context,
	c domain.PortsCriteria,
	pg domain.Pagination,
) ([]domain.UtmPorts, int64, error) {
	page, size := normalizePage(pg)
	q := r.db.WithContext(ctx).Model(&domain.UtmPorts{})
	q = applyPortsCriteria(q, c)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.UtmPorts
	if err := q.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgPortsRepository) CountByCriteria(ctx context.Context, c domain.PortsCriteria) (int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmPorts{})
	q = applyPortsCriteria(q, c)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *pgPortsRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmPorts{}, id).Error
}

func (r *pgPortsRepository) DeleteByScanID(ctx context.Context, scanID uint64) error {
	return r.db.WithContext(ctx).Where("scan_id = ?", scanID).Delete(&domain.UtmPorts{}).Error
}

func (r *pgPortsRepository) DeleteByScanIDIn(ctx context.Context, scanIDs []uint64) error {
	if len(scanIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("scan_id IN ?", scanIDs).Delete(&domain.UtmPorts{}).Error
}

// UpdateInBatchForScan replaces all ports for a given scan ID atomically.
func (r *pgPortsRepository) UpdateInBatchForScan(
	ctx context.Context,
	scanID uint64,
	ports []domain.UtmPorts,
) ([]domain.UtmPorts, error) {
	out := ports
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("scan_id = ?", scanID).Delete(&domain.UtmPorts{}).Error; err != nil {
			return err
		}
		if len(out) == 0 {
			return nil
		}
		for i := range out {
			out[i].ScanID = scanID
		}
		return tx.Create(&out).Error
	})
	return out, err
}

func applyPortsCriteria(q *gorm.DB, c domain.PortsCriteria) *gorm.DB {
	if c.ID != nil {
		q = q.Where("id = ?", *c.ID)
	}
	if c.ScanID != nil {
		q = q.Where("scan_id = ?", *c.ScanID)
	}
	if c.Port != nil {
		q = q.Where("port = ?", *c.Port)
	}
	if s := strings.TrimSpace(c.TCP); s != "" {
		q = q.Where("tcp ILIKE ?", "%"+s+"%")
	}
	if s := strings.TrimSpace(c.UDP); s != "" {
		q = q.Where("udp ILIKE ?", "%"+s+"%")
	}
	return q
}
