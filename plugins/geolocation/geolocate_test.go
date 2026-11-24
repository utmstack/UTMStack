package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func writeGeoFixtures(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func standardFixtures() map[string]string {
	return map[string]string{
		"locations-en.csv": "geoname_id,locale_code,continent_code,continent_name,country_iso_code,country_name,subdivision_1_iso_code,subdivision_1_name,subdivision_2_iso_code,subdivision_2_name,city_name,metro_code,time_zone,is_in_european_union\n" +
			"1001,en,EU,Europe,ES,Spain,MD,Madrid,,,Madrid,,Europe/Madrid,1\n" +
			"1002,en,NA,North America,US,United States,VA,Virginia,,,Ashburn,,America/New_York,0\n" +
			"1003,en,EU,Europe,DE,Germany,HE,Hesse,,,Frankfurt,,Europe/Berlin,1\n",
		"asn-blocks-v4.csv": "network,autonomous_system_number,autonomous_system_organization\n" +
			"8.0.0.0/9,3356,\"Level 3 Parent\"\n" +
			"198.51.100.0/24,13335,\"Cloudflare, Inc.\"\n" +
			"203.0.113.0/24,64500,\n",
		"asn-blocks-v6.csv": "network,autonomous_system_number,autonomous_system_organization\n" +
			"2001:db8::/32,64496,\"Example IPv6 ASN\"\n" +
			"2606:4700::/32,13335,\"Cloudflare, Inc.\"\n",
		"blocks-v4.csv": "network,geoname_id,registered_country_geoname_id,represented_country_geoname_id,is_anonymous_proxy,is_satellite_provider,postal_code,latitude,longitude,accuracy_radius,is_anycast\n" +
			"8.0.0.0/9,1002,1002,,0,0,,39.0438,-77.4874,100,\n" +
			"198.51.100.0/24,1001,1001,,0,0,,40.4168,-3.7038,50,\n" +
			"203.0.113.0/24,,1001,,0,0,,50.1109,8.6821,20,\n",
		"blocks-v6.csv": "network,geoname_id,registered_country_geoname_id,represented_country_geoname_id,is_anonymous_proxy,is_satellite_provider,postal_code,latitude,longitude,accuracy_radius,is_anycast\n" +
			"2001:db8::/32,1001,1001,,0,0,,40.4168,-3.7038,50,\n" +
			"2606:4700::/32,1003,1003,,0,0,,50.1109,8.6821,20,\n",
	}
}

func TestIsLocal_AllSixCIDRsAndIPv6(t *testing.T) {
	cases := []struct {
		ip    string
		local bool
	}{
		// 127.0.0.0/8
		{"127.0.0.1", true},
		{"127.255.255.254", true},
		{"126.255.255.255", false},
		{"128.0.0.1", false},
		// 10.0.0.0/8
		{"10.0.0.0", true},
		{"10.255.255.255", true},
		{"11.0.0.1", false},
		// 172.16.0.0/12
		{"172.15.255.255", false},
		{"172.16.0.0", true},
		{"172.31.255.255", true},
		{"172.32.0.0", false},
		// 192.168.0.0/16
		{"192.167.255.255", false},
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"192.169.0.1", false},
		// 169.254.0.0/16
		{"169.253.255.255", false},
		{"169.254.169.254", true},
		{"169.255.0.1", false},
		// 224.0.0.0/24
		{"224.0.0.0", true},
		{"224.0.0.251", true},
		{"224.0.0.255", true},
		{"224.0.1.1", false},
		// IPv6 local equivalents
		{"::1", true},
		{"fc00::1", true},
		{"fd12:3456::1", true},
		{"fe80::1", true},
		{"ff02::1", true},
		// IPv4-mapped IPv6 local & public
		{"::ffff:192.168.1.10", true},
		{"::ffff:8.8.8.8", false},
		// Public IPs
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"2606:4700::1111", false},
	}

	for _, tc := range cases {
		ip := net.ParseIP(tc.ip)
		if got := IsLocal(ip); got != tc.local {
			t.Errorf("IsLocal(%q) = %v, want %v", tc.ip, got, tc.local)
		}
	}
	if IsLocal(nil) {
		t.Errorf("IsLocal(nil) = true, want false")
	}
}

