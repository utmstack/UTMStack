package main

import (
	"net"
	"sync"

	go_sdk "github.com/threatwinds/go-sdk"
)

var mu = &sync.RWMutex{}

type asnBlock struct {
	network *net.IPNet
	asn     int64
	aso     string
}

var asnBlocks []asnBlock

type cityBlock struct {
	network        *net.IPNet
	geonameID      int
	latitude       float64
	longitude      float64
	accuracyRadius int32
}

var cityBlocks []cityBlock

type cityLocation struct {
	geonameID      int
	countryISOCode string
	countryName    string
	cityName       string
}

var cityLocations []cityLocation

func IsLocal(a net.IP) bool {
	_, r127, _ := net.ParseCIDR("127.0.0.0/8")
	_, r10, _ := net.ParseCIDR("10.0.0.0/8")
	_, r172, _ := net.ParseCIDR("172.16.0.0/12")
	_, r192, _ := net.ParseCIDR("192.168.0.0/16")
	_, r169, _ := net.ParseCIDR("169.254.0.0/16")
	_, r224, _ := net.ParseCIDR("224.0.0.0/24")

	if r192.Contains(a) || r172.Contains(a) ||
		r127.Contains(a) || r10.Contains(a) ||
		r169.Contains(a) || r224.Contains(a) {
		return true
	}

	return false
}

func getCity(a string) cityBlock {
	ip := net.ParseIP(a)

	var city cityBlock

	if IsLocal(ip) {
		return city
	}

	for _, e := range cityBlocks {
		if e.network.Contains(ip) {
			city = e
		}
	}
	return city
}

func getASN(a string) asnBlock {
	ip := net.ParseIP(a)

	var asn asnBlock

	if IsLocal(ip) {
		return asn
	}

	for _, e := range asnBlocks {
		if e.network.Contains(ip) {
			asn = e
		}
	}
	return asn
}

func getLocation(geonameID int) cityLocation {
	var location cityLocation
	for _, e := range cityLocations {
		if geonameID == e.geonameID {
			location = e
		}
	}
	return location
}

func geolocate(ip string) go_sdk.Geolocation {
	mu.RLock()
	defer mu.RUnlock()
	
	asn := getASN(ip)
	city := getCity(ip)
	location := getLocation(city.geonameID)

	return go_sdk.Geolocation{
		Country:     location.countryName,
		CountryCode: location.countryISOCode,
		City:        location.cityName,
		Latitude:    city.latitude,
		Longitude:   city.longitude,
		Accuracy:    city.accuracyRadius,
		Asn:         asn.asn,
		Aso:         asn.aso,
	}
}
