package main

import (
	"net"
	"strings"
	"sync"

	go_sdk "github.com/threatwinds/go-sdk"
)

var mu = &sync.RWMutex{}

type asnBlock struct {
	network *net.IPNet
	asn     int64
	aso     string
}

var asnBlocks = make(map[string][]*asnBlock)

type cityBlock struct {
	network        *net.IPNet
	geonameID      int64
	latitude       float64
	longitude      float64
	accuracyRadius int32
}

var cityBlocks =  make(map[string][]*cityBlock)

type cityLocation struct {
	geonameID      int64
	countryISOCode string
	countryName    string
	cityName       string
}

var cityLocations = make(map[int64]*cityLocation)

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

func getCity(a string) *cityBlock {
	ip := net.ParseIP(a)

	var city = new(cityBlock)

	start := getStart(ip.String())

	for _, e := range cityBlocks[start] {
		if e.network.Contains(ip) {
			city = e
		}
	}
	return city
}

func getASN(a string) *asnBlock {
	ip := net.ParseIP(a)

	var asn = new(asnBlock)

	start := getStart(ip.String())

	for _, e := range asnBlocks[start] {
		if e.network.Contains(ip) {
			asn = e
		}
	}
	return asn
}

func getLocation(geonameID int64) *cityLocation {
	location, ok := cityLocations[geonameID]
	if !ok {
		return nil
	}
	
	return location
}

func geolocate(ip string) *go_sdk.Geolocation {
	mu.RLock()
	defer mu.RUnlock()

	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		go_sdk.Logger().LogF(100, "source field is not a valid IP")
		return nil
	}

	if IsLocal(parsedIp) {
		go_sdk.Logger().LogF(100, "cannot geolocate local IP")
		return nil
	}

	var geo = new(go_sdk.Geolocation)

	asn := getASN(ip)
	city := getCity(ip)

	if asn != nil {
		geo.Asn = asn.asn
		geo.Aso = asn.aso
	}

	if city != nil {
		location := getLocation(city.geonameID)
		geo.City = location.cityName
		geo.Country = location.countryName
		geo.CountryCode = location.countryISOCode
		geo.Latitude = city.latitude
		geo.Longitude = city.longitude
	}

	return geo
}

func getStart(cidr string) string {
	if strings.Contains(cidr, ":") {
		parts := strings.Split(cidr, ":")
		return strings.Join([]string{parts[0], parts[1]}, "-")
	}

	parts := strings.Split(cidr, ".")
	return strings.Join([]string{parts[0], parts[1]}, "-")
}
