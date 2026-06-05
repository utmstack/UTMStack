package usecase

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

//go:embed templates/network_scan_report.html
var networkScanReportTpl string

type reportUsecase struct {
	repo connectors.NetworkScanRepository
	tpl  *template.Template
}

func NewReportUsecase(repo connectors.NetworkScanRepository) connectors.ReportUsecase {
	tpl := template.Must(template.New("network_scan_report").Parse(networkScanReportTpl))
	return &reportUsecase{repo: repo, tpl: tpl}
}

func (u *reportUsecase) GenerateNetworkScanReport(
	ctx context.Context,
	f domain.NetworkScanFilter,
) ([]byte, error) {
	items, _, err := u.repo.SearchByFilters(ctx, f, domain.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, err
	}

	type row struct {
		Name             string
		IP               string
		MAC              string
		OS               string
		Alive            bool
		Status           string
		Severity         string
		Type             string
		DiscoveredAt     string
		ModifiedAt       string
		Addresses        string
		Aliases          string
		Notes            string
		Ports            []domain.UtmPorts
	}
	rows := make([]row, 0, len(items))
	for _, e := range items {
		r := row{
			Name:      e.AssetName,
			IP:        e.AssetIP,
			MAC:       e.AssetMAC,
			OS:        e.AssetOS,
			Status:    e.AssetStatus,
			Addresses: e.AssetAddresses,
			Aliases:   e.AssetAliases,
			Notes:     e.AssetNotes,
			Ports:     e.Ports,
		}
		if e.AssetAlive != nil {
			r.Alive = *e.AssetAlive
		}
		if e.AssetSeverityMetric != nil {
			r.Severity = formatFloat(*e.AssetSeverityMetric)
		}
		if e.AssetType != nil {
			r.Type = e.AssetType.TypeName
		}
		if e.DiscoveredAt != nil {
			r.DiscoveredAt = e.DiscoveredAt.Format("2006-01-02 15:04:05")
		}
		if e.ModifiedAt != nil {
			r.ModifiedAt = e.ModifiedAt.Format("2006-01-02 15:04:05")
		}
		rows = append(rows, r)
	}

	var buf bytes.Buffer
	if err := u.tpl.Execute(&buf, map[string]any{"Assets": rows}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
