package main

import (
	"net"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/plugins"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
)

// loadGeolocationData loads the geolocation files and populates the maps
func loadGeolocationData() {
	// Get the geolocation directory from environment variable or use default
	workdir := plugins.WorkDir
	geoDir, err := utils.MkdirJoin(workdir, "geolocation")
	if err != nil {
		_ = catcher.Error("could not create geolocation directory", err, map[string]any{"process": "plugin_com.utmstack.geolocation"})
		return
	}

	// Define the geolocation files to be loaded
	var files = []string{
		"asn-blocks-v4.csv",
		"asn-blocks-v6.csv",
		"blocks-v4.csv",
		"blocks-v6.csv",
		"locations-en.csv",
	}

	// Use deferring to ensure the mutex is always unlocked, even if there's an error
	func() {
		mu.Lock()
		defer mu.Unlock()

		// Create new maps to avoid partial updates if there's an error
		newAsnBlocks := make(map[string][]*asnBlock)
		newCityBlocks := make(map[string][]*cityBlock)
		newCityLocations := make(map[uint64]*cityLocation)

		// Process each file with retry logic
		maxRetries := 3
		for _, filename := range files {
			filePath := geoDir.FileJoin(filename)
			var csv [][]string
			var err error

			// Retry logic for reading the file
			for retry := 0; retry < maxRetries; retry++ {
				csv, err = utils.ReadCSV(filePath)
				if err == nil {
					break
				}

				_ = catcher.Error("could not read geolocation file, retrying", err, map[string]any{
					"process":    "plugin_com.utmstack.geolocation",
					"file":       filePath,
					"retry":      retry + 1,
					"maxRetries": maxRetries,
				})

				if retry < maxRetries-1 {
					// Exponential backoff
					time.Sleep(time.Duration(1<<uint(retry)) * time.Second)
				}
			}

			if err != nil {
				_ = catcher.Error("all retries failed when reading geolocation file", err, map[string]any{
					"process": "plugin_com.utmstack.geolocation",
					"file":    filePath,
				})
				continue
			}

			switch filename {
			case "asn-blocks-v4.csv":
				populateASNBlocksMap(csv, newAsnBlocks)
			case "asn-blocks-v6.csv":
				populateASNBlocksMap(csv, newAsnBlocks)
			case "blocks-v4.csv":
				populateCityBlocksMap(csv, newCityBlocks)
			case "blocks-v6.csv":
				populateCityBlocksMap(csv, newCityBlocks)
			case "locations-en.csv":
				populateCityLocationsMap(csv, newCityLocations)
			}
		}

		// Only update the global maps if all files were processed successfully
		if len(newAsnBlocks) > 0 && len(newCityBlocks) > 0 && len(newCityLocations) > 0 {
			asnBlocks = newAsnBlocks
			cityBlocks = newCityBlocks
			cityLocations = newCityLocations
		}
	}()
}

// populateASNBlocksMap populates the provided ASN blocks map with data from the CSV
func populateASNBlocksMap(csv [][]string, blocks map[string][]*asnBlock) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			_ = catcher.Error("could not parse CIDR", err, map[string]any{
				"process": "plugin_com.utmstack.geolocation",
				"cidr":    line[0],
			})
			continue
		}

		asn, err := strconv.ParseUint(func() string {
			if line := line[1]; line != "" {
				return line
			}
			return "0"
		}(), 10, 64)
		if err != nil {
			_ = catcher.Error("could not parse ASN", err, map[string]any{
				"process": "plugin_com.utmstack.geolocation",
				"asn":     line[1],
			})
			continue
		}

		t := &asnBlock{
			network: n,
			asn:     asn,
			aso: func() string {
				if line := line[2]; line != "" {
					return line
				}
				return "-"
			}(),
		}

		start := getStart(n.IP.String())

		blocks[start] = append(blocks[start], t)
	}
}

// populateCityBlocksMap populates the provided city blocks map with data from the CSV
func populateCityBlocksMap(csv [][]string, blocks map[string][]*cityBlock) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			_ = catcher.Error("could not parse CIDR", err, map[string]any{
				"process": "plugin_com.utmstack.geolocation",
				"cidr":    line[0],
			})
			continue
		}

		geonameID, err := strconv.ParseUint(func() string {
			if line := line[1]; line != "" {
				return line
			}
			return "0"
		}(), 10, 64)
		if err != nil {
			_ = catcher.Error("could not parse geonameID", err, map[string]any{
				"process":   "plugin_com.utmstack.geolocation",
				"geonameID": line[1],
			})
			continue
		}

		latitude, err := strconv.ParseFloat(func() string {
			if line := line[7]; line != "" {
				return line
			}
			return "0.0"
		}(), 64)
		if err != nil {
			_ = catcher.Error("could not parse latitude", err, map[string]any{
				"process":  "plugin_com.utmstack.geolocation",
				"latitude": line[7],
			})
			continue
		}

		longitude, err := strconv.ParseFloat(func() string {
			if line := line[8]; line != "" {
				return line
			}
			return "0.0"
		}(), 64)
		if err != nil {
			_ = catcher.Error("could not parse longitude", err, map[string]any{
				"process":   "plugin_com.utmstack.geolocation",
				"longitude": line[8],
			})
			continue
		}

		accuracyRadius, err := strconv.Atoi(func() string {
			if line := line[9]; line != "" {
				return line
			}
			return "0"
		}())
		if err != nil {
			_ = catcher.Error("could not parse accuracyRadius", err, map[string]any{
				"process":        "plugin_com.utmstack.geolocation",
				"accuracyRadius": line[9],
			})
			continue
		}

		t := &cityBlock{
			network:        n,
			geonameID:      geonameID,
			latitude:       latitude,
			longitude:      longitude,
			accuracyRadius: int32(accuracyRadius),
		}

		start := getStart(n.IP.String())

		blocks[start] = append(blocks[start], t)
	}
}

// populateCityLocationsMap populates the provided city locations map with data from the CSV
func populateCityLocationsMap(csv [][]string, locations map[uint64]*cityLocation) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		geonameID, err := strconv.ParseUint(line[0], 10, 64)
		if err != nil {
			_ = catcher.Error("could not parse geonameID", err, map[string]any{
				"process":   "plugin_com.utmstack.geolocation",
				"geonameID": line[0],
			})
			continue
		}

		t := &cityLocation{
			geonameID:      geonameID,
			countryISOCode: line[4],
			countryName:    line[5],
			cityName:       line[10],
		}

		locations[geonameID] = t
	}
}
