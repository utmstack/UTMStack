package main

import (
	"os"
	"os/signal"
	"syscall"

	go_sdk "github.com/threatwinds/go-sdk"
)

var localNotificationsChannel chan *go_sdk.Message
var logsQueue = make(chan *go_sdk.Log)

func main() {
	mode := go_sdk.GetCfg().Env.Mode
	if mode != "worker" {
		os.Exit(0)
	}

	go_sdk.Logger().Info("Starting GCP plugin...")

	localNotificationsChannel = make(chan *go_sdk.Message)
	logsQueue = make(chan *go_sdk.Log)

	go processLogs()
	go processNotification()

	StartGroupModuleManager()

	// lock main until signal
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs
}
