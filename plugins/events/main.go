package main

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"io"
	"net"
	"os"

	"google.golang.org/grpc"
)

type analysisServer struct {
	plugins.UnimplementedAnalysisServer
}

func main() {
	filePath, err := utils.MkdirJoin(plugins.WorkDir, "sockets")
	if err != nil {
		_ = catcher.Error("cannot create socket directory", err, nil)
		os.Exit(1)
	}

	socketPath := filePath.FileJoin("com.utmstack.events_analysis.sock")
	_ = os.Remove(socketPath)

	unixAddress, err := net.ResolveUnixAddr("unix", socketPath)
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
