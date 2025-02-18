package geo

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/levigross/grequests"
)

var mu = &sync.Mutex{}

var myExternalIP net.IP
var myExternalIPOnce sync.Once

type asnBlock struct {
	network *net.IPNet
	asn     int
	aso     string
}

var asnBlocks []asnBlock

type cityBlock struct {
	network             *net.IPNet
	geonameID           int
	isAnonymousProxy    bool
	isSatelliteProvider bool
	latitude            float64
	longitude           float64
	accuracyRadius      int
}

var cityBlocks []cityBlock

type cityLocation struct {
	geonameID         int
	localeCode        string
	continentCode     string
	continentName     string
	countryISOCode    string
	countryName       string
	cityName          string
	metroCode         string
	timeZone          string
	isInEuropeanUnion bool
}

var cityLocations []cityLocation

func IsLocal(a net.IP) bool {
	_, r127, _ := net.ParseCIDR("127.0.0.0/8")
	_, r10, _ := net.ParseCIDR("10.0.0.0/8")
	_, r172, _ := net.ParseCIDR("172.16.0.0/12")
	_, r192, _ := net.ParseCIDR("192.168.0.0/16")
	_, r169, _ := net.ParseCIDR("169.254.0.0/16")
	_, r224, _ := net.ParseCIDR("224.0.0.0/24")

	if r192.Contains(a) || r172.Contains(a) || r127.Contains(a) || r10.Contains(a) || r169.Contains(a) || r224.Contains(a) {
		return true
	}
	return false
}

func GetExternal() net.IP {
	if resp, err := grequests.Get("http://myexternalip.com/raw", nil); err == nil {
		ip := net.ParseIP(strings.TrimSpace(resp.String()))
		return ip
	}
	return nil
}

func GetExternalOnce() net.IP {
	myExternalIPOnce.Do(func() {
		myExternalIP = GetExternal()
	})
	return myExternalIP
}

func getCity(a string) cityBlock {
	ip := net.ParseIP(a)

	if IsLocal(ip) {
		ip = GetExternalOnce()
	}

	var city cityBlock
	for _, e := range cityBlocks {
		if e.network.Contains(ip) {
			city = e
		}
	}
	return city
}

func getASN(a string) asnBlock {
	ip := net.ParseIP(a)

	if IsLocal(ip) {
		ip = GetExternalOnce()
	}

	var asn asnBlock
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

func Geolocate(ip string) map[string]string {
	mu.Lock()
	asn := getASN(ip)
	city := getCity(ip)
	location := getLocation(city.geonameID)
	mu.Unlock()
	return map[string]string{
		"country":             location.countryName,
		"countryCode":         location.countryISOCode,
		"city":                location.cityName,
		"latitude":            fmt.Sprintf("%f", city.latitude),
		"longitude":           fmt.Sprintf("%f", city.longitude),
		"accuracyRadius":      fmt.Sprintf("%d", city.accuracyRadius),
		"asn":                 fmt.Sprintf("%d", asn.asn),
		"aso":                 asn.aso,
		"isSatelliteProvider": fmt.Sprintf("%v", city.isSatelliteProvider),
		"isAnonymousProxy":    fmt.Sprintf("%v", city.isAnonymousProxy),
	}
}
