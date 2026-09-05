package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	soardomain "github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// Notifier is the narrow slice of the notifications usecase that the SOAR
// notify executor consumes. Keeps the soar/executor package free of a broader
// notifications import and lets tests swap in a fake.
type Notifier interface {
	Notify(ctx context.Context, source domain.NotificationSource, ntype domain.NotificationType, message string) error
}

// Notify posts an entry into the in-app notification stream — visible in the
// UI's bell menu. Meant for kind=executor: the branch outcome tracks whether
// the notification was accepted.
type Notify struct {
	client Notifier
}

func NewNotify(c Notifier) *Notify { return &Notify{client: c} }

func (Notify) Type() string { return "notify" }

type notifyParams struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

func (n *Notify) Execute(ctx context.Context, exec *soardomain.SoarExecution) (json.RawMessage, error) {
	if n.client == nil {
		return nil, errors.New("soar notify: client not configured")
	}
	var p notifyParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar notify: params: %w", err)
		}
	}
	if p.Message == "" {
		return nil, errors.New("soar notify: message is required")
	}
	ntype := domain.TypeInfo
	switch domain.NotificationType(p.Type) {
	case domain.TypeWarning:
		ntype = domain.TypeWarning
	case domain.TypeError:
		ntype = domain.TypeError
	}
	if err := n.client.Notify(ctx, domain.SourceSystem, ntype, p.Message); err != nil {
		return nil, err
	}
	exec.Result = fmt.Sprintf("notified (%s): %s", ntype, p.Message)
	return nil, nil
}
