package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func main() {
	go loadGeolocationData()

	err := plugins.InitParsingPlugin("com.utmstack.geolocation", parseLog)
	if err != nil {
		_ = catcher.Error("com.utmstack.geolocation", err, map[string]any{
			"process": "plugin_com.utmstack.geolocation",
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func parseLog(_ context.Context, transform *plugins.Transform) (*plugins.Draft, error) {
	source, ok := transform.Step.Dynamic.Params["source"]
	if !ok {
		return transform.Draft, catcher.Error("'source' parameter required", nil, map[string]any{"process": "plugin_com.utmstack.geolocation"})
	}

	destination, ok := transform.Step.Dynamic.Params["destination"]
	if !ok {
		return transform.Draft, catcher.Error("'destination' parameter required", nil, map[string]any{"process": "plugin_com.utmstack.geolocation"})
	}

	sourceField := source.GetStringValue()
	utils.SanitizeField(&sourceField)

	err := utils.ValidateReservedField(sourceField, false)
	if err != nil {
		return transform.Draft, catcher.Error("cannot parse log", err, map[string]any{
			"process": "plugin_com.utmstack.geolocation",
			"field":   sourceField,
		})
	}

	destinationField := destination.GetStringValue()
	utils.SanitizeField(&destinationField)

	err = utils.ValidateReservedField(destinationField, false)
	if err != nil {
		return transform.Draft, catcher.Error("cannot parse log", err, map[string]any{
			"process": "plugin_com.utmstack.geolocation",
			"field":   destinationField,
		})
	}

	value := gjson.Get(transform.Draft.Log, sourceField).String()
	if value == "" {
		return transform.Draft, nil
	}

	geo := geolocate(value)

	if geo == nil {
		return transform.Draft, nil
	}

	transform.Draft.Log, err = sjson.Set(transform.Draft.Log, destinationField, geo)
	if err != nil {
		return transform.Draft, catcher.Error("failed to set geolocation", err, map[string]any{"process": "plugin_com.utmstack.geolocation"})
	}

	return transform.Draft, nil
}
