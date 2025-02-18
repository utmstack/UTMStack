package correlation

import (
	"net"

	"github.com/utmstack/UTMStack/correlation/geo"
	"github.com/utmstack/UTMStack/correlation/rules"
	"github.com/tidwall/gjson"
)

func processResponse(logs []string, rule rules.Rule, save []rules.SavedField, tmpLogs *[20][]map[string]string,
	steps, step, minCount int) {
	if len(logs) >= func()int{
		switch minCount{
		case 0:
			return 1
		default:
			return minCount
		}
	}() {
		for _, l := range logs {
			var fields = map[string]string{
				"id": gjson.Get(l, "id").String(),
			}
			//User saved fields
			for _, save := range save {
				fields[save.Alias] = gjson.Get(l, save.Field).String()
			}
			// Try to resolve SourceHost if SourceIP exists but not SourceHost
			if fields["SourceHost"] == "" && fields["SourceIP"] != "" {
				host, _ := net.LookupHost(fields["SourceIP"])
				fields["SourceHost"] = host[0]
			}
			// Try to resolve DestinationHost if DestinationIP exists but not DestinationHost
			if fields["DestinationHost"] == "" && fields["DestinationIP"] != "" {
				host, _ := net.LookupHost(fields["DestinationIP"])
				fields["DestinationHost"] = host[0]
			}
			// Try to resolve SourceIP if SourceHost exists but not SourceIP
			if fields["SourceHost"] != "" && fields["SourceIP"] == "" {
				ip, _ := net.LookupIP(fields["SourceHost"])
				if len(ip) != 0 && ip[0].String() != "<nil>" {
					fields["SourceIP"] = ip[0].String()
				}
			}
			// Try to resolve DestinationIP if DestinationHost exists but not DestinationIP
			if fields["DestinationHost"] != "" && fields["DestinationIP"] == "" {
				ip, _ := net.LookupIP(fields["DestinationHost"])
				if len(ip) != 0 && ip[0].String() != "<nil>" {
					fields["DestinationIP"] = ip[0].String()
				}
			}
			// Try to geolocate SourceIP if exists
			if fields["SourceIP"] != "" {
				location := geo.Geolocate(fields["SourceIP"])
				fields["SourceCountry"] = location["country"]
				fields["SourceCountryCode"] = location["countryCode"]
				fields["SourceCity"] = location["city"]
				fields["SourceLat"] = location["latitude"]
				fields["SourceLon"] = location["longitude"]
				fields["SourceAccuracyRadius"] = location["accuracyRadius"]
				fields["SourceASN"] = location["asn"]
				fields["SourceASO"] = location["aso"]
				fields["SourceIsSatelliteProvider"] = location["isSatelliteProvider"]
				fields["SourceIsAnonymousProxy"] = location["isAnonymousProxy"]
			}
			// Try to geolocate DetinationIP if exists
			if fields["DestinationIP"] != "" {
				location := geo.Geolocate(fields["DestinationIP"])
				fields["DestinationCountry"] = location["country"]
				fields["DestinationCountryCode"] = location["countryCode"]
				fields["DestinationCity"] = location["city"]
				fields["DestinationLat"] = location["latitude"]
				fields["DestinationLon"] = location["longitude"]
				fields["DestinationAccuracyRadius"] = location["accuracyRadius"]
				fields["DestinationASN"] = location["asn"]
				fields["DestinationASO"] = location["aso"]
				fields["DestinationIsSatelliteProvider"] = location["isSatelliteProvider"]
				fields["DestinationIsAnonymousProxy"] = location["isAnonymousProxy"]
			}

			// Alert in the last step or save data to next cicle
			if steps-1 == step {
				// Use content of AlertName as Name if exists
				var alertName string
				if fields["AlertName"] != "" {
					alertName = fields["AlertName"]
				} else {
					alertName = rule.Name
				}
				// Use content of AlertCategory as Category if exists
				var alertCategory string
				if fields["AlertCategory"] != "" {
					alertCategory = fields["AlertCategory"]
				} else {
					alertCategory = rule.Category
				}

				// Run alert function
				Alert(
					alertName,
					rule.Severity,
					rule.Description,
					rule.Solution,
					alertCategory,
					rule.Tactic,
					rule.Reference,
					gjson.Get(l, "dataType").String(),
					gjson.Get(l, "dataSource").String(),
					fields,
				)
			} else {
				tmpLogs[step] = append(tmpLogs[step], fields)
			}
		}
	}
}