func TestGeolocate_IPv4IPv6BroadAndEdgeCases(t *testing.T) {
	dir := t.TempDir()
	writeGeoFixtures(t, dir, standardFixtures())

	snap := loadGeolocationFromDir(dir)
	if snap == nil {
		t.Fatal("expected non-nil geoSnapshot")
	}
	geoData.Store(snap)
	defer geoData.Store(nil)

	if snap.asnV4Overlap || snap.asnV6Overlap || snap.cityV4Overlap || snap.cityV6Overlap {
		t.Fatalf("expected non-overlapping flags on disjoint fixtures")
	}

	// Standard IPv4 hit + exact 4-decimal coordinates
	g := geolocate("198.51.100.42")
	if g == nil {
		t.Fatal("expected match for 198.51.100.42")
	}
	if g.Asn != 13335 || g.Aso != "Cloudflare, Inc." ||
		g.CountryCode != "ES" || g.Country != "Spain" || g.City != "Madrid" ||
		g.Latitude != 40.4168 || g.Longitude != -3.7038 {
		t.Fatalf("unexpected geo for 198.51.100.42: %+v", g)
	}

	// IPv4-mapped IPv6 resolves to the same IPv4 block
	gMapped := geolocate("::ffff:198.51.100.42")
	if gMapped == nil || gMapped.CountryCode != "ES" || gMapped.Asn != 13335 {
		t.Fatalf("unexpected geo for IPv4-mapped IPv6: %+v", gMapped)
	}

	// Broad /9 CIDR where second octet differs from network base (legacy getStart bug!)
	gBroad := geolocate("8.120.45.67")
	if gBroad == nil || gBroad.Asn != 3356 || gBroad.City != "Ashburn" || gBroad.CountryCode != "US" {
		t.Fatalf("expected broad /9 hit for 8.120.45.67, got %+v", gBroad)
	}

	// Empty ASO defaults to "-" and empty geoname_id still returns lat/lon + ASN
	gNoLoc := geolocate("203.0.113.5")
	if gNoLoc == nil || gNoLoc.Asn != 64500 || gNoLoc.Aso != "-" ||
		gNoLoc.City != "" || gNoLoc.Latitude != 50.1109 || gNoLoc.Longitude != 8.6821 {
		t.Fatalf("unexpected geo for 203.0.113.5: %+v", gNoLoc)
	}

	// IPv6 hit
	gV6 := geolocate("2606:4700::1111")
	if gV6 == nil || gV6.Asn != 13335 || gV6.Aso != "Cloudflare, Inc." ||
		gV6.CountryCode != "DE" || gV6.City != "Frankfurt" ||
		gV6.Latitude != 50.1109 || gV6.Longitude != 8.6821 {
		t.Fatalf("unexpected geo for 2606:4700::1111: %+v", gV6)
	}

	// Boundary checks: network address and broadcast address
	if geolocate("198.51.100.0") == nil || geolocate("198.51.100.255") == nil {
		t.Fatal("expected boundary IPs 198.51.100.0 and .255 to match")
	}
	// Just outside range -> nil (unknown public IP)
	if gMiss := geolocate("198.51.101.0"); gMiss != nil {
		t.Fatalf("expected nil for unknown public IP, got %+v", gMiss)
	}
	// Local and invalid IPs -> nil
	if geolocate("192.168.1.1") != nil || geolocate("not-an-ip") != nil || geolocate("") != nil {
		t.Fatal("expected nil for local/invalid/empty IPs")
	}
}

func TestGeolocate_NestedAndUnsortedCIDRs(t *testing.T) {
	dir := t.TempDir()
	files := standardFixtures()
	// Unsorted + nested IPv4 and IPv6 prefixes
	files["blocks-v4.csv"] = "network,geoname_id,reg,rep,anon,sat,postal,latitude,longitude,acc,anycast\n" +
		"198.51.100.128/25,1003,1003,,0,0,,50.1109,8.6821,20,\n" +
		"198.51.0.0/16,1001,1001,,0,0,,40.4168,-3.7038,50,\n"
	files["blocks-v6.csv"] = "network,geoname_id,reg,rep,anon,sat,postal,latitude,longitude,acc,anycast\n" +
		"2001:db8:1::/48,1003,1003,,0,0,,50.1109,8.6821,20,\n" +
		"2001:db8::/32,1001,1001,,0,0,,40.4168,-3.7038,50,\n"
	writeGeoFixtures(t, dir, files)

	snap := loadGeolocationFromDir(dir)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	geoData.Store(snap)
	defer geoData.Store(nil)

	if !snap.cityV4Overlap || !snap.cityV6Overlap {
		t.Fatalf("expected overlap flags true for nested CIDRs")
	}

	// Narrower child wins
	if g := geolocate("198.51.100.200"); g == nil || g.City != "Frankfurt" {
		t.Fatalf("expected child /25 match Frankfurt, got %+v", g)
	}
	// Sibling outside child falls back to parent /16
	if g := geolocate("198.51.100.10"); g == nil || g.City != "Madrid" {
		t.Fatalf("expected parent /16 match Madrid, got %+v", g)
	}
	if g := geolocate("198.51.200.1"); g == nil || g.City != "Madrid" {
		t.Fatalf("expected parent /16 match Madrid after child range, got %+v", g)
	}
	// Same for IPv6
	if g := geolocate("2001:db8:1::42"); g == nil || g.City != "Frankfurt" {
		t.Fatalf("expected IPv6 child /48 match Frankfurt, got %+v", g)
	}
	if g := geolocate("2001:db8:2::42"); g == nil || g.City != "Madrid" {
		t.Fatalf("expected IPv6 parent /32 match Madrid after child range, got %+v", g)
	}
}

