package dispatcher

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/threatwinds/go-sdk/plugins"
	amgr "github.com/utmstack/UTMStack/agent-manager/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const internalKeyHeader = "internal-key"

var errNoAddress = errors.New("agent-manager host/port not configured")

func dial() (amgr.PanelServiceClient, amgr.AgentServiceClient, *grpc.ClientConn, error) {
	cfg := plugins.PluginCfg("com.utmstack").Get("soar.agentManager")
	host := cfg.Get("host").String()
	port := cfg.Get("port").Int()
	if host == "" || port == 0 {
		return nil, nil, nil, errNoAddress
	}

	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(unaryAuth),
		grpc.WithStreamInterceptor(streamAuth),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return amgr.NewPanelServiceClient(conn), amgr.NewAgentServiceClient(conn), conn, nil
}

func attachAuth(ctx context.Context) context.Context {
	key := os.Getenv("INTERNAL_KEY")
	if key == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, internalKeyHeader, key)
}

func unaryAuth(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(attachAuth(ctx), method, req, reply, cc, opts...)
}

func streamAuth(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
	method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(attachAuth(ctx), desc, cc, method, opts...)
}
