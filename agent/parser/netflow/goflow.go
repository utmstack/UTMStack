package netflow

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	goflownetflow "github.com/netsampler/goflow2/decoders/netflow"
	"github.com/netsampler/goflow2/decoders/netflowlegacy"
)

// PrepareGoflowV5 converts goflow2 NetFlow v5 packet to metrics
func PrepareGoflowV5(addr string, p *netflowlegacy.PacketNetFlowV5) []Metric {
	nfExporter, _, _ := net.SplitHostPort(addr)
	var metrics []Metric

	for _, r := range p.Records {
		met := Metric{OutBytes: "0", InBytes: "0", OutPacket: "0", InPacket: "0", NFSender: nfExporter}
		met.FlowVersion = "Netflow-V5"

		// Convert timestamps (First and Last are relative to SysUptime in milliseconds)
		met.First = time.Unix(int64(p.UnixSecs), int64(p.UnixNSecs)).Add(-time.Duration(p.SysUptime-r.First) * time.Millisecond).Format(time.RFC3339Nano)
		met.Last = time.Unix(int64(p.UnixSecs), int64(p.UnixNSecs)).Add(-time.Duration(p.SysUptime-r.Last) * time.Millisecond).Format(time.RFC3339Nano)

		met.Protocol = ProtoToName(fmt.Sprintf("%v", r.Proto))
		met.Bytes = fmt.Sprintf("%v", r.DOctets)
		met.Packets = fmt.Sprintf("%v", r.DPkts)
		met.TCPFlags = fmt.Sprintf("%v", r.TCPFlags)
		met.SrcAs = fmt.Sprintf("%v", r.SrcAS)
		met.DstAs = fmt.Sprintf("%v", r.DstAS)
		met.SrcMask = fmt.Sprintf("%v", r.SrcMask)
		met.DstMask = fmt.Sprintf("%v", r.DstMask)

		// Convert uint32 IPs to string
		srcIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(srcIP, r.SrcAddr)
		met.SrcIP = srcIP.String()

		dstIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(dstIP, r.DstAddr)
		met.DstIP = dstIP.String()

		nextHop := make(net.IP, 4)
		binary.BigEndian.PutUint32(nextHop, r.NextHop)
		met.NextHop = nextHop.String()

		met.SrcPort = fmt.Sprintf("%v", r.SrcPort)
		met.DstPort = fmt.Sprintf("%v", r.DstPort)

		met.InEthernet = fmt.Sprintf("%v", r.Input)
		met.OutEthernet = fmt.Sprintf("%v", r.Output)

		metrics = append(metrics, met)
	}

	return metrics
}

// PrepareGoflowV9 converts goflow2 NetFlow v9 packet to metrics
func PrepareGoflowV9(addr string, p *goflownetflow.NFv9Packet) []Metric {
	nfExporter, _, _ := net.SplitHostPort(addr)
	var metrics []Metric

	for _, fs := range p.FlowSets {
		// FlowSets can be TemplateFlowSet, DataFlowSet, or OptionsDataFlowSet
		switch ds := fs.(type) {
		case goflownetflow.DataFlowSet:
			for _, record := range ds.Records {
				met := Metric{OutBytes: "0", InBytes: "0", OutPacket: "0", InPacket: "0", NFSender: nfExporter}
				met.FlowVersion = "Netflow-V9"

				for _, field := range record.Values {
					extractFieldValue(&met, field.Type, field.Value, p.UnixSeconds)
				}

				metrics = append(metrics, met)
			}
		}
	}

	return metrics
}

// PrepareGoflowIPFIX converts goflow2 IPFIX packet to metrics
func PrepareGoflowIPFIX(addr string, p *goflownetflow.IPFIXPacket) []Metric {
	nfExporter, _, _ := net.SplitHostPort(addr)
	var metrics []Metric

	for _, fs := range p.FlowSets {
		// FlowSets can be TemplateFlowSet, DataFlowSet, or OptionsDataFlowSet
		switch ds := fs.(type) {
		case goflownetflow.DataFlowSet:
			for _, record := range ds.Records {
				met := Metric{OutBytes: "0", InBytes: "0", OutPacket: "0", InPacket: "0", NFSender: nfExporter}
				met.FlowVersion = "IPFIX"

				for _, field := range record.Values {
					extractFieldValue(&met, field.Type, field.Value, p.ExportTime)
				}

				metrics = append(metrics, met)
			}
		}
	}

	return metrics
}

