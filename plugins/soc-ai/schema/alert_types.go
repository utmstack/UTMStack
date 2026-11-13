package schema

import (
	"encoding/json"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type AlertFields struct {
	Timestamp         string         `json:"@timestamp"`
	Status            string         `json:"status,omitempty"`
	StatusObservation string         `json:"statusObservation,omitempty"`
	IsIncident        bool           `json:"isIncident,omitempty"`
	IncidentDetail    IncidentDetail `json:"incidentDetail,omitzero"`
	Solution          string         `json:"solution,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Notes             string         `json:"notes,omitempty"`
	TagRulesApplied   []string       `json:"tagRulesApplied,omitempty"`
	LastEvent         *plugins.Event `json:"lastEvent,omitempty"`
	AnonymizedFields  []string       `json:"anonymizedFields,omitempty"`
	plugins.Alert
}

func (a AlertFields) MarshalJSON() ([]byte, error) {
	type alias AlertFields
	raw, err := json.Marshal(alias(a))
	if err != nil {
		return nil, err
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if a.LastEvent != nil {
		event, err := utils.ProtoMessageToBytes(a.LastEvent)
		if err != nil {
			return nil, err
		}
		document["lastEvent"] = *event
	}
	if len(a.Events) > 0 {
		events := make([]json.RawMessage, 0, len(a.Events))
		for _, event := range a.Events {
			if event == nil {
				events = append(events, json.RawMessage("null"))
				continue
			}
			data, err := utils.ProtoMessageToBytes(event)
			if err != nil {
				return nil, err
			}
			events = append(events, *data)
		}
		document["events"], err = json.Marshal(events)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(document)
}

func (a *AlertFields) UnmarshalJSON(data []byte) error {
	type alias AlertFields
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var doc struct {
		LastEvent json.RawMessage   `json:"lastEvent"`
		Events    []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	if len(doc.LastEvent) > 0 && string(doc.LastEvent) != "null" {
		var event plugins.Event
		raw := string(doc.LastEvent)
		if err := utils.StringToProtoMessage(&raw, &event); err != nil {
			return err
		}
		decoded.LastEvent = &event
	}
	if doc.Events != nil {
		decoded.Events = make([]*plugins.Event, len(doc.Events))
		for i, data := range doc.Events {
			if string(data) == "null" {
				continue
			}
			var event plugins.Event
			raw := string(data)
			if err := utils.StringToProtoMessage(&raw, &event); err != nil {
				return err
			}
			decoded.Events[i] = &event
		}
	}
	*a = AlertFields(decoded)
	return nil
}

type AlertCorrelation struct {
	CurrentAlert  AlertFields
	RelatedAlerts []AlertFields
	Counts        MatchTypeCounts
}

type AlertCounts struct {
	Incidents     int
	FalsePositive int
	Standard      int
	Unclassified  int
}

type MatchTypeCounts struct {
	OriginIP   AlertCounts
	TargetIP   AlertCounts
	OriginUser AlertCounts
	TargetUser AlertCounts
}
