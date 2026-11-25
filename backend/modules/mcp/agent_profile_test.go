package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// What the agent sees may name only tools that exist. A rename in a tools_*.go file would otherwise
// drop the tool from the agent without anyone noticing: the list would still compile.
func TestEveryAgentToolIsARegisteredTool(t *testing.T) {
	registered := map[string]bool{}
	for _, n := range registeredToolNames(t) {
		registered[n] = true
	}

	var missing []string
	for name := range agentTools {
		if !registered[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("the agent profile names tools that are not registered:\n  %s", strings.Join(missing, "\n  "))
	}
}

// The reason the list exists. The internal key passes every gate, so anything in here is something the
// agent can do to the whole installation.
func TestTheAgentCannotReachWhatManagesPeopleMoneyOrSettings(t *testing.T) {
	forbiddenPrefix := []string{"iam.", "tenants.", "billing.", "branding.", "config.", "audit.", "platform.",
		"socai.", "integrations.", "notifications."}
	forbiddenName := []string{"store.search_sql", "store.search_csv"}

	for name := range agentTools {
		for _, p := range forbiddenPrefix {
			if strings.HasPrefix(name, p) {
				t.Errorf("%s: the agent must not reach %s tools", name, strings.TrimSuffix(p, "."))
			}
		}

		for _, f := range forbiddenName {
			if name == f {
				t.Errorf("%s: raw queries and exports are not for the agent", name)
			}
		}

		if strings.HasSuffix(name, ".delete") || strings.Contains(name, ".bulk_") || strings.Contains(name, "terminate") {
			t.Errorf("%s: the agent does not delete or act in bulk", name)
		}
	}
}

// A list that only grows is how the agent got a hundred tools in the first place.
func TestTheAgentListStaysShort(t *testing.T) {
	core := 0

	for _, g := range agentTools {
		if g == groupCore {
			core++
		}
	}

	if n := len(agentTools); n > 45 {
		t.Errorf("the agent can reach %d tools; a model that is given this many stops reaching answers. Justify the new "+
			"one and drop another, or raise this limit on purpose", n)
	}

	if core > 12 {
		t.Errorf("the core is %d tools and every request has it; keep it to what investigating an alert takes", core)
	}
}

func TestEveryToolOfTheAgentBelongsToAGroupACallerCanAskFor(t *testing.T) {
	for name, g := range agentTools {
		if g != groupCore && !knownGroups[g] {
			t.Errorf("%s is in group %q, which nobody can turn on", name, g)
		}
	}
}

func tools(t *testing.T, s *mcp.Server) []string {
	t.Helper()

	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, tl := range res.Tools {
		names = append(names, tl.Name)
	}
	sort.Strings(names)

	return names
}

func testModule() *Module {
	m := &Module{
		deps:   &Deps{},
		server: mcp.NewServer(&mcp.Implementation{Name: "full", Version: "1"}, nil),
		agent:  mcp.NewServer(&mcp.Implementation{Name: "agent", Version: "1"}, nil),
	}
	m.agent.AddReceivingMiddleware(agentFilter)

	noop := func(context.Context, *authz.Actor, struct{}) (any, error) { return map[string]string{"ok": "1"}, nil }

	for _, name := range []string{"store.search", "alerts.update_status", "soar.rule.create", "tenants.terminate", "iam.user.list"} {
		Add(m, &mcp.Tool{Name: name, Description: name}, Gate{}, noop)
	}

	return m
}

func TestATooledAddedToTheModuleReachesTheAgentOnlyWhenItIsOnTheList(t *testing.T) {
	m := testModule()

	if got := strings.Join(tools(t, m.server), ","); got != "alerts.update_status,iam.user.list,soar.rule.create,store.search,tenants.terminate" {
		t.Errorf("the full server lost tools: %s", got)
	}

	// Without a request that turned groups on, the agent server lists only what is in the core.
	if got := strings.Join(tools(t, m.agent), ","); got != "store.search" {
		t.Errorf("the agent server shows %q, want only the core tool store.search", got)
	}

	if m.ToolCount() != 5 || m.AgentToolCount() != 3 {
		t.Errorf("counts: %d full, %d agent", m.ToolCount(), m.AgentToolCount())
	}
}

// headers adds the same headers to every request, standing in for the gin middleware that lifts them.
type headers struct{ h map[string]string }

func (r headers) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range r.h {
		req.Header.Set(k, v)
	}

	return http.DefaultTransport.RoundTrip(req)
}

// serve runs the module's real routing over HTTP. The wrapper plays the part of authentication and
// liftAgentGroups: who the caller is comes from X-Test-Caller, and for the internal one the groups come
// from the header, exactly as in production.
func serve(t *testing.T, m *Module) *httptest.Server {
	t.Helper()

	h := mcp.NewStreamableHTTPHandler(m.serverFor, &mcp.StreamableHTTPOptions{DisableLocalhostProtection: true, Stateless: true})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Header.Get("X-Test-Caller") == "internal" {
			ctx = WithActor(ctx, &authz.Actor{Internal: true})
			ctx = WithAgentGroups(ctx, r.Header.Get(agentHeader))
		} else {
			ctx = WithActor(ctx, &authz.Actor{})
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	}))
	t.Cleanup(srv.Close)

	return srv
}

