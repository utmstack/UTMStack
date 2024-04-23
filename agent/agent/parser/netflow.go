package parser

import (
	"fmt"
	"sync"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/ipfix"
	"github.com/tehmaze/netflow/netflow1"
	"github.com/tehmaze/netflow/netflow5"
	"github.com/tehmaze/netflow/netflow6"
	"github.com/tehmaze/netflow/netflow7"
	"github.com/tehmaze/netflow/netflow9"
	"github.com/threatwinds/logger"
	pnf "github.com/utmstack/UTMStack/agent/agent/parser/netflow"
)

var (
	netflowParser = NetflowParser{}
	netflowOnce   sync.Once
)

type NetflowParser struct {
}

func GetNetflowParser() *NetflowParser {
	netflowOnce.Do(func() {
		netflowParser = NetflowParser{}
	})
	return &netflowParser
}

type NetflowObject struct {
	Remote  string
	Message netflow.Message
}

func (p *NetflowParser) ProcessData(logBatch interface{}, h *logger.Logger) (map[string][]string, error) {
	var metrics []pnf.Metric
	var remote string

	switch l := logBatch.(type) {
	case NetflowObject:
		remote = l.Remote
		switch m := l.Message.(type) {
		case *netflow1.Packet:
			metrics = pnf.PrepareV1(remote, m)
		case *netflow5.Packet:
			metrics = pnf.PrepareV5(remote, m)
		case *netflow6.Packet:
			metrics = pnf.PrepareV6(remote, m)
		case *netflow7.Packet:
			metrics = pnf.PrepareV7(remote, m)
		case *netflow9.Packet:
			metrics = pnf.PrepareV9(remote, m)
		case *ipfix.Message:
			metrics = pnf.PrepareIPFIX(remote, m)
		}
	default:
		return nil, fmt.Errorf("unknown log batch type: %T", l)
	}

	return map[string][]string{
		"netflow": pnf.Dump(metrics, remote),
	}, nil
}
