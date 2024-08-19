package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/events/search"
	"google.golang.org/grpc"
)

type analysisServer struct {
	plugins.UnimplementedAnalysisServer
}

func main() {
	os.Remove(path.Join(
		helpers.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.events_analysis.sock"))

	laddr, err := net.ResolveUnixAddr("unix",
		path.Join(helpers.GetCfg().Env.Workdir, "sockets",
			"com.utmstack.events_analysis.sock"))

	if err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterAnalysisServer(grpcServer, &analysisServer{})

	if err := grpcServer.Serve(listener); err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *analysisServer) Analyze(ctx context.Context,
	event *plugins.Event) (*plugins.Alert, error) {

	jLog, e := helpers.ToString(event)
	if e != nil {
		return nil, fmt.Errorf(e.Message)
	}

	helpers.Logger().LogF(100, "received event: %s", *jLog)

	search.AddToQueue(*jLog)

	return nil, nil
}
