package main

import (
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
)

var localNotificationsChannel chan *plugins.Message
var logsQueue = make(chan *plugins.Log)

func main() {
	helpers.Logger().Info("Starting GCP plugin...")

	localNotificationsChannel = make(chan *plugins.Message)
	logsQueue = make(chan *plugins.Log)

	go processLogs()
	go processNotification()

	StartGroupModuleManager()
}
