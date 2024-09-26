package main

import (
	"context"
	"net"
	"os"
	"path"

	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	go_sdk.UnimplementedNotificationServer
}

type Message struct {
	Cause      *string `json:"cause,omitempty"`
	DataType   string  `json:"dataType"`
	DataSource string  `json:"dataSource"`
}

const (
	TOPIC_ENQUEUE_FAILURE = "enqueue_failure"
	TOPIC_ENQUEUE_SUCCESS = "enqueue_success"
)

func main() {
	os.Remove(path.Join(go_sdk.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.stats_notification.sock"))

	laddr, err := net.ResolveUnixAddr(
		"unix", path.Join(go_sdk.GetCfg().Env.Workdir,
			"sockets", "com.utmstack.stats_notification.sock"))

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
	go_sdk.RegisterNotificationServer(grpcServer, &notificationServer{})

	if err := grpcServer.Serve(listener); err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *notificationServer) Notify(ctx context.Context, msg *go_sdk.Message) (*emptypb.Empty, error) {
	go_sdk.Logger().LogF(100, "%s: %s", msg.Topic, msg.Message)

	// TODO: implement statistics logic here

	return &emptypb.Empty{}, nil
}
