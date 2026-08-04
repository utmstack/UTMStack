package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/store"
	ch "github.com/threatwinds/go-sdk/store/clickhouse"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/models"
)

const datasetAlerts = store.Dataset("alerts")

// AlertClient reads alerts an incident refers to. Read-only: this plugin
// enriches incidents from alerts and never writes one.
type AlertClient struct {
	store *ch.Driver
}

func NewAlertClient(cfg *config.TWConfig) (*AlertClient, error) {
	d, err := ch.New(ch.Config{
		Addr:     []string{cfg.ClickHouseHost + ":" + cfg.ClickHousePort},
		Database: cfg.ClickHouseDatabase,
		Username: cfg.ClickHouseUser,
		Password: cfg.ClickHousePassword,
		Tables: map[store.Dataset]string{
			datasetAlerts: cfg.AlertsTable,
		},
		TenantColumn:   "tenantId",
		TimeColumn:     "@timestamp",
		DataTypeColumn: "dataType",
	})
	if err != nil {
		return nil, catcher.Error("failed to create the alert store", err, nil)
	}

	catcher.Info("alert store connected successfully", nil)

	return &AlertClient{store: d}, nil
}

func (c *AlertClient) Close() error { return c.store.Close() }

// GetAlertByID looks an alert up across every tenant, which the scope has to
// name explicitly. An incident carries no tenant of its own, so there is none
// to scope by; the identifier is a UUID, so the lookup stays unambiguous even
// though it is unscoped.
func (c *AlertClient) GetAlertByID(ctx context.Context, alertID string) (*models.Alert, error) {
	scope := store.Scope{Tenant: store.AllTenants, Dataset: datasetAlerts}

	raw, err := c.store.FetchByID(ctx, scope, alertID)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, catcher.Error("alert not found", nil, map[string]any{"alert_id": alertID})
		}
		return nil, catcher.Error("alert lookup failed", err, map[string]any{"alert_id": alertID})
	}

	var alert models.Alert
	if err := json.Unmarshal(raw, &alert); err != nil {
		return nil, catcher.Error("failed to decode alert", err, map[string]any{"alert_id": alertID})
	}
	if alert.ID == "" {
		return nil, fmt.Errorf("alert %s came back without an id", alertID)
	}

	return &alert, nil
}
