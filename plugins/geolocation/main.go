package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
)

type parsingServer struct {
	plugins.UnimplementedParsingServer
}

func main() {
	// Retry logic for creating socket directory
	maxRetries := 3
	retryDelay := 2 * time.Second

	var filePath utils.Folder
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		filePath, err = utils.MkdirJoin(plugins.WorkDir, "sockets")
		if err == nil {
			break
		}

		_ = catcher.Error("cannot create directory, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when creating directory", err, nil)
			return
		}
	}

	socketPath := filePath.FileJoin("com.utmstack.geolocation_parsing.sock")
	_ = os.Remove(socketPath)

	// Retry logic for resolving unix address
	retryDelay = 2 * time.Second
	var unixAddress *net.UnixAddr

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

	// Retry logic for listening to unix socket
	retryDelay = 2 * time.Second
	var listener *net.UnixListener

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
	plugins.RegisterParsingServer(grpcServer, &parsingServer{})

	go loadGeolocationData()

	// Serve with error handling and retry logic
	for {
		err := grpcServer.Serve(listener)
		if err == nil {
			break
		}

		_ = catcher.Error("cannot serve grpc, retrying", err, nil)
		time.Sleep(5 * time.Second)
	}
}

func (p *parsingServer) ParseLog(_ context.Context, transform *plugins.Transform) (*plugins.Draft, error) {
	source, ok := transform.Step.Dynamic.Params["source"]
	if !ok {
		return transform.Draft, catcher.Error("'source' parameter required", nil, nil)
	}

	destination, ok := transform.Step.Dynamic.Params["destination"]
	if !ok {
		return transform.Draft, catcher.Error("'destination' parameter required", nil, nil)
	}

	sourceField := source.GetStringValue()
	utils.SanitizeField(&sourceField)

	err := utils.ValidateReservedField(sourceField, false)
	if err != nil {
		return transform.Draft, catcher.Error("cannot parse log", err, map[string]any{
			"field": sourceField,
		})
	}

	destinationField := destination.GetStringValue()
	utils.SanitizeField(&destinationField)

	err = utils.ValidateReservedField(destinationField, false)
	if err != nil {
		return transform.Draft, catcher.Error("cannot parse log", err, map[string]any{
			"field": destinationField,
		})
	}

	value := gjson.Get(transform.Draft.Log, sourceField).String()
	if value == "" {
		return transform.Draft, nil
	}

	geo := geolocate(value)

	if geo == nil {
		return transform.Draft, nil
	}

	transform.Draft.Log, err = sjson.Set(transform.Draft.Log, destinationField, geo)
	if err != nil {
		return transform.Draft, catcher.Error("failed to set geolocation", err, nil)
	}

	return transform.Draft, nil
}
