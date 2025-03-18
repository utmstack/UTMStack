package netflow

import (
	"fmt"
	"net"
	"time"

	"github.com/tehmaze/netflow/netflow7"
)

func PrepareV7(addr string, p *netflow7.Packet) []Metric {
	nfExporter, _, _ := net.SplitHostPort(addr)
	var metrics []Metric
	var met Metric
	for _, r := range p.Records {
		met = Metric{OutBytes: "0", InBytes: "0", OutPacket: "0", InPacket: "0", NFSender: nfExporter}
		met.FlowVersion = "Netflow-V7"
		met.First = time.Unix(int64(r.First), 0).Format(time.RFC3339Nano)
		met.Last = time.Unix(int64(r.Last), 0).Format(time.RFC3339Nano)
		met.Protocol = ProtoToName(fmt.Sprintf("%v", r.Protocol))
		met.Bytes = fmt.Sprintf("%v", r.Bytes)
		met.Packets = fmt.Sprintf("%v", r.Packets)
		met.TCPFlags = fmt.Sprintf("%v", r.TCPFlags)
		met.SrcAs = fmt.Sprintf("%v", r.SrcAS)
		met.DstAs = fmt.Sprintf("%v", r.DstAS)
		met.SrcMask = fmt.Sprintf("%v", r.SrcMask)
		met.DstMask = fmt.Sprintf("%v", r.DstMask)
		met.NextHop = fmt.Sprintf("%v", r.NextHop)

		met.SrcIP = fmt.Sprintf("%v", r.SrcAddr)
		met.DstIP = fmt.Sprintf("%v", r.DstAddr)

		met.SrcPort = fmt.Sprintf("%v", r.SrcPort)

		met.DstPort = fmt.Sprintf("%v", r.DstPort)

		metrics = append(metrics, met)

	}

	return metrics
}
