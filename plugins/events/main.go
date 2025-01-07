package main

import (
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"io"
	"net"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
)

type analysisServer struct {
	plugins.UnimplementedAnalysisServer
}

func main() {
	_ = os.Remove(path.Join(plugins.GetCfg().Env.Workdir, "sockets", "com.utmstack.events_analysis.sock"))

	unixAddress, err := net.ResolveUnixAddr("unix", path.Join(plugins.GetCfg().Env.Workdir, "sockets",
		"com.utmstack.events_analysis.sock"))
	if err != nil {
		_ = catcher.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	startQueue()

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = catcher.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterAnalysisServer(grpcServer, &analysisServer{})

	if err := grpcServer.Serve(listener); err != nil {
		_ = catcher.Error("cannot serve grpc", err, nil)
		os.Exit(1)
	}
}

func (p *analysisServer) Analyze(event *plugins.Event, _ grpc.ServerStreamingServer[plugins.Alert]) error {
	jLog, err := utils.ToString(event)
	if err != nil {
		return catcher.Error("cannot convert event to json", err, nil)
	}

	addToQueue(*jLog)

	return io.EOF
}

func indexBuilder(name string, timestamp time.Time) string {
	return fmt.Sprintf("%s-%s", name, timestamp.Format("2006.01.02"))
}
