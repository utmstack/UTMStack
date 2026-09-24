package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

// checkWidgetSpec refuses a widget whose question the event store cannot
// answer. A widget saved with an unknown field is not an empty chart: the store
// rejects the query, so it shows an error for as long as it exists. The UI
// editor previews the question before saving; a client without a preview — the
// SOC-AI agent — gets this check instead.
func (m *Module) checkWidgetSpec(ctx context.Context, specJSON string) error {
	var spec domain.Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		return fmt.Errorf("invalid spec JSON: %s", err.Error())
	}
	if err := spec.Validate(); err != nil {
		return err
	}

	events := m.deps.Events
	if events == nil {
		return errors.New("the event store is not available, so this widget's query cannot be checked and it was not saved")
	}

	// A short window is enough to prove the question is well formed.
	to := time.Now()
	from := to.Add(-time.Hour)
	probe := spec
	probe.From, probe.To = &from, &to
	if _, err := usecase.NewQueryService(events).Run(ctx, probe); err != nil {
		return explainSpecFailure(ctx, events, spec, err)
	}
	return nil
}

// explainSpecFailure turns "the query could not run" into what the author has
// to change, when the cause is a field name the dataset does not have.
func explainSpecFailure(ctx context.Context, events *eventstore.Store, spec domain.Spec, cause error) error {
	if scope, err := storeScope(ctx, spec.Dataset, ""); err == nil {
		if fields, err := events.DescribeFields(ctx, scope); err == nil {
			names := make([]string, 0, len(fields))
			for _, f := range fields {
				names = append(names, f.Name)
			}
			if msg := unknownFieldMessage(spec, names); msg != "" {
				return errors.New(msg + " The widget was not saved.")
			}
		}
	}
	return fmt.Errorf("the event store could not run this widget's query (%s), so it was not saved — "+
		"check the field names with store.dataset.fields and try the spec with visualizations.query", cause)
}

// unknownFieldMessage names the first field of the spec that the dataset does
// not have, with the closest real one when there is an obvious match.
func unknownFieldMessage(spec domain.Spec, fields []string) string {
	known := make(map[string]bool, len(fields))
	for _, f := range fields {
		known[f] = true
	}
	check := func(role, field string) string {
		if field == "" || known[field] {
			return ""
		}
		msg := fmt.Sprintf("%s %q is not a field of the %s dataset", role, field, spec.Dataset)
		if near := closestField(field, fields); near != "" {
			msg += fmt.Sprintf(" — did you mean %q?", near)
		}
		return msg + fmt.Sprintf(" Call store.dataset.fields with dataset %q for the valid names.", spec.Dataset)
	}

	if msg := check("dimension", spec.Dimension); msg != "" {
		return msg
	}
	for _, c := range spec.Columns {
		if msg := check("column", c); msg != "" {
			return msg
		}
	}
	for _, f := range spec.Filters {
		if msg := check("filter field", f.Field); msg != "" {
			return msg
		}
	}
	return ""
}

var fieldNormalizer = strings.NewReplacer("_", "", "-", "", ".", "")

// closestField finds the real field a guessed name most likely meant: the same
// letters ignoring case and separators (data_source → dataSource).
func closestField(guess string, fields []string) string {
	want := strings.ToLower(fieldNormalizer.Replace(guess))
	for _, f := range fields {
		if strings.ToLower(fieldNormalizer.Replace(f)) == want {
			return f
		}
	}
	return ""
}