func TestGeolocate_ConcurrentReloadRace(t *testing.T) {
	dir := t.TempDir()
	writeGeoFixtures(t, dir, standardFixtures())
	geoData.Store(loadGeolocationFromDir(dir))
	defer geoData.Store(nil)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ips := []string{"198.51.100.42", "2606:4700::1111", "10.0.0.1", "93.184.216.34"}
			for j := 0; ; j++ {
				select {
				case <-stop:
					return
				default:
				}
				_ = geolocate(ips[(id+j)%len(ips)])
			}
		}(i)
	}

	for r := 0; r < 5; r++ {
		if snap := loadGeolocationFromDir(dir); snap != nil {
			geoData.Store(snap)
		}
	}
	close(stop)
	wg.Wait()
}

func BenchmarkGeolocateV4Hit(b *testing.B) {
	// Build a 100k-entry synthetic table for realistic binary search timing
	snap := &geoSnapshot{
		asnV4:     make([]asnBlockV4, 100000),
		cityV4:    make([]cityBlockV4, 100000),
		locations: []cityLocation{{}, {countryISOCode: "ES", countryName: "Spain", cityName: "Madrid"}},
		asos:      []string{"-", "Example ASN"},
	}
	for i := 0; i < 100000; i++ {
		base := uint32(0x0b000000) + uint32(i)*256 // starts at 11.0.0.0/24
		snap.asnV4[i] = asnBlockV4{start: base, end: base + 255, asn: uint32(1000 + i), asoIdx: 1}
		snap.cityV4[i] = cityBlockV4{start: base, end: base + 255, locIdx: 1, latE6: 40416800, lonE6: -3703800}
	}
	geoData.Store(snap)
	defer geoData.Store(nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if g := geolocate("11.123.45.67"); g == nil {
			b.Fatal("expected hit")
		}
	}
}

func BenchmarkIsLocal(b *testing.B) {
	ip := net.ParseIP("192.168.1.42")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !IsLocal(ip) {
			b.Fatal("expected local")
		}
	}
}

func TestPrefixRangeV4AndV6(t *testing.T) {
	// Ensure /0 and /32 and /128 never overflow bit shifts
	dir := t.TempDir()
	files := standardFixtures()
	files["blocks-v4.csv"] += "198.51.100.77/32,1003,1003,,0,0,,50.1109,8.6821,1,\n"
	files["blocks-v6.csv"] += "2001:db8::99/128,1003,1003,,0,0,,50.1109,8.6821,1,\n"
	writeGeoFixtures(t, dir, files)

	snap := loadGeolocationFromDir(dir)
	geoData.Store(snap)
	defer geoData.Store(nil)

	if g := geolocate("198.51.100.77"); g == nil || g.City != "Frankfurt" {
		t.Fatalf("expected /32 hit Frankfurt, got %+v", g)
	}
	if g := geolocate("2001:db8::99"); g == nil || g.City != "Frankfurt" {
		t.Fatalf("expected /128 hit Frankfurt, got %+v", g)
	}
	_ = fmt.Sprintf("loaded %d v4 city blocks", len(snap.cityV4))
}

func TestLoadGeolocationFromDir_RejectsPartialFailure(t *testing.T) {
	dir := t.TempDir()
	files := standardFixtures()
	delete(files, "blocks-v6.csv") // simulate missing/unreadable required file
	writeGeoFixtures(t, dir, files)

	if snap := loadGeolocationFromDir(dir); snap != nil {
		t.Fatalf("expected nil snapshot when required CSV fails to load, got %+v", snap)
	}
}
