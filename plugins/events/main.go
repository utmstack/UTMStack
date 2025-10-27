package main

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"io"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

type analysisServer struct {
	plugins.UnimplementedAnalysisServer
}

func main() {
	// Retry logic for initialization
	maxRetries := 3
	retryDelay := 2 * time.Second

	// Initialize with retry logic instead of exiting
	var filePath utils.Folder
	var err error
	var socketPath string
	var unixAddress *net.UnixAddr
	var listener *net.UnixListener

	// Retry loop for creating socket directory
	for retry := 0; retry < maxRetries; retry++ {
		filePath, err = utils.MkdirJoin(plugins.WorkDir, "sockets")
		if err == nil {
			break
		}

		_ = catcher.Error("cannot create socket directory, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when creating socket directory", err, nil)
			return
		}
	}

	socketPath = filePath.FileJoin("com.utmstack.events_analysis.sock")
	_ = os.Remove(socketPath)

	// Retry loop for resolving unix address
	retryDelay = 2 * time.Second
	for retry := 0; retry < maxRetries; retry++ {
		unixAddress, err = net.ResolveUnixAddr("unix", socketPath)
		if err == nil {
			break
		}

		_ = catcher.Error("cannot resolve unix address, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when resolving unix address", err, nil)
			return
		}
	}

	startQueue()

	// Retry loop for listening to unix socket
	retryDelay = 2 * time.Second
	for retry := 0; retry < maxRetries; retry++ {
		listener, err = net.ListenUnix("unix", unixAddress)
		if err == nil {
			break
		}

		_ = catcher.Error("cannot listen to unix socket, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when listening to unix socket", err, nil)
			return
		}
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterAnalysisServer(grpcServer, &analysisServer{})

	// Serve with error handling
	if err := grpcServer.Serve(listener); err != nil {
		_ = catcher.Error("cannot serve grpc", err, nil)
		// Instead of exiting, restart the main function
		time.Sleep(5 * time.Second)
		go main()
		return
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
