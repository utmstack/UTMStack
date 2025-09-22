package main

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func sendLog() {
	// Initialize with retry logic instead of exiting
	var socketsFolder utils.Folder
	var err error
	var socketFile string
	var conn *grpc.ClientConn
	var client plugins.EngineClient
	var inputClient plugins.Engine_InputClient

	// Retry loop for initialization
	for {
		socketsFolder, err = utils.MkdirJoin(plugins.WorkDir, "sockets")
		if err != nil {
			_ = catcher.Error("cannot create socket directory", err, nil)
			time.Sleep(5 * time.Second)
			continue
		}

		socketFile = socketsFolder.FileJoin("engine_server.sock")

		conn, err = grpc.NewClient(fmt.Sprintf("unix://%s", socketFile),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			_ = catcher.Error("failed to connect to engine server", err, nil)
			time.Sleep(5 * time.Second)
			continue
		}

		client = plugins.NewEngineClient(conn)

		inputClient, err = client.Input(context.Background())
		if err != nil {
			_ = catcher.Error("failed to create input client", err, nil)
			if conn != nil {
				_ = conn.Close()
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// If we got here, initialization was successful
		break
	}

	defer conn.Close()

	var restart = make(chan bool)

	go func() {
		for {
			l := <-localLogsChannel

			err := inputClient.Send(l)
			if err != nil {
				_ = catcher.Error("failed to send log", err, nil)
				restart <- true
				return
			}
		}
	}()

	go func() {
		for {
			_, err = inputClient.Recv()
			if err != nil {
				_ = catcher.Error("failed to receive ack", err, nil)
				restart <- true
				return
			}
		}
	}()

	select {
	case <-restart:
		time.Sleep(5 * time.Second)
		go sendLog()
		return
	}
}
