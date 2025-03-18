package parser

type Parser interface {
	ProcessData(logMessage interface{}) (map[string][]string, error)
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
