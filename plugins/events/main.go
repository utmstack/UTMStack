package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/events/search"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
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

	logBytes, err := protojson.Marshal(event)
	if err != nil {
		return nil, err
	}

	logBuffer := bytes.NewBuffer(logBytes)

	jLog := logBuffer.String()

	search.AddToQueue(jLog)

	return nil, nil
}
