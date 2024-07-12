package main

import (
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
)

// Goroutine for update geolocalization databases
func update() {
	first := true
	for {
		helpers.Logger().Info("updating GeoIP databases")
		workdir := path.Join(helpers.GetCfg().Env.Workdir, "geolocation")
		var files = map[string]string{
			filepath.Join(workdir, "asn-blocks-v4.csv"): "https://cdn.utmstack.com/geoip/asn-blocks-v4.csv",
			filepath.Join(workdir, "asn-blocks-v6.csv"): "https://cdn.utmstack.com/geoip/asn-blocks-v6.csv",
			filepath.Join(workdir, "blocks-v4.csv"):     "https://cdn.utmstack.com/geoip/blocks-v4.csv",
			filepath.Join(workdir, "blocks-v6.csv"):     "https://cdn.utmstack.com/geoip/blocks-v6.csv",
			filepath.Join(workdir, "locations-en.csv"):  "https://cdn.utmstack.com/geoip/locations-en.csv",
		}

		if _, err := os.Stat(workdir); os.IsNotExist(err) {
			os.MkdirAll(workdir, os.ModeDir)
		}

		if _, err := os.Stat(filepath.Join(workdir, "locations-en.csv")); os.IsNotExist(err) || !first {
			for file, url := range files {
				if err := helpers.Download(url, file); err != nil {
					continue
				}
			}
		}

		mu.Lock()

		asnBlocks = nil
		cityBlocks = nil
		cityLocations = nil

		for file := range files {
			csv, err := helpers.ReadCSV(file)
			if err != nil {
				continue
			}

			switch file {
			case filepath.Join(workdir, "asn-blocks-v4.csv"):
				populateASNBlocks(csv)
			case filepath.Join(workdir, "asn-blocks-v6.csv"):
				populateASNBlocks(csv)
			case filepath.Join(workdir, "blocks-v4.csv"):
				populateCityBlocks(csv)
			case filepath.Join(workdir, "blocks-v6.csv"):
				populateCityBlocks(csv)
			case filepath.Join(workdir, "locations-en.csv"):
				populateCityLocations(csv)
			}
		}

		mu.Unlock()

		if first {
			first = false
		}

		time.Sleep(48 * time.Hour)
	}
}

func populateASNBlocks(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			helpers.Logger().ErrorF("could not parse CIDR in populateASNBlocks: %s", err.Error())
			continue
		}

		asn, err := strconv.Atoi(func() string {
			if line := line[1]; line != "" {
				return line
			}
			return "0"
		}())
		if err != nil {
			helpers.Logger().ErrorF("could not parse ASN in populateASNBlocks: %s", err.Error())
			continue
		}

		t := asnBlock{
			network: n,
			asn:     int64(asn),
			aso: func() string {
				if line := line[2]; line != "" {
					return line
				}
				return "-"
			}(),
		}

		asnBlocks = append(asnBlocks, t)
	}
}

func populateCityBlocks(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			helpers.Logger().ErrorF("could not parse CIDR in populateCityBlocks: %s", err.Error())
			continue
		}

		geonameID, err := strconv.Atoi(func() string {
			if line := line[1]; line != "" {
				return line
			}
			return "0"
		}())
		if err != nil {
			helpers.Logger().ErrorF("could not parse geonameID in populateCityBlocks: %s", err.Error())
			continue
		}

		latitude, err := strconv.ParseFloat(func() string {
			if line := line[7]; line != "" {
				return line
			}
			return "0.0"
		}(), 64)
		if err != nil {
			helpers.Logger().ErrorF("could not parse latitude in populateCityBlocks: %s", err.Error())
			continue
		}

		longitude, err := strconv.ParseFloat(func() string {
			if line := line[8]; line != "" {
				return line
			}
			return "0.0"
		}(), 64)
		if err != nil {
			helpers.Logger().ErrorF("could not parse longitude in populateCityBlocks: %s", err.Error())
			continue
		}

		accuracyRadius, err := strconv.Atoi(func() string {
			if line := line[9]; line != "" {
				return line
			}
			return "0"
		}())
		if err != nil {
			helpers.Logger().ErrorF("could not parse accuracyRadius in populateCityBlocks: %s", err.Error())
			continue
		}

		t := cityBlock{
			network:        n,
			geonameID:      geonameID,
			latitude:       latitude,
			longitude:      longitude,
			accuracyRadius: int32(accuracyRadius),
		}

		cityBlocks = append(cityBlocks, t)
	}
}

func populateCityLocations(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		geonameID, err := strconv.Atoi(line[0])
		if err != nil {
			helpers.Logger().ErrorF("could not parse geonameID in populateCityLocations: %s", err.Error())
			continue
		}

		t := cityLocation{
			geonameID:      geonameID,
			countryISOCode: line[4],
			countryName:    line[5],
			cityName:       line[10],
		}

		cityLocations = append(cityLocations, t)
	}
}