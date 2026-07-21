package parselog

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/utmstack/UTMStack/plugins/enrichment/config"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/registry"
)

type parseParams struct {
	dataset      string
	source       string
	matchColumn  string
	outputColumn string
	destination  string
}

func ParseLog(_ context.Context, transform *plugins.Transform) (*plugins.Draft, error) {
	params, err := extractParams(transform.Step.Dynamic.Params)
	if err != nil {
		return transform.Draft, err
	}

	if err := validateFieldPaths(&params); err != nil {
		return transform.Draft, err
	}

	sourceValue := gjson.Get(transform.Draft.Log, params.source).String()
	if sourceValue == "" {
		return transform.Draft, nil
	}

	outputValue, ok := lookupEnrichment(params, sourceValue)
	if !ok {
		return transform.Draft, nil
	}

	return writeEnrichment(transform.Draft, params.destination, outputValue)
}

func extractParams(raw map[string]*structpb.Value) (parseParams, error) {
	dataset, err := requireString(raw, "dataset")
	if err != nil {
		return parseParams{}, err
	}
	source, err := requireString(raw, "source")
	if err != nil {
		return parseParams{}, err
	}
	matchColumn, err := requireString(raw, "match_column")
	if err != nil {
		return parseParams{}, err
	}
	outputColumn, err := requireString(raw, "output_column")
	if err != nil {
		return parseParams{}, err
	}
	destination, err := requireString(raw, "destination")
	if err != nil {
		return parseParams{}, err
	}
	return parseParams{
		dataset:      dataset,
		source:       source,
		matchColumn:  matchColumn,
		outputColumn: outputColumn,
		destination:  destination,
	}, nil
}

func requireString(raw map[string]*structpb.Value, key string) (string, error) {
	v, ok := raw[key]
	if !ok {
		return "", catcher.Error("'"+key+"' param required", nil, procMeta())
	}
	return v.GetStringValue(), nil
}

func validateFieldPaths(p *parseParams) error {
	utils.SanitizeField(&p.source)
	if err := utils.ValidateReservedField(p.source, false); err != nil {
		return catcher.Error("invalid source field", err, procMeta())
	}
	utils.SanitizeField(&p.destination)
	if err := utils.ValidateReservedField(p.destination, false); err != nil {
		return catcher.Error("invalid destination field", err, procMeta())
	}
	return nil
}

func lookupEnrichment(p parseParams, sourceValue string) (string, bool) {
	ds, ok := registry.Get(p.dataset)
	if !ok {
		if registry.MarkMissingDataset(p.dataset) {
			catcher.Warn("dataset not found in registry (fail-open)", map[string]any{
				"process":   config.ProcessName,
				"datasetId": p.dataset,
			})
		}
		return "", false
	}

	idx, err := ds.GetOrBuildIndex(p.matchColumn)
	if err != nil {
		if registry.MarkMissingColumn(p.dataset, p.matchColumn) {
			_ = catcher.Error("match_column not in dataset", err, map[string]any{
				"process":   config.ProcessName,
				"datasetId": p.dataset,
				"column":    p.matchColumn,
			})
		}
		return "", false
	}

	outputColIdx, ok := ds.ColIndex[p.outputColumn]
	if !ok {
		if registry.MarkMissingColumn(p.dataset, p.outputColumn) {
			_ = catcher.Error("output_column not in dataset", nil, map[string]any{
				"process":   config.ProcessName,
				"datasetId": p.dataset,
				"column":    p.outputColumn,
			})
		}
		return "", false
	}

	rowIndices, ok := idx[registry.NormalizeKey(sourceValue)]
	if !ok || len(rowIndices) == 0 {
		return "", false
	}

	row := ds.RawRows[rowIndices[len(rowIndices)-1]]
	if outputColIdx >= len(row) {
		return "", false
	}
	return row[outputColIdx], true
}

func writeEnrichment(draft *plugins.Draft, destination, value string) (*plugins.Draft, error) {
	updatedLog, err := sjson.Set(draft.Log, destination, value)
	if err != nil {
		return draft, catcher.Error("failed to set enriched field", err, map[string]any{
			"process":     config.ProcessName,
			"destination": destination,
		})
	}
	draft.Log = updatedLog
	return draft, nil
}

func procMeta() map[string]any {
	return map[string]any{"process": config.ProcessName}
}
