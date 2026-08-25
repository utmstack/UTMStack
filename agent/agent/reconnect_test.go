package agent

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/utmstack/UTMStack/agent/config"
)

// rstStream is what nginx sends an idle bidirectional stream when it decides the
// request body has stopped arriving. gRPC surfaces it as codes.Internal, which
// used to be read as "not worth reconnecting for".
var rstStream = status.Error(codes.Internal,
	"stream terminated by RST_STREAM with error code: PROTOCOL_ERROR")

type agentStream = grpc.BidiStreamingServer[BidirectionalStream, BidirectionalStream]

type testAgentService struct {
	UnimplementedAgentServiceServer
	streams atomic.Int32
	handle  func(agentStream) error
}

func (s *testAgentService) AgentStream(srv agentStream) error {
	s.streams.Add(1)
	return s.handle(srv)
}

// breakImmediately is the failure this whole file is about: the server hangs up
// the way a proxy does, before the stream has carried anything.
func breakImmediately(agentStream) error { return rstStream }

// startTestServer runs a TLS gRPC server on a loopback port and points the
// agent's connection settings at it. Everything it changes is restored.
func startTestServer(t *testing.T, handle func(agentStream) error) *testAgentService {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	svc := &testAgentService{handle: handle}
	srv := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{selfSignedCert(t)},
	})))
	RegisterAgentServiceServer(srv, svc)
	go func() { _ = srv.Serve(lis) }()

	_, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}

	prevPort, prevSleep, prevBeat := config.AgentManagerPort, timeToSleep, heartbeatInterval
	config.AgentManagerPort = port
	timeToSleep = 20 * time.Millisecond
	heartbeatInterval = 20 * time.Millisecond

	t.Cleanup(func() {
		srv.Stop()
		config.AgentManagerPort, timeToSleep, heartbeatInterval = prevPort, prevSleep, prevBeat
		agentManagerEntry.mu.Lock()
		if agentManagerEntry.conn != nil {
			_ = agentManagerEntry.conn.Close()
			agentManagerEntry.conn = nil
		}
		agentManagerEntry.mu.Unlock()
	})

	return svc
}

func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// A stream that dies with RST_STREAM used to leave the agent calling Recv on a
// corpse: the error was not Unavailable or Canceled, so the loop "retried" a
// stream gRPC had already finished, and every later command was answered with
// "agent not found or is disconnected" until someone restarted the service.
func TestAgentStreamReconnectsAfterRSTStream(t *testing.T) {
	svc := startTestServer(t, breakImmediately)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go IncidentResponseStream(&config.Config{
		Server:             "127.0.0.1",
		SkipCertValidation: true,
	}, ctx)

	deadline := time.After(10 * time.Second)
	for svc.streams.Load() < 3 {
		select {
		case <-deadline:
			t.Fatalf("the agent opened %d streams; a broken stream is not being reconnected",
				svc.streams.Load())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// Reconnecting forever is only correct while someone still wants the agent
// talking. The loop used to ignore ctx entirely, so a shutdown sat through its
// full 30s timeout and logged "some goroutines may not have stopped".
func TestAgentStreamStopsWhenContextIsCancelled(t *testing.T) {
	startTestServer(t, breakImmediately)

	ctx, cancel := context.WithCancel(t.Context())
	returned := make(chan struct{})

	go func() {
		IncidentResponseStream(&config.Config{
			Server:             "127.0.0.1",
			SkipCertValidation: true,
		}, ctx)
		close(returned)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-returned:
	case <-time.After(5 * time.Second):
		t.Fatal("IncidentResponseStream kept reconnecting after its context was cancelled")
	}
}

// Commands arrive rarely, so without this the stream carries nothing between
// them and a proxy ends it as an abandoned request body. The heartbeat is the
// traffic that keeps it open.
func TestAgentStreamSendsHeartbeats(t *testing.T) {
	beats := make(chan struct{}, 16)

	startTestServer(t, func(srv agentStream) error {
		for {
			in, err := srv.Recv()
			if err != nil {
				return err
			}
			if _, ok := in.StreamMessage.(*BidirectionalStream_Heartbeat); ok {
				select {
				case beats <- struct{}{}:
				default:
				}
			}
		}
	})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go IncidentResponseStream(&config.Config{
		Server:             "127.0.0.1",
		SkipCertValidation: true,
	}, ctx)

	// Two, not one: the first proves it starts beating, the second that it
	// keeps beating on the same stream rather than only on connect.
	for i := range 2 {
		select {
		case <-beats:
		case <-time.After(10 * time.Second):
			t.Fatalf("the server received %d heartbeats; an idle stream is still silent", i)
		}
	}
}
