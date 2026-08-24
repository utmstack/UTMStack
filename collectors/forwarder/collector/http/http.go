package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	stdhttp "net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/configwatcher"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/schema"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

type HTTPCollector struct {
	instances map[string]*httpInstance // keyed by "DataType:proto" e.g. "my-fw:http"
	mu        sync.RWMutex
	queue     chan *plugins.Log
	ctx       context.Context
}

type httpConfig struct {
	dataType        string
	proto           string // "http" | "https"
	port            string
	path            string // URL path, default "/logs"
	bind            string // "127.0.0.1" default
	auth            string // "" | "bearer" | "hmac"
	signatureHeader string // e.g. "X-Hub-Signature-256"
	tokenPath       string // path to token file (used by bearer and hmac)
}

type httpInstance struct {
	httpConfig

	server   *stdhttp.Server
	listener net.Listener
}

func New() *HTTPCollector {
	return &HTTPCollector{
		instances: make(map[string]*httpInstance),
	}
}

func (h *HTTPCollector) Name() string { return "http" }

func (h *HTTPCollector) Start(ctx context.Context, queue chan *plugins.Log) {
	h.ctx = ctx
	h.queue = queue
	configwatcher.Watch(ctx, "http collector", h.reconcile)
}

func (h *HTTPCollector) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, inst := range h.instances {
		shutdownInstance(inst)
		delete(h.instances, key)
	}
}

func (h *HTTPCollector) reconcile() {
	cfg, err := schema.ReadCollectorConfig()
	if err != nil {
		utils.Logger.ErrorF("http: error reading collector config: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	desired := map[string]*httpInstance{}
	for name, integration := range cfg.Integrations {
		if integration.HTTP != nil && integration.HTTP.Enabled {
			key := name + ":http"
			inst, err := buildInstance(name, "http", integration.HTTP)
			if err != nil {
				utils.Logger.ErrorF("http: skipping integration %s: %v", key, err)
			} else {
				desired[key] = inst
			}
		}
		if integration.HTTPS != nil && integration.HTTPS.Enabled {
			key := name + ":https"
			inst, err := buildInstance(name, "https", integration.HTTPS)
			if err != nil {
				utils.Logger.ErrorF("http: skipping integration %s: %v", key, err)
			} else {
				desired[key] = inst
			}
		}
	}

	// Everything that must change goes down before anything comes up.
	// Shutting down and restarting one key at a time cannot handle two
	// integrations swapping ports: whichever is processed first fails to bind
	// because the other still holds the port, and stays down until the next
	// reconcile.
	for key, running := range h.instances {
		inst, wanted := desired[key]
		// Compare the whole config: a presence check silently ignores port,
		// path and auth edits. New fields must go in httpConfig, not here.
		if wanted && running.httpConfig == inst.httpConfig {
			continue
		}
		if wanted {
			utils.Logger.Info("http[%s]: configuration changed, restarting on %s:%s%s (auth=%q)",
				key, inst.bind, inst.port, inst.path, inst.auth)
		}
		shutdownInstance(running)
		delete(h.instances, key)
	}

	for key, inst := range desired {
		if _, isRunning := h.instances[key]; isRunning {
			continue
		}
		if err := h.startInstance(inst); err != nil {
			utils.Logger.ErrorF("http: failed to start %s: %v", key, err)
			continue
		}
		h.instances[key] = inst
	}
}

func (h *HTTPCollector) startInstance(inst *httpInstance) error {
	addr := inst.bind + ":" + inst.port

	mux := stdhttp.NewServeMux()
	mux.Handle(inst.path, h.newHandler(inst))

	srv := &stdhttp.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 14, // 16 KB headers
	}

	var ln net.Listener
	var err error

	if inst.proto == "https" {
		tlsCfg, tlsErr := utils.LoadIntegrationTLSConfig(config.IntegrationCertPath, config.IntegrationKeyPath)
		if tlsErr != nil {
			return fmt.Errorf("https: load TLS: %w", tlsErr)
		}
		ln, err = tls.Listen("tcp", addr, tlsCfg)
	} else {
		ln, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	inst.server = srv
	inst.listener = ln

	go func() {
		if serveErr := srv.Serve(ln); serveErr != nil && serveErr != stdhttp.ErrServerClosed {
			utils.Logger.ErrorF("http[%s:%s]: serve error: %v", inst.dataType, inst.proto, serveErr)
		}
	}()

	utils.Logger.Info("http[%s]: listening on %s", inst.dataType, addr)
	return nil
}

func shutdownInstance(inst *httpInstance) {
	if inst.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := inst.server.Shutdown(ctx); err != nil {
			utils.Logger.ErrorF("http[%s:%s]: shutdown error: %v", inst.dataType, inst.proto, err)
			if err := inst.server.Close(); err != nil {
				utils.Logger.ErrorF("http[%s:%s]: force close error: %v", inst.dataType, inst.proto, err)
			}
		}
	}
}

// defaultPath is used when the operator leaves the path empty.
const defaultPath = "/logs"

func normalizeAndValidatePath(path string) (normalized string, repaired bool, err error) {
	if path == "" {
		return defaultPath, false, nil
	}

	normalized = strings.TrimSpace(path)
	if normalized == "" {
		normalized = defaultPath
	}

	if i := strings.IndexAny(normalized, " \t"); i >= 0 {
		return "", false, fmt.Errorf("invalid path %q: contains whitespace at offset %d; "+
			"a URL path cannot contain a space or tab, and method-scoped patterns such as %q are not supported",
			path, i, "GET /logs")
	}

	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	if err := probeMuxPattern(normalized); err != nil {
		return "", false, fmt.Errorf("invalid path %q: %w", path, err)
	}

	return normalized, normalized != path, nil
}

func probeMuxPattern(pattern string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("rejected by http.ServeMux: %v", r)
		}
	}()
	stdhttp.NewServeMux().Handle(pattern, stdhttp.NotFoundHandler())
	return nil
}

func buildInstance(name, proto string, port *schema.HTTPPort) (*httpInstance, error) {
	bind := port.Bind
	if bind == "" {
		bind = "127.0.0.1"
	}
	path, repaired, err := normalizeAndValidatePath(port.Path)
	if err != nil {
		return nil, err
	}
	if repaired {
		utils.Logger.Info("http[%s:%s]: path %q is malformed and was read as %q; fix the integration configuration",
			name, proto, port.Path, path)
	}
	sigHeader := port.SignatureHeader
	if sigHeader == "" {
		sigHeader = "X-Hub-Signature-256" // sensible default
	}

	return &httpInstance{
		httpConfig: httpConfig{
			dataType:        name,
			proto:           proto,
			port:            port.Port,
			path:            path,
			bind:            bind,
			auth:            port.Auth,
			signatureHeader: sigHeader,
			tokenPath:       tokenPath(name),
		},
	}, nil
}

func tokenPath(name string) string {
	return filepath.Join(config.HTTPTokenDir, "integration-http-"+name+".token")
}
