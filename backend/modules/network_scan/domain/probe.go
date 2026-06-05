package domain

// NetworkScan is the JSON object returned by the external probe scanner.
// Mirrors the Java POJO `com.park.utmstack.domain.network_scan.NetworkScan`.
type NetworkScan struct {
	IP          string     `json:"ip"`
	MAC         string     `json:"mac"`
	OS          string     `json:"os"`
	Name        string     `json:"name"`
	Alive       bool       `json:"alive"`
	AddressList []string   `json:"addressList"`
	AliasList   []string   `json:"aliasList"`
	Ports       []ProbePort `json:"ports"`
}

type ProbePort struct {
	Port int    `json:"port"`
	TCP  string `json:"tcp"`
	UDP  string `json:"udp"`
}
