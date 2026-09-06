package ingest

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	"github.com/utmstack/UTMStack/log-input/auth"
	"github.com/utmstack/UTMStack/log-input/config"
)

func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

const processName = "log-input"

type Publisher interface {
	Publish(ctx context.Context, l *plugins.Log) error
}

type Server struct {
	plugins.UnimplementedIntegrationServer

	cfg *config.Config
	pub Publisher

	grpc   *grpc.Server
	health *health.Server
}

func NewServer(cfg *config.Config, a *auth.Service, pub Publisher) (*Server, error) {
	tlsCfg, err := loadTLSConfig(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, catcher.Error("cannot read the certificate files", err, map[string]any{
			"process": processName,
			"cert":    cfg.CertFile,
		})
	}

	creds := credentials.NewTLS(tlsCfg)

	m := newMiddlewares(a)

	s := &Server{
		cfg: cfg,
		pub: pub,
		// Keepalive pings keep the collector's idle ProcessLog stream alive
		// through the frontend nginx grpc_read_timeout (900s): the PING ACKs we
		// answer reset the upstream read timer on every hop.
		grpc: grpc.NewServer(
			grpc.Creds(creds),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				Time:    30 * time.Second,
				Timeout: 10 * time.Second,
			}),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             10 * time.Second,
				PermitWithoutStream: true,
			}),
			grpc.ChainUnaryInterceptor(m.unary),
			grpc.ChainStreamInterceptor(m.stream),
		),
		health: health.NewServer(),
	}

	plugins.RegisterIntegrationServer(s.grpc, s)
	grpcHealth.RegisterHealthServer(s.grpc, s.health)
	s.health.SetServingStatus("", grpcHealth.HealthCheckResponse_SERVING)

	return s, nil
}

func (s *Server) Serve() error {
	listener, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return catcher.Error("cannot listen", err, map[string]any{
			"process": processName,
			"addr":    s.cfg.ListenAddr,
		})
	}

	catcher.Info("log input listening", map[string]any{
		"process": processName,
		"addr":    s.cfg.ListenAddr,
	})

	return s.grpc.Serve(listener)
}

// Stop drains: a log accepted but not yet acked is one the sender still holds,
// and cutting the connection under it would have it sent twice.
func (s *Server) Stop() {
	s.health.SetServingStatus("", grpcHealth.HealthCheckResponse_NOT_SERVING)
	s.grpc.GracefulStop()
}

// ProcessLog acks only after the publish is confirmed, so the ack means
// durable rather than received.
func (s *Server) ProcessLog(srv plugins.Integration_ProcessLogServer) error {
	ctx := srv.Context()

	for {
		l, err := srv.Recv()
		if err != nil {
			return err
		}

		applyDefaults(ctx, s.cfg, l)

		if err := s.pub.Publish(ctx, l); err != nil {
			return catcher.Error("cannot publish the log", err, map[string]any{
				"process":    processName,
				"id":         l.Id,
				"dataSource": l.DataSource,
			})
		}

		if err := srv.Send(&plugins.Ack{LastId: l.Id}); err != nil {
			return catcher.Error("cannot send the ack", err, map[string]any{
				"process": processName,
				"lastId":  l.Id,
			})
		}
	}
}

func applyDefaults(ctx context.Context, cfg *config.Config, l *plugins.Log) {
	if l.Id == "" {
		l.Id = uuid.NewString()
	}
	if tenant := tenantFrom(ctx); tenant != "" {
		l.TenantId = tenant
	}
	if l.TenantId == "" {
		l.TenantId = cfg.DefaultTenant
	}
	if l.DataType == "" {
		l.DataType = "generic"
	}
	if l.DataSource == "" {
		l.DataSource = "unknown"
	}
	if l.Timestamp == "" {
		l.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
}