func connect(t *testing.T, srv *httptest.Server, h map[string]string) *mcp.ClientSession {
	t.Helper()

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "t", Version: "1"}, nil).Connect(context.Background(),
		&mcp.StreamableClientTransport{Endpoint: srv.URL, HTTPClient: &http.Client{Transport: headers{h}}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = cs.Close() })

	return cs
}

func listed(t *testing.T, cs *mcp.ClientSession) string {
	t.Helper()

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, tl := range res.Tools {
		names = append(names, tl.Name)
	}
	sort.Strings(names)

	return strings.Join(names, ",")
}

func TestTheAgentSeesOnlyTheCoreAndTheGroupsItAskedFor(t *testing.T) {
	srv := serve(t, testModule())
	internal := map[string]string{"X-Test-Caller": "internal"}

	if got := listed(t, connect(t, srv, internal)); got != "store.search" {
		t.Errorf("no groups asked for: %q", got)
	}

	if got := listed(t, connect(t, srv, map[string]string{"X-Test-Caller": "internal", agentHeader: "soar"})); got != "soar.rule.create,store.search" {
		t.Errorf("soar asked for: %q", got)
	}

	both := map[string]string{"X-Test-Caller": "internal", agentHeader: " alert_actions , soar ,nonsense,iam"}
	if got := listed(t, connect(t, srv, both)); got != "alerts.update_status,soar.rule.create,store.search" {
		t.Errorf("two groups, plus names that mean nothing: %q", got)
	}
}

func TestAToolTheRequestDidNotTurnOnCannotBeCalledEither(t *testing.T) {
	srv := serve(t, testModule())
	ctx := context.Background()

	cs := connect(t, srv, map[string]string{"X-Test-Caller": "internal"})

	for _, name := range []string{"soar.rule.create", "tenants.terminate", "iam.user.list"} {
		if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: map[string]any{}}); err == nil {
			t.Errorf("%s answered although the agent was never shown it", name)
		}
	}

	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "store.search", Arguments: map[string]any{}}); err != nil {
		t.Errorf("a core tool must answer: %v", err)
	}

	on := connect(t, srv, map[string]string{"X-Test-Caller": "internal", agentHeader: "soar"})
	if _, err := on.CallTool(ctx, &mcp.CallToolParams{Name: "soar.rule.create", Arguments: map[string]any{}}); err != nil {
		t.Errorf("a tool of a group that was turned on must answer: %v", err)
	}
}

func TestAPersonKeepsTheFullServerWhateverTheyAsk(t *testing.T) {
	srv := serve(t, testModule())
	person := connect(t, srv, map[string]string{"X-Test-Caller": "person", agentHeader: "soar"})

	if got := listed(t, person); got != "alerts.update_status,iam.user.list,soar.rule.create,store.search,tenants.terminate" {
		t.Errorf("a person must see everything the full server has: %q", got)
	}
}

func TestTheCallerDecidesTheServerNotTheRequest(t *testing.T) {
	full, agent := mcp.NewServer(&mcp.Implementation{Name: "f", Version: "1"}, nil),
		mcp.NewServer(&mcp.Implementation{Name: "a", Version: "1"}, nil)
	m := &Module{server: full, agent: agent}

	req := func(actor *authz.Actor, header string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		if header != "" {
			r.Header.Set("X-MCP-Profile", header)
		}
		if actor != nil {
			r = r.WithContext(WithActor(r.Context(), actor))
		}

		return r
	}

	if m.serverFor(req(&authz.Actor{Internal: true}, "")) != agent {
		t.Error("the internal caller must get the agent's list")
	}

	if m.serverFor(req(&authz.Actor{Internal: true}, "full")) != agent {
		t.Error("an internal caller cannot ask for the full list")
	}

	if m.serverFor(req(&authz.Actor{}, "agent")) != full || m.serverFor(req(nil, "")) != full {
		t.Error("a person, or nobody, gets the full server")
	}

	if (&Module{server: full}).serverFor(req(&authz.Actor{Internal: true}, "")) != full {
		t.Error("without an agent server everything falls back to the full one")
	}
}

// A client rejects a result whose structuredContent is not an object, and a list is what the tools that
// rank values or list rows return. One such tool took the agent's whole chat down with it.
func TestAToolThatReturnsAListStillAnswersWithAnObject(t *testing.T) {
	m := testModule()

	Add(m, &mcp.Tool{Name: "store.property_values", Description: "values"}, Gate{},
		func(context.Context, *authz.Actor, struct{}) (any, error) {
			return []map[string]any{{"value": "high", "count": 3}}, nil
		})

	srv := serve(t, m)
	cs := connect(t, srv, map[string]string{"X-Test-Caller": "internal"})

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "store.property_values", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}

	obj, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structuredContent is %T, want an object", res.StructuredContent)
	}

	rows, ok := obj["result"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("the list must sit under \"result\": %v", obj)
	}
}

func TestAsObjectLeavesObjectsAlone(t *testing.T) {
	in := map[string]any{"total": 1}
	if got, _ := any(asObject[any](in)).(map[string]any); got["total"] != 1 || len(got) != 1 {
		t.Fatalf("an object must pass through untouched: %v", got)
	}

	if asObject[any](nil) != nil {
		t.Fatal("nil stays nil")
	}
}
