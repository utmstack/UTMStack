package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
)

type analysisServer struct {
	go_sdk.UnimplementedAnalysisServer
}

type Config struct {
	Elasticsearch string `yaml:"elasticsearch"`
}

func main() {
	os.Remove(path.Join(
		go_sdk.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.events_analysis.sock"))

	laddr, err := net.ResolveUnixAddr("unix",
		path.Join(go_sdk.GetCfg().Env.Workdir, "sockets",
			"com.utmstack.events_analysis.sock"))

	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	go_sdk.RegisterAnalysisServer(grpcServer, &analysisServer{})

	if err := grpcServer.Serve(listener); err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *analysisServer) Analyze(event *go_sdk.Event, srv grpc.ServerStreamingServer[go_sdk.Alert]) error {
	jLog, e := go_sdk.ToString(event)
	if e != nil {
		return fmt.Errorf("error converting to string during analysis")
	}

	go_sdk.Logger().LogF(100, "received event: %s", *jLog)

	addToQueue(*jLog)

	return io.EOF
}

func indexBuilder(name string, timestamp time.Time) string {
	fst := timestamp.Format("2006.01.02")

	index := fmt.Sprintf(
		"%s-%s",
		name,
		fst,
	)

	return index
}
