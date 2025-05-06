package processor

import (
	"net"
	"os"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/opensearch"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/grpc"
)

type socAiServer struct {
	plugins.UnimplementedCorrelationServer
}

func (p *Processor) socAiServer() {
	socketsFolder, err := utils.MkdirJoin(plugins.WorkDir, "sockets")
	if err != nil {
		_ = catcher.Error("cannot create socket directory", err, nil)
		os.Exit(1)
	}

	socketFile := socketsFolder.FileJoin("com.utmstack.soc_ai_correlation.sock")
	_ = os.Remove(socketFile)

	unixAddress, err := net.ResolveUnixAddr("unix", socketFile)
	if err != nil {
		_ = catcher.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = catcher.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterCorrelationServer(grpcServer, &socAiServer{})

	osUrl := plugins.PluginCfg("com.utmstack", false).Get("opensearch").String()
	err = opensearch.Connect([]string{osUrl})
	if err != nil {
		_ = catcher.Error("cannot connect to OpenSearch", err, nil)
		os.Exit(1)
	}

	if err := grpcServer.Serve(listener); err != nil {
		_ = catcher.Error("cannot serve grpc", err, nil)
		os.Exit(1)
	}
}
