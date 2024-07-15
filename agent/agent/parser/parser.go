package parser

import "github.com/threatwinds/go-sdk/plugins"

type Parser interface {
	ProcessData(logMessage interface{}, datasource string, queue chan *plugins.Log) error
}

func GetParser(typ string) Parser {
	switch typ {
	case "cisco":
		return GetCiscoParser()
	case "beats":
		return GetBeatsParser()
	case "netflow":
		return GetNetflowParser()
	default:
		return nil
	}
}
