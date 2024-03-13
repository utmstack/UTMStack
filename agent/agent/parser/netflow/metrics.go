package netflow

type Metric struct {
	FlowVersion         string `header:"version"`
	NFSender            string `header:"exporter"`
	Last                string `header:"last"`
	First               string `header:"first"`
	Bytes               string `header:"bytes"`
	Packets             string `header:"packets"`
	InBytes             string `header:"bytesIn"`
	InPacket            string `header:"packetsIn"`
	OutBytes            string `header:"bytesOut"`
	OutPacket           string `header:"packetsOut"`
	InEthernet          string `header:"inEth"`
	OutEthernet         string `header:"outEth"`
	SrcIP               string `header:"srcIp"`
	SrcIp2lCountryShort string `header:"sCountry_S"`
	SrcIp2lCountryLong  string `header:"sCountry_L"`
	SrcIp2lState        string `header:"sState"`
	SrcIp2lCity         string `header:"sCity"`
	SrcIp2lLat          string `header:"sLat"`
	SrcIp2lLong         string `header:"sLong"`
	DstIP               string `header:"dstIp"`
	DstIp2lCountryShort string `header:"dCountry_S"`
	DstIp2lCountryLong  string `header:"dCountry_L"`
	DstIp2lState        string `header:"dState"`
	DstIp2lCity         string `header:"dCity"`
	DstIp2lLat          string `header:"dLat"`
	DstIp2lLong         string `header:"dLong"`
	Protocol            string `header:"proto"`
	SrcToS              string `header:"srcToS"`
	SrcPort             string `header:"srcPort"`
	DstPort             string `header:"dstPort"`
	FlowSamplerId       string `header:"flowId"`
	VendorPROPRIETARY   string `header:"vendor"`
	NextHop             string `header:"nextHop"`
	DstMask             string `header:"dstMask"`
	SrcMask             string `header:"srcMask"`
	TCPFlags            string `header:"tcpFlags"`
	Direction           string `header:"direction"`
	DstAs               string `header:"dstAs"`
	SrcAs               string `header:"srcAs"`
}