// extractFieldValue extracts field values based on IPFIX/NetFlow v9 field type IDs
// Field type IDs are defined in RFC 5102 and RFC 3954
func extractFieldValue(met *Metric, fieldType uint16, value interface{}, exportTime uint32) {
	switch fieldType {
	// Timestamps
	case 21: // flowEndSysUpTime
		if v, ok := toUint32(value); ok {
			met.Last = time.Unix(int64(exportTime), 0).Add(-time.Duration(v) * time.Millisecond).Format(time.RFC3339Nano)
		}
	case 22: // flowStartSysUpTime
		if v, ok := toUint32(value); ok {
			met.First = time.Unix(int64(exportTime), 0).Add(-time.Duration(v) * time.Millisecond).Format(time.RFC3339Nano)
		}
	case 150: // flowStartSeconds
		if v, ok := toUint32(value); ok {
			met.First = time.Unix(int64(v), 0).Format(time.RFC3339Nano)
		}
	case 151: // flowEndSeconds
		if v, ok := toUint32(value); ok {
			met.Last = time.Unix(int64(v), 0).Format(time.RFC3339Nano)
		}

	// Byte and packet counts
	case 1: // octetDeltaCount
		met.Bytes = fmt.Sprintf("%v", value)
	case 2: // packetDeltaCount
		met.Packets = fmt.Sprintf("%v", value)
	case 85: // octetTotalCount
		met.Bytes = fmt.Sprintf("%v", value)
	case 86: // packetTotalCount
		met.Packets = fmt.Sprintf("%v", value)

	// Interfaces
	case 10: // ingressInterface
		met.InEthernet = fmt.Sprintf("%v", value)
	case 14: // egressInterface
		met.OutEthernet = fmt.Sprintf("%v", value)

	// IPv4 addresses
	case 8: // sourceIPv4Address
		if ip, ok := toIP(value); ok {
			met.SrcIP = ip.String()
		} else {
			met.SrcIP = fmt.Sprintf("%v", value)
		}
	case 12: // destinationIPv4Address
		if ip, ok := toIP(value); ok {
			met.DstIP = ip.String()
		} else {
			met.DstIP = fmt.Sprintf("%v", value)
		}
	case 15: // ipNextHopIPv4Address
		if ip, ok := toIP(value); ok {
			met.NextHop = ip.String()
		} else {
			met.NextHop = fmt.Sprintf("%v", value)
		}

	// IPv6 addresses
	case 27: // sourceIPv6Address
		if ip, ok := toIP(value); ok {
			met.SrcIP = ip.String()
		} else {
			met.SrcIP = fmt.Sprintf("%v", value)
		}
	case 28: // destinationIPv6Address
		if ip, ok := toIP(value); ok {
			met.DstIP = ip.String()
		} else {
			met.DstIP = fmt.Sprintf("%v", value)
		}
	case 62: // ipNextHopIPv6Address
		if ip, ok := toIP(value); ok {
			met.NextHop = ip.String()
		} else {
			met.NextHop = fmt.Sprintf("%v", value)
		}

	// Protocol and ports
	case 4: // protocolIdentifier
		met.Protocol = ProtoToName(fmt.Sprintf("%v", value))
	case 7: // sourceTransportPort
		met.SrcPort = fmt.Sprintf("%v", value)
	case 11: // destinationTransportPort
		met.DstPort = fmt.Sprintf("%v", value)

	// Masks
	case 9: // sourceIPv4PrefixLength
		met.SrcMask = fmt.Sprintf("%v", value)
	case 13: // destinationIPv4PrefixLength
		met.DstMask = fmt.Sprintf("%v", value)
	case 29: // sourceIPv6PrefixLength
		met.SrcMask = fmt.Sprintf("%v", value)
	case 30: // destinationIPv6PrefixLength
		met.DstMask = fmt.Sprintf("%v", value)

	// AS numbers
	case 16: // bgpSourceAsNumber
		met.SrcAs = fmt.Sprintf("%v", value)
	case 17: // bgpDestinationAsNumber
		met.DstAs = fmt.Sprintf("%v", value)

	// TCP flags
	case 6: // tcpControlBits
		met.TCPFlags = fmt.Sprintf("%v", value)

	// Flow direction
	case 61: // flowDirection
		switch fmt.Sprintf("%v", value) {
		case "0":
			met.Direction = "Ingress"
		case "1":
			met.Direction = "Egress"
		default:
			met.Direction = fmt.Sprintf("%v", value)
		}
	}
}

// toUint32 attempts to convert value to uint32
func toUint32(value interface{}) (uint32, bool) {
	switch v := value.(type) {
	case uint32:
		return v, true
	case uint64:
		return uint32(v), true
	case int64:
		return uint32(v), true
	case int:
		return uint32(v), true
	case []byte:
		if len(v) == 4 {
			return binary.BigEndian.Uint32(v), true
		}
	}
	return 0, false
}

// toIP attempts to convert value to net.IP
func toIP(value interface{}) (net.IP, bool) {
	switch v := value.(type) {
	case net.IP:
		return v, true
	case []byte:
		if len(v) == 4 || len(v) == 16 {
			return net.IP(v), true
		}
	case uint32:
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, v)
		return ip, true
	}
	return nil, false
}
