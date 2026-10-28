package main

import "github.com/utmstack/utmstack/backend/pkg/http/middleware"

// internalKeyRoutes are the only routes that honour the internal key: what the
// other services call on the backend. A service added to the stack that talks to
// the backend with the key has to be listed here; until then its calls are
// refused with 401 and logged as "internal key presented where it is not
// accepted".
var internalKeyRoutes = []string{
	"/api/v1/mcp",                  // soc-ai's tool surface
	"/api/v1/soc-ai/quota/consume", // soc-ai plugin
	"/api/v1/soar/rule-executions", // soar plugin
	"/api/v1/internal",             // alert tag rules for the alerts plugin
	"/api/v1/ad-audit",             // ad-audit plugin
	"/api/v1/incidents",            // feeds plugin
	"/api/v1/incident-alerts",      // feeds plugin
	"/api/v1/notifications",        // rule-flood-guard
	"/api/v1/eventprocessing/correlation-rule/search-by-filters",   // rule-flood-guard
	"/api/v1/eventprocessing/correlation-rule/activate-deactivate", // rule-flood-guard
	"/api/v1/datasources/ping",                                     // agent-manager
	"/api/v1/datasources/usage-report",                             // installer's usage job
	"/api/v1/api-keys/authenticate",                                // log-input
	"/api/v1/branding/seed",                                        // installer
}

func internalKeyPolicy(cidrs string) (middleware.InternalPolicy, error) {
	nets, err := middleware.ParseCIDRs(cidrs)
	if err != nil {
		return middleware.InternalPolicy{}, err
	}
	return middleware.InternalPolicy{Paths: internalKeyRoutes, CIDRs: nets}, nil
}
