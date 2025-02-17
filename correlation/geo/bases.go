package geo

import (
	"github.com/utmstack/UTMStack/correlation/utils"
	"log"
	"net"
	"path/filepath"
	"strconv"
)

func Load() {
	log.Printf("Loading GeoIP databases")

	var files = []string{
		"asn-blocks-v4.csv",
		"asn-blocks-v6.csv",
		"blocks-v4.csv",
		"blocks-v6.csv",
		"locations-en.csv",
	}

	for _, file := range files {
		csv := utils.ReadCSV(filepath.Join("/app", file))
		switch file {
		case "asn-blocks-v4.csv":
			asnBlocks = nil
			populateASNBlocks(csv)
		case "asn-blocks-v6.csv":
			populateASNBlocks(csv)
		case "blocks-v4.csv":
			cityBlocks = nil
			populateCityBlocks(csv)
		case "blocks-v6.csv":
			populateCityBlocks(csv)
		case "locations-en.csv":
			cityLocations = nil
			populateCityLocations(csv)
		}
	}

	log.Printf("asnBlocks rows: %v", len(asnBlocks))
	log.Printf("cityBlocks rows: %v", len(cityBlocks))
	log.Printf("cityLocations rows: %v", len(cityLocations))
	log.Printf("GeoIP databases loaded")
}

func populateASNBlocks(csv [][]string) {
	for key, line := range csv {
		if key == 0 {
			continue
		}
		_, n, err := net.ParseCIDR(line[0])
		if err != nil {
			log.Printf("Could not get CIDR in populateASNBlocks: %v", err)
			continue
		}

		asn, err := strconv.Atoi(line[1])
		if err != nil {
			log.Printf("Could not get ASN in populateASNBlocks: %v", err)
			continue
		}

		t := asnBlock{
			network: n,
			asn:     asn,
			aso:     line[2],
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
			log.Printf("Could not parse CIDR in populateCityBlocks: %v", err)
			continue
		}

		if line[1] == "" {
			continue
		}

		geonameID, err := strconv.Atoi(line[1])
		if err != nil {
			log.Printf("Could not parse geonameID in populateCityBlocks: %v", err)
			continue
		}

		isAnonymousProxy, err := strconv.Atoi(line[4])
		if err != nil {
			log.Printf("Could not parse isAnonymousProxy in populateCityBlocks: %v", err)
			continue
		}

		var iap bool
		if isAnonymousProxy == 1 {
			iap = true
		}

		isSatelliteProvider, err := strconv.Atoi(line[5])
		if err != nil {
			log.Printf("Could not parse isSatelliteProvider in populateCityBlocks: %v", err)
			continue
		}

		var isp bool
		if isSatelliteProvider == 1 {
			isp = true
		}

		latitude, err := strconv.ParseFloat(line[7], 64)
		if err != nil {
			log.Printf("Could not parse latitude in populateCityBlocks: %v", err)
			continue
		}

		longitude, err := strconv.ParseFloat(line[8], 64)
		if err != nil {
			log.Printf("Could not parse longitude in populateCityBlocks: %v", err)
			continue
		}

		accuracyRadius, err := strconv.Atoi(line[9])
		if err != nil {
			log.Printf("Could not parse accuracyRadius in populateCityBlocks: %v", err)
			continue
		}

		t := cityBlock{
			network:             n,
			geonameID:           geonameID,
			isAnonymousProxy:    iap,
			isSatelliteProvider: isp,
			latitude:            latitude,
			longitude:           longitude,
			accuracyRadius:      accuracyRadius,
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
			log.Printf("Could not parse geonameID in populateCityLocations: %v", err)
			continue
		}

		isInEuropeanUnion, err := strconv.Atoi(line[13])
		if err != nil {
			log.Printf("Could not parse isInEuropeanUnion in populateCityLocations: %v", err)
			continue
		}

		var iieu bool
		if isInEuropeanUnion == 1 {
			iieu = true
		}

		t := cityLocation{
			geonameID:         geonameID,
			localeCode:        line[1],
			continentCode:     line[2],
			continentName:     line[3],
			countryISOCode:    line[4],
			countryName:       line[5],
			cityName:          line[10],
			metroCode:         line[11],
			timeZone:          line[12],
			isInEuropeanUnion: iieu,
		}

		cityLocations = append(cityLocations, t)
	}
}
