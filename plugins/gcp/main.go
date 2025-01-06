package main

import (
	"github.com/threatwinds/go-sdk/plugins"
	"os"
	"os/signal"
	"syscall"
)

var localNotificationsChannel chan *plugins.Message
var logsQueue = make(chan *plugins.Log)

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "worker" {
		os.Exit(0)
	}

	localNotificationsChannel = make(chan *plugins.Message)
	logsQueue = make(chan *plugins.Log)

	go processLogs()
	go plugins.SendNotificationsFromChannel()

	StartGroupModuleManager()

	// lock main until signal
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs
}
