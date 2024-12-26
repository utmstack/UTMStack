package main

import (
	"os"
	"os/signal"
	"syscall"

	gosdk "github.com/threatwinds/go-sdk"
)

var localNotificationsChannel chan *gosdk.Message
var logsQueue = make(chan *gosdk.Log)

func main() {
	mode := gosdk.GetCfg().Env.Mode
	if mode != "worker" {
		os.Exit(0)
	}

	localNotificationsChannel = make(chan *gosdk.Message)
	logsQueue = make(chan *gosdk.Log)

	go processLogs()
	go processNotification()

	StartGroupModuleManager()

	// lock main until signal
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs
}
