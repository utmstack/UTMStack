package main

import (
	"net"
	"os"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

func update() {
	first := true
	for {
		workdir, err := utils.MkdirJoin(plugins.WorkDir, "geolocation")
		if err != nil {
			_ = catcher.Error("could not create geolocation directory", err, nil)
			continue
		}
		var files = map[string]string{
			workdir.FileJoin("asn-blocks-v4.csv"): "https://cdn.utmstack.com/geoip/asn-blocks-v4.csv",
			workdir.FileJoin("asn-blocks-v6.csv"): "https://cdn.utmstack.com/geoip/asn-blocks-v6.csv",
			workdir.FileJoin("blocks-v4.csv"):     "https://cdn.utmstack.com/geoip/blocks-v4.csv",
			workdir.FileJoin("blocks-v6.csv"):     "https://cdn.utmstack.com/geoip/blocks-v6.csv",
			workdir.FileJoin("locations-en.csv"):  "https://cdn.utmstack.com/geoip/locations-en.csv",
		}

		mode := os.Getenv("MODE")
		if mode == "manager" {
			if _, err := os.Stat(workdir.FileJoin("locations-en.csv")); os.IsNotExist(err) || !first {
				for file, url := range files {
					if err := utils.Download(url, file); err != nil {
						_ = catcher.Error("could not download geolocation file", err, nil)
						continue
					}
				}
			}
		} else {
			time.Sleep(5 * time.Minute)
		}

		mu.Lock()

		asnBlocks = make(map[string][]*asnBlock)
		cityBlocks = make(map[string][]*cityBlock)
		cityLocations = make(map[int64]*cityLocation)

		for file := range files {
			csv, err := utils.ReadCSV(file)
			if err != nil {
				_ = catcher.Error("could not read geolocation file", err, nil)
				continue
			}

			switch file {
			case workdir.FileJoin("asn-blocks-v4.csv"):
				populateASNBlocks(csv)
			case workdir.FileJoin("asn-blocks-v6.csv"):
				populateASNBlocks(csv)
			case workdir.FileJoin("blocks-v4.csv"):
				populateCityBlocks(csv)
			case workdir.FileJoin("blocks-v6.csv"):
				populateCityBlocks(csv)
			case workdir.FileJoin("locations-en.csv"):
				populateCityLocations(csv)
			}
		}

		mu.Unlock()

		if first {
			first = false
		}

		time.Sleep(168 * time.Hour)
	}
}

func populateASNBlocks(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			_ = catcher.Error("could not parse CIDR", err, map[string]any{
				"cidr": line[0],
			})
			continue
		}

		asn, err := strconv.Atoi(func() string {
			if line := line[1]; line != "" {
				return line
			}
			return "0"
		}())
		if err != nil {
			_ = catcher.Error("could not parse ASN", err, map[string]any{
				"asn": line[1],
			})
			continue
		}

		t := &asnBlock{
			network: n,
			asn:     int64(asn),
			aso: func() string {
				if line := line[2]; line != "" {
					return line
				}
				return "-"
			}(),
		}

		start := getStart(n.IP.String())

		asnBlocks[start] = append(asnBlocks[start], t)
	}
}

func populateCityBlocks(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			_ = catcher.Error("could not parse CIDR", err, map[string]any{
				"cidr": line[0],
			})
			continue
		}

		geonameID, err := strconv.ParseInt(func() string {
			if line := line[1]; line != "" {
				return line
			}
			return "0"
		}(), 10, 64)
		if err != nil {
			_ = catcher.Error("could not parse geonameID", err, map[string]any{
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

		cityBlocks[start] = append(cityBlocks[start], t)
	}
}

func populateCityLocations(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}

		geonameID, err := strconv.ParseInt(line[0], 10, 64)
		if err != nil {
			_ = catcher.Error("could not parse geonameID", err, map[string]any{
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

		cityLocations[geonameID] = t
	}
}
