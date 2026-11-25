package socai

import (
	"context"
	"encoding/json"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/go-sdk/v2/store"

	"github.com/utmstack/utmstack/backend/modules/socai/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// maxMessage keeps a message well under NATS's default limit of a megabyte.
const maxMessage = 900 * 1024

type configReader interface {
	Get(ctx context.Context) (*dto.ConfigResponse, error)
}

// Docs is the part of the event store the dossier needs.
type Docs interface {
	FetchByID(ctx context.Context, s store.Scope, id string) (json.RawMessage, error)
	FetchN(ctx context.Context, s store.Scope, filters []store.Filter, n int) ([]json.RawMessage, error)
	Count(ctx context.Context, s store.Scope, filters []store.Filter) (int64, error)
}

// ScoreFunc runs the deterministic scoring of one alert.
type ScoreFunc func(ctx context.Context, alertID string) (any, error)

// Publisher puts a message where the SOC AI service reads it.
type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// Triage sends each new alert to the SOC AI service for assessment, with everything the service would
// otherwise ask the backend for, one request at a time, already worked out. Doing it here is a function
// call against the database; from the service it was a round trip for each piece.
type Triage struct {
	config configReader
	quota  *AIQuota
	score  ScoreFunc
	docs   Docs
	pub    Publisher
}

func NewTriage(config configReader, quota *AIQuota, score ScoreFunc, docs Docs, pub Publisher) *Triage {
	return &Triage{config: config, quota: quota, score: score, docs: docs, pub: pub}
}

// Message is what the service receives. Facts is the dossier; it is left out when it could not be built,
// and the service then asks for the pieces itself.
type Message struct {
	TenantID string         `json:"tenantId"`
	AlertID  string         `json:"alertId"`
	Facts    map[string]any `json:"facts,omitempty"`
}

// AlertRaised is registered as a hook of the alerts module. It returns an error only for what the
// operator needs to know: an alert that was due for triage and could not be queued.
func (t *Triage) AlertRaised(ctx context.Context, tenantID, alertID string) error {
	if t == nil || tenantID == "" || alertID == "" {
		return nil
	}

	ctx = authz.WithTenantID(ctx, tenantID)

	cfg, err := t.config.Get(ctx)
	if err != nil {
		return err
	}

	if !cfg.Configured || !cfg.AutoAnalyze {
		return nil
	}

	if v := t.quota.Check(ctx, tenantID); !v.Allowed {
		catcher.Info("soc-ai: this tenant has used the AI requests its plan allows today; the alert is not analysed",
			map[string]any{"tenant": tenantID, "limit": v.Limit, "alertId": alertID})

		return nil
	}

	msg := Message{TenantID: tenantID, AlertID: alertID}

	facts, err := t.dossier(ctx, tenantID, alertID)
	if err != nil {
		catcher.Warn("soc-ai: the alert's dossier could not be built; the service will ask for it", map[string]any{
			"alertId": alertID, "error": err.Error()})
	} else {
		msg.Facts = facts
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if len(data) > maxMessage {
		msg.Facts = nil
		if data, err = json.Marshal(msg); err != nil {
			return err
		}
	}

	pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return t.pub.Publish(pctx, AlertsSubject, data)
}
