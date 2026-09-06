package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	incidentsdomain "github.com/utmstack/utmstack/backend/modules/incidents/domain"
	incidentsdto "github.com/utmstack/utmstack/backend/modules/incidents/dto"
	soardomain "github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// IncidentOpener is the narrow slice of the incidents usecase that the SOAR
// incident executor consumes. Keeps this package free of the broader incidents
// import and lets tests swap in a fake.
type IncidentOpener interface {
	Create(ctx context.Context, userEmail string, req incidentsdto.CreateIncidentRequest) (*incidentsdomain.Incident, error)
}

// Incident opens an incident and links the alert that fired the flow. Params
// are just name + description — alert identity (id/name/severity) comes from
// the exec's built-in AlertID and the context bag populated by the dispatcher.
// ponytail: reuses incidents.CreateIncidentRequest verbatim — no shadow DTO.
type Incident struct{ client IncidentOpener }

func NewIncident(c IncidentOpener) *Incident { return &Incident{client: c} }

func (Incident) Type() string { return "incident" }

type incidentParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (i *Incident) Execute(ctx context.Context, exec *soardomain.SoarExecution) (json.RawMessage, error) {
	if i.client == nil {
		return nil, errors.New("soar incident: client not configured")
	}
	var p incidentParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar incident: params: %w", err)
		}
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, errors.New("soar incident: name is required")
	}
	if strings.TrimSpace(exec.AlertID) == "" {
		return nil, errors.New("soar incident: no alert linked to this execution")
	}

	src := string(exec.Context)
	if src == "" {
		src = "{}"
	}
	alertName := gjson.Get(src, "alert.name").String()
	if alertName == "" {
		alertName = exec.AlertID
	}
	severity := gjson.Get(src, "alert.severity").String()
	if severity == "" {
		severity = "Low"
	}

	req := incidentsdto.CreateIncidentRequest{
		IncidentName: name,
		AlertList: []incidentsdto.AlertLinkItem{{
			AlertID:       exec.AlertID,
			AlertName:     alertName,
			AlertSeverity: severity,
		}},
	}
	if desc := strings.TrimSpace(p.Description); desc != "" {
		req.IncidentDescription = &desc
	}

	inc, err := i.client.Create(ctx, "", req)
	if err != nil {
		return nil, fmt.Errorf("soar incident: create: %w", err)
	}
	exec.Result = fmt.Sprintf("opened incident %s: %s", inc.ID.String(), inc.Name)
	return nil, nil
}
