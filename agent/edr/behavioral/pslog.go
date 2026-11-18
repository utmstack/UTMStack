package behavioral

import (
	"encoding/xml"
	"strings"
)

type ps4104 struct {
	recordID   int64
	scriptText string
}

// parse4104 extracts EventRecordID and the ScriptBlockText EventData field from
// wevtutil XML output (one or more concatenated <Event> blocks). Pure so it is
// unit-testable on any OS.
func parse4104(xmlOut string) []ps4104 {
	var out []ps4104
	dec := xml.NewDecoder(strings.NewReader("<root>" + xmlOut + "</root>"))
	type evtData struct {
		Data []struct {
			Name string `xml:"Name,attr"`
			Val  string `xml:",chardata"`
		} `xml:"EventData>Data"`
		RecordID int64 `xml:"System>EventRecordID"`
	}
	var root struct {
		Events []evtData `xml:"Event"`
	}
	if err := dec.Decode(&root); err != nil {
		return out
	}
	for _, e := range root.Events {
		var script string
		for _, d := range e.Data {
			if d.Name == "ScriptBlockText" {
				script = d.Val
			}
		}
		out = append(out, ps4104{recordID: e.RecordID, scriptText: script})
	}
	return out
}
