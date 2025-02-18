package utils

import (
	"github.com/tidwall/gjson"
	"net"
)

type SavedField struct {
	Field string `yaml:"field"`
	Alias string `yaml:"alias"`
}

func ExtractDetails(save []SavedField, l string) map[string]string {
	var fields = map[string]string{
		"id": gjson.Get(l, "id").String(),
	}

	//User saved fields
	for _, save := range save {
		saveValue := gjson.Get(l, save.Field).String()
		if saveValue == "" {
			continue
		}

		fields[save.Alias] = saveValue
	}

	_, srcHostOk := fields["SourceHost"]
	_, dstHostOk := fields["DestinationHost"]
	_, srcIPOk := fields["SourceIP"]
	_, dstIPOk := fields["DestinationIP"]

	// Try to resolve SourceHost if SourceIP exists but not SourceHost
	if !srcHostOk && srcIPOk {
		host, _ := net.LookupHost(fields["SourceIP"])
		fields["SourceHost"] = host[0]
	}

	// Try to resolve DestinationHost if DestinationIP exists but not DestinationHost
	if !dstHostOk && dstIPOk {
		host, _ := net.LookupHost(fields["DestinationIP"])
		fields["DestinationHost"] = host[0]
	}

	// Try to resolve SourceIP if SourceHost exists but not SourceIP
	if srcHostOk && !srcIPOk {
		ip, _ := net.LookupIP(fields["SourceHost"])
		if len(ip) != 0 && ip[0].String() != "<nil>" {
			fields["SourceIP"] = ip[0].String()
		}
	}

	// Try to resolve DestinationIP if DestinationHost exists but not DestinationIP
	if dstHostOk && !dstIPOk {
		ip, _ := net.LookupIP(fields["DestinationHost"])
		if len(ip) != 0 && ip[0].String() != "<nil>" {
			fields["DestinationIP"] = ip[0].String()
		}
	}

	return fields
}
