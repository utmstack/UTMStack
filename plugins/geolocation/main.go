package main

import (
	"context"
	"net"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
)

type parsingServer struct {
	plugins.UnimplementedParsingServer
}

func main() {
	os.Remove(path.Join(helpers.GetCfg().Env.Workdir, "sockets", "com.utmstack.geolocation_parsing.sock"))

	laddr, err := net.ResolveUnixAddr("unix", path.Join(helpers.GetCfg().Env.Workdir, "sockets", "com.utmstack.geolocation_parsing.sock"))
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
	plugins.RegisterParsingServer(grpcServer, &parsingServer{})

	go update()

	if err := grpcServer.Serve(listener); err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *parsingServer) ParseLog(ctx context.Context, jLog *plugins.JLog) (*plugins.JLog, error) {
	fromIp := gjson.Get(jLog.Log, "from.ip").String()
	if fromIp != "" {
		geo := geolocate(fromIp)
		jLog.Log, _ = sjson.Set(jLog.Log, "from.geolocation", geo)
	}

	remoteIp := gjson.Get(jLog.Log, "remote.ip").String()
	if remoteIp != "" {
		geo := geolocate(remoteIp)
		jLog.Log, _ = sjson.Set(jLog.Log, "remote.geolocation", geo)
	}

	toIp := gjson.Get(jLog.Log, "to.ip").String()
	if toIp != "" {
		geo := geolocate(toIp)
		jLog.Log, _ = sjson.Set(jLog.Log, "to.geolocation", geo)
	}

	localIp := gjson.Get(jLog.Log, "local.ip").String()
	if localIp != "" {
		geo := geolocate(localIp)
		jLog.Log, _ = sjson.Set(jLog.Log, "local.geolocation", geo)
	}

	return jLog, nil
}
