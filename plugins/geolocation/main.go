package main

import (
	"context"
	"net"
	"os"
	"path"

	gosdk "github.com/threatwinds/go-sdk"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
)

type parsingServer struct {
	gosdk.UnimplementedParsingServer
}

func main() {
	_ = os.Remove(path.Join(gosdk.GetCfg().Env.Workdir, "sockets", "com.utmstack.geolocation_parsing.sock"))

	unixAddress, err := net.ResolveUnixAddr("unix", path.Join(gosdk.GetCfg().Env.Workdir, "sockets", "com.utmstack.geolocation_parsing.sock"))
	if err != nil {
		_ = gosdk.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = gosdk.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	gosdk.RegisterParsingServer(grpcServer, &parsingServer{})

	go update()

	if err := grpcServer.Serve(listener); err != nil {
		_ = gosdk.Error("cannot serve grpc", err, nil)
		os.Exit(1)
	}
}

func (p *parsingServer) ParseLog(_ context.Context, transform *gosdk.Transform) (*gosdk.Draft, error) {
	m := gosdk.NewMeter("ParseLog")
	defer m.Elapsed("finished")

	source, ok := transform.Step.Dynamic.Params["source"]
	if !ok {
		return transform.Draft, gosdk.Error("'source' parameter required", nil, nil)
	}

	destination, ok := transform.Step.Dynamic.Params["destination"]
	if !ok {
		return transform.Draft, gosdk.Error("'destination' parameter required", nil, nil)
	}

	value := gjson.Get(transform.Draft.Log, source.GetStringValue()).String()
	if value == "" {
		return transform.Draft, nil
	}

	geo := geolocate(value)

	if geo == nil {
		return transform.Draft, nil
	}

	var err error
	transform.Draft.Log, err = sjson.Set(transform.Draft.Log, destination.GetStringValue(), geo)
	if err != nil {
		return transform.Draft, gosdk.Error("failed to set geolocation", err, nil)
	}

	return transform.Draft, nil
}
