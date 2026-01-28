package parser

import (
	"fmt"
	"sync"

	goflownetflow "github.com/netsampler/goflow2/decoders/netflow"
	"github.com/netsampler/goflow2/decoders/netflowlegacy"
	"github.com/tehmaze/netflow/netflow1"
	"github.com/tehmaze/netflow/netflow6"
	"github.com/tehmaze/netflow/netflow7"
	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	pnf "github.com/utmstack/UTMStack/agent/parser/netflow"
	"github.com/utmstack/UTMStack/agent/utils"
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
	Message interface{}
}

func (p *NetflowParser) ProcessData(logMessage interface{}, _ string, queue chan *plugins.Log) error {
	var metrics []pnf.Metric
	var remote string

	switch l := logMessage.(type) {
	case NetflowObject:
		remote = l.Remote
		switch m := l.Message.(type) {
		// goflow2 types (primary for v5, v9, IPFIX)
		case netflowlegacy.PacketNetFlowV5:
			metrics = pnf.PrepareGoflowV5(remote, &m)
		case *netflowlegacy.PacketNetFlowV5:
			metrics = pnf.PrepareGoflowV5(remote, m)
		case goflownetflow.NFv9Packet:
			metrics = pnf.PrepareGoflowV9(remote, &m)
		case *goflownetflow.NFv9Packet:
			metrics = pnf.PrepareGoflowV9(remote, m)
		case goflownetflow.IPFIXPacket:
			metrics = pnf.PrepareGoflowIPFIX(remote, &m)
		case *goflownetflow.IPFIXPacket:
			metrics = pnf.PrepareGoflowIPFIX(remote, m)

		// tehmaze types (fallback for v1, v6, v7)
		case *netflow1.Packet:
			metrics = pnf.PrepareV1(remote, m)
		case *netflow6.Packet:
			metrics = pnf.PrepareV6(remote, m)
		case *netflow7.Packet:
			metrics = pnf.PrepareV7(remote, m)

		default:
			return fmt.Errorf("unknown netflow message type: %T", m)
		}
	default:
		return fmt.Errorf("unknown log batch type: %T", l)
	}

	messages := pnf.Dump(metrics)

	for _, msg := range messages {
		message, _, err := entities.ValidateString(msg, false)
		if err != nil {
			utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
			continue
		}
		queue <- &plugins.Log{
			DataType:   string(config.DataTypeNetflow),
			DataSource: remote,
			Raw:        msg,
		}
	}

	return nil
}
