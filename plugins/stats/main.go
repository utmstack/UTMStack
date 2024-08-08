package main

import (
	"context"
	"net"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	plugins.UnimplementedNotificationServer
}

type Message struct {
	Cause string `json:"cause"`
	DataType string `json:"data_type"`
	DataSource string `json:"data_source"`
}

func main() {
	os.Remove(path.Join(helpers.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.stats_notification.sock"))

	laddr, err := net.ResolveUnixAddr(
		"unix", path.Join(helpers.GetCfg().Env.Workdir,
			"sockets", "com.utmstack.stats_notification.sock"))

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
	plugins.RegisterNotificationServer(grpcServer, &notificationServer{})

	if err := grpcServer.Serve(listener); err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *notificationServer) Notify(ctx context.Context, msg *plugins.Message) (*emptypb.Empty, error) {
	helpers.Logger().Info("received message: %v", msg.Message)
	
	// implement statistics logic here

	return &emptypb.Empty{}, nil
}
