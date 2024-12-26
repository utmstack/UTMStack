package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	gosdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
)

type analysisServer struct {
	gosdk.UnimplementedAnalysisServer
}

func main() {
	_ = os.Remove(path.Join(gosdk.GetCfg().Env.Workdir, "sockets", "com.utmstack.events_analysis.sock"))

	unixAddress, err := net.ResolveUnixAddr("unix", path.Join(gosdk.GetCfg().Env.Workdir, "sockets",
		"com.utmstack.events_analysis.sock"))
	if err != nil {
		_ = gosdk.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	startQueue()

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = gosdk.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	gosdk.RegisterAnalysisServer(grpcServer, &analysisServer{})

	if err := grpcServer.Serve(listener); err != nil {
		_ = gosdk.Error("cannot serve grpc", err, nil)
		os.Exit(1)
	}
}

func (p *analysisServer) Analyze(event *gosdk.Event, _ grpc.ServerStreamingServer[gosdk.Alert]) error {
	jLog, err := gosdk.ToString(event)
	if err != nil {
		return gosdk.Error("cannot convert event to json", err, nil)
	}

	addToQueue(*jLog)

	return io.EOF
}

func indexBuilder(name string, timestamp time.Time) string {
	return fmt.Sprintf("%s-%s", name, timestamp.Format("2006.01.02"))
}
