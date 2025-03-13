package netflow

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/tehmaze/netflow/ipfix"
)

func PrepareIPFIX(addr string, m *ipfix.Message) []Metric {
	nfExporter, _, _ := net.SplitHostPort(addr)

	var metrics []Metric
	var met Metric

	for _, ds := range m.DataSets {
		if ds.Records == nil {
			continue
		}

		for _, dr := range ds.Records {
			met = Metric{OutBytes: "0", InBytes: "0", OutPacket: "0", InPacket: "0", NFSender: nfExporter}
			met.FlowVersion = "IPFIX"
			for _, f := range dr.Fields {
				if f.Translated != nil {
					if f.Translated.Name != "" {
						//fmt.Printf("        NN %s: %v\n", f.Translated.Name, f.Translated.Value)
						switch strings.ToLower(f.Translated.Name) {
						case strings.ToLower("flowEndSysUpTime"):
							value, ok := f.Translated.Value.(int64)
							if ok {
								met.First = time.Unix(value, 0).Format(time.RFC3339Nano)
							} else {
								value, ok := f.Translated.Value.(uint32)
								if ok {
									met.First = time.Unix(int64(value), 0).Format(time.RFC3339Nano)
								}
							}

						case strings.ToLower("flowStartSysUpTime"):
							value, ok := f.Translated.Value.(int64)
							if ok {
								met.Last = time.Unix(value, 0).Format(time.RFC3339Nano)
							} else {
								value, ok := f.Translated.Value.(uint32)
								if ok {
									met.Last = time.Unix(int64(value), 0).Format(time.RFC3339Nano)
								}
							}

						case strings.ToLower("octetDeltaCount"):
							met.Bytes = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("packetDeltaCount"):
							met.Packets = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("ingressInterface"):
							met.InEthernet = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("egressInterface"):
							met.OutEthernet = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("sourceIPv4Address"):
							met.SrcIP = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("destinationIPv4Address"):
							met.DstIP = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("protocolIdentifier"):
							met.Protocol = ProtoToName(fmt.Sprintf("%v", f.Translated.Value))

						case strings.ToLower("sourceTransportPort"):
							met.SrcPort = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("destinationTransportPort"):
							met.DstPort = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("ipNextHopIPv4Address"):
							met.NextHop = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("destinationIPv4PrefixLength"):
							met.DstMask = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("sourceIPv4PrefixLength"):
							met.SrcMask = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("tcpControlBits"):
							met.TCPFlags = fmt.Sprintf("%v", f.Translated.Value)

						case strings.ToLower("flowDirection"):
							met.Direction = fmt.Sprintf("%v", f.Translated.Value)
							switch met.Direction {
							case "0":
								met.Direction = "Ingress"
							case "1":
								met.Direction = "Egress"
							}
						}
					} else {
						//fmt.Printf("        TT %d: %v\n", f.Translated.Type, f.Bytes)
						return nil
					}
				} else {
					//fmt.Printf("        RR %d: %v (raw)\n", f.Type, f.Bytes)
					return nil
				}
			}
			metrics = append(metrics, met)
		}
	}

	return metrics
}
