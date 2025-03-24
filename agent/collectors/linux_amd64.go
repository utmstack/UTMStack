//go:build linux && amd64
// +build linux,amd64

package collectors

func getCollectorsInstances() []Collector {
	var collectors []Collector
	collectors = append(collectors, Filebeat{})

	return collectors
}
