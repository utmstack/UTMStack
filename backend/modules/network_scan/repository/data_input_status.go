package repository

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dataInputStatusRow struct {
	ID        string  `gorm:"column:id;primaryKey"`
	Source    string  `gorm:"column:source"`
	DataType  string  `gorm:"column:data_type"`
	Timestamp int64   `gorm:"column:timestamp"`
	Median    *int64  `gorm:"column:median"`
	Alias     *string `gorm:"column:alias"`
}

func (dataInputStatusRow) TableName() string { return "utm_data_input_status" }

type checkpointRow struct {
	ID                     int64     `gorm:"column:id;primaryKey;autoIncrement"`
	LastProcessedTimestamp time.Time `gorm:"column:last_processed_timestamp"`
}

func (checkpointRow) TableName() string { return "utm_data_input_status_checkpoint" }

type pgDataInputStatusGateway struct {
	db *gorm.DB
}

func NewDataInputStatusGateway(db *gorm.DB) connectors.DataInputStatusGateway {
	return &pgDataInputStatusGateway{db: db}
}

func (g *pgDataInputStatusGateway) DeleteByDataSource(ctx context.Context, source string) error {
	return g.db.WithContext(ctx).Where("source = ?", source).Delete(&dataInputStatusRow{}).Error
}

func (g *pgDataInputStatusGateway) FindAll(ctx context.Context) ([]connectors.DataInputStatus, error) {
	var rows []dataInputStatusRow
	if err := g.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]connectors.DataInputStatus, 0, len(rows))
	for _, r := range rows {
		out = append(out, connectors.DataInputStatus{
			ID:        r.ID,
			Source:    r.Source,
			DataType:  r.DataType,
			Timestamp: r.Timestamp,
			Median:    r.Median,
			Alias:     r.Alias,
		})
	}
	return out, nil
}

func (g *pgDataInputStatusGateway) Upsert(ctx context.Context, s connectors.DataInputStatus) error {
	row := dataInputStatusRow{
		ID:        s.ID,
		Source:    s.Source,
		DataType:  s.DataType,
		Timestamp: s.Timestamp,
		Median:    s.Median,
		Alias:     s.Alias,
	}
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"source", "data_type", "timestamp", "median", "alias"}),
	}).Create(&row).Error
}

func (g *pgDataInputStatusGateway) GetCheckpoint(ctx context.Context) (*connectors.Checkpoint, error) {
	var row checkpointRow
	err := g.db.WithContext(ctx).First(&row, 1).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &connectors.Checkpoint{
		ID:                     row.ID,
		LastProcessedTimestamp: row.LastProcessedTimestamp,
	}, nil
}

func (g *pgDataInputStatusGateway) SaveCheckpoint(ctx context.Context, cp connectors.Checkpoint) error {
	row := checkpointRow{
		ID:                     cp.ID,
		LastProcessedTimestamp: cp.LastProcessedTimestamp,
	}
	if row.ID == 0 {
		row.ID = 1
	}
	return g.db.WithContext(ctx).Save(&row).Error
}
