package netflow

import (
	"fmt"
	"reflect"
	"strings"

	goflownetflow "github.com/netsampler/goflow2/decoders/netflow"
	"github.com/netsampler/goflow2/decoders/netflowlegacy"
	"github.com/tehmaze/netflow/netflow1"
	"github.com/tehmaze/netflow/netflow6"
	"github.com/tehmaze/netflow/netflow7"
	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

// processMessage converts a decoded netflow message to log entries and sends them to the queue.
func processMessage(remote string, message interface{}, queue chan *plugins.Log) error {
	var metrics []Metric

	switch m := message.(type) {
	// goflow2 types (primary for v5, v9, IPFIX)
	case netflowlegacy.PacketNetFlowV5:
		metrics = PrepareGoflowV5(remote, &m)
	case *netflowlegacy.PacketNetFlowV5:
		metrics = PrepareGoflowV5(remote, m)
	case goflownetflow.NFv9Packet:
		metrics = PrepareGoflowV9(remote, &m)
	case *goflownetflow.NFv9Packet:
		metrics = PrepareGoflowV9(remote, m)
	case goflownetflow.IPFIXPacket:
		metrics = PrepareGoflowIPFIX(remote, &m)
	case *goflownetflow.IPFIXPacket:
		metrics = PrepareGoflowIPFIX(remote, m)

	// tehmaze types (fallback for v1, v6, v7)
	case *netflow1.Packet:
		metrics = PrepareV1(remote, m)
	case *netflow6.Packet:
		metrics = PrepareV6(remote, m)
	case *netflow7.Packet:
		metrics = PrepareV7(remote, m)

	default:
		return fmt.Errorf("unknown netflow message type: %T", m)
	}

	messages := dumpMetrics(metrics)

	for _, msg := range messages {
		message, _, err := entities.ValidateString(msg, false)
		if err != nil {
			utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
			continue
		}
		select {
		case queue <- &plugins.Log{
			DataType:   string(config.DataTypeNetflow),
			DataSource: remote,
			Raw:        msg,
		}:
		default:
			// queue full — drop log rather than block the UDP listener
		}
	}

	return nil
}

// dumpMetrics converts metrics to key-value string format.
func dumpMetrics(metrics []Metric) []string {
	var allKVPairs []string
	for _, metric := range metrics {
		t := reflect.TypeOf(metric)
		v := reflect.ValueOf(metric)
		var kvPairs []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			if value.String() == "" {
				continue
			}
			header := field.Tag.Get("header")
			kvPairs = append(kvPairs, fmt.Sprintf("%s=\"%v\"", header, value))
		}
		allKVPairs = append(allKVPairs, strings.Join(kvPairs, " "))
	}
	return allKVPairs
}
