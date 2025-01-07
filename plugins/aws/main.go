package main

import (
	"context"
	"fmt"

	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const (
	delayCheck    = 300
	defaultTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

var logQueue = make(chan *plugins.Log, 100*runtime.NumCPU())

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for t := 0; t < runtime.NumCPU(); t++ {
		go processLogs()
	}

	st := time.Now().Add(-600 * time.Second)
	for {
		utmConfig := plugins.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		et := st.Add(299 * time.Second)
		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = catcher.Error("cannot obtain module configuration", err, nil)
			}

			time.Sleep(time.Second * delayCheck)
			st = et.Add(1)
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							skip = true
							break
						}
					}

					if !skip {
						pull(st, et, group)
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		}

		time.Sleep(time.Second * delayCheck)
		st = et.Add(1)
	}
}

func pull(startTime time.Time, endTime time.Time, group types.ModuleGroup) {
	agent := GetAWSProcessor(group)

	logs, err := agent.GetLogs(startTime, endTime)
	if err != nil {
		_ = catcher.Error("cannot get logs", err, map[string]any{
			"startTime": startTime,
			"endTime":   endTime,
			"group":     group.GroupName,
		})
		return
	}

	for _, log := range logs {
		logQueue <- &plugins.Log{
			Id:         uuid.NewString(),
			TenantId:   defaultTenant,
			DataType:   "aws",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		}
	}
}

func processLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s",
		path.Join(plugins.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = catcher.Error("cannot connect to engine server", err, nil)
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		_ = catcher.Error("cannot create input client", err, nil)
		// TODO: notify engine about this error before exit
		os.Exit(1)
	}

	// TODO: implement retry system using ack and timeouts
	go func() {
		for {
			log := <-logQueue
			err := inputClient.Send(log)
			if err != nil {
				_ = catcher.Error("cannot send log", err, nil)
				// TODO: notify engine about this error before exit
				os.Exit(1)
			}
		}
	}()

	for {
		_, err = inputClient.Recv()
		if err != nil {
			_ = catcher.Error("cannot receive ack", err, nil)
			// TODO: notify engine about this error before exit
			os.Exit(1)
		}
	}
}
