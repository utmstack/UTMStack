package parser

import go_sdk "github.com/threatwinds/go-sdk"

type Parser interface {
	ProcessData(logMessage interface{}, datasource string, queue chan *go_sdk.Log) error
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
