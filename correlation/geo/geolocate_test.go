package geo

import (
	"net"
	"testing"
)

func TestIsLocal(t *testing.T) {
	type args struct {
		a net.IP
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Local IP",
			args: args{net.ParseIP("127.0.0.1")},
			want: true,
		},
		{
			name: "Local IP",
			args: args{net.ParseIP("10.0.0.1")},
			want: true,
		},
		{
			name: "Local IP",
			args: args{net.ParseIP("172.16.0.1")},
			want: true,
		},
		{
			name: "Local IP",
			args: args{net.ParseIP("192.168.0.1")},
			want: true,
		},
		{
			name: "Local IP",
			args: args{net.ParseIP("169.254.0.1")},
			want: true,
		},
		{
			name: "Local IP",
			args: args{net.ParseIP("224.0.0.1")},
			want: true,
		},
		{
			name: "Non-local IP",
			args: args{net.ParseIP("8.8.8.8")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocal(tt.args.a); got != tt.want {
				t.Errorf("IsLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExternalOnce(t *testing.T) {
	tests := []struct {
		name string
		want net.IP
	}{
		{
			name: "Get external IP",
			want: net.ParseIP("8.8.8.8"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExternalOnce(); !got.Equal(tt.want) {
				t.Errorf("GetExternalOnce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeolocate(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want map[string]string
	}{
		{
			name: "Local IP",
			ip:   "127.0.0.1",
			want: map[string]string{
				"country":             "United States",
				"countryCode":         "US",
				"city":                "Mountain View",
				"latitude":            "37.419200",
				"longitude":           "-122.057400",
				"accuracyRadius":      "1000",
				"asn":                 "AS15169",
				"aso":                 "Google LLC",
				"isSatelliteProvider": "false",
				"isAnonymousProxy":    "false",
			},
		},
		{
			name: "Non-local IP",
			ip:   "8.8.8.8",
			want: map[string]string{
				"country":             "United States",
				"countryCode":         "US",
				"city":                "Mountain View",
				"latitude":            "37.419200",
				"longitude":           "-122.057400",
				"accuracyRadius":      "1000",
				"asn":                 "AS15169",
				"aso":                 "Google LLC",
				"isSatelliteProvider": "false",
				"isAnonymousProxy":    "false",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Geolocate(tt.ip); !compareMaps(got, tt.want) {
				t.Errorf("Geolocate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}