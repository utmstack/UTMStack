package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	notificationdomain "github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// CommandSummary fills `e.Command` for flow-origin node executions whose action
// does not live in the command column — http, mail, llm, notify, incident and
// conditional nodes store theirs in Params. It derives a short human-readable
// line so the command column always shows what the node was configured to do;
// the frontend clamps long lines with a tooltip for the full text.
//
// Params were already interpolated by the dispatcher, so a secret the user
// inlined in the node config appears here in plain text. `vars` (may be nil)
// re-applies secret masking to the derived line. Manual executions and rows
// that already carry a command (shell nodes) are left untouched.
func CommandSummary(ctx context.Context, vars connectors.VariableUsecase, e *domain.SoarExecution) {
	if e.Origin != domain.ExecutionOriginFlow || strings.TrimSpace(e.Command) != "" {
		return
	}
	summary := summarizeNodeAction(e.Executor, e.Params)
	if summary == "" {
		return
	}
	if vars != nil {
		if masked, err := vars.MaskSecrets(ctx, summary); err == nil {
			summary = masked
		}
	}
	e.Command = summary
}

// summarizeNodeAction renders one configured action as a single line.
func summarizeNodeAction(executor string, params json.RawMessage) string {
	if len(params) == 0 {
		return ""
	}
	g := gjson.ParseBytes(params)
	switch executor {
	case "http":
		method := strings.ToUpper(strings.TrimSpace(g.Get("method").Str))
		if method == "" {
			if g.Get("body").Exists() {
				method = http.MethodPost
			} else {
				method = http.MethodGet
			}
		}
		return method + " " + g.Get("url").Str
	case "mail":
		s := "mail to " + strings.TrimSpace(g.Get("to").Str)
		if subj := strings.TrimSpace(g.Get("subject").Str); subj != "" {
			s += " — " + subj
		}
		return s
	case "llm_enrich", "llm_action":
		if p := strings.TrimSpace(g.Get("prompt").Str); p != "" {
			return "LLM: " + p
		}
		return ""
	case "notify":
		label := "notify (INFO)"
		if g.Get("type").Str == string(notificationdomain.TypeWarning) {
			label = "notify (WARNING)"
		}
		return label + ": " + g.Get("message").Str
	case "incident":
		return "open incident: " + g.Get("name").Str
	case "conditional":
		conds := g.Get("conditions").Array()
		if len(conds) == 0 {
			return ""
		}
		first := summarizeCondition(conds[0].Value())
		if len(conds) == 1 {
			return "if: " + first
		}
		return fmt.Sprintf("if: %s AND %d more", first, len(conds)-1)
	}
	return ""
}

// summarizeCondition renders one conditional predicate, e.g. `severity IS High`.
func summarizeCondition(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	c := gjson.ParseBytes(b)
	return strings.TrimSpace(c.Get("field").Str + " " + c.Get("operator").Str + " " + c.Get("value").String())
}
