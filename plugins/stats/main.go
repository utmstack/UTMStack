package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/threatwinds/go-sdk/opensearch"
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

var statisticsQueue chan Message
var fails map[string]map[string]map[string]int64
var success map[string]map[string]int64
var failsLock sync.RWMutex
var successLock sync.RWMutex

func main() {
	ctx, cancel := context.WithCancel(context.Background())

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

	statisticsQueue = make(chan Message, 1000)
	fails = make(map[string]map[string]map[string]int64)
	success = make(map[string]map[string]int64)

	grpcServer := grpc.NewServer()
	go_sdk.RegisterNotificationServer(grpcServer, &notificationServer{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(listener); err != nil {
			go_sdk.Logger().ErrorF(err.Error())
			os.Exit(1)
		}
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processStatistics(ctx)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		saveToDB(ctx, "success")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		saveToDB(ctx, "fails")
	}()

	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs

	grpcServer.GracefulStop()
	cancel()
	go_sdk.Logger().Info("shutting down...")
}

func (p *notificationServer) Notify(ctx context.Context, msg *go_sdk.Message) (*emptypb.Empty, error) {
	go_sdk.Logger().LogF(100, "%s: %s", msg.Topic, msg.Message)

	if msg.Topic != TOPIC_ENQUEUE_FAILURE && msg.Topic != TOPIC_ENQUEUE_SUCCESS {
		return &emptypb.Empty{}, nil
	}

	mbytes := []byte(msg.Message)

	var pMsg Message

	err := json.Unmarshal(mbytes, &pMsg)
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		return &emptypb.Empty{}, err
	}

	statisticsQueue <- pMsg

	return &emptypb.Empty{}, nil
}

func processStatistics(ctx context.Context) {
	for {
		select {
		case msg := <-statisticsQueue:
			if msg.Cause == nil {
				successLock.Lock()
				if _, ok := success[msg.DataSource]; !ok {
					success[msg.DataSource] = make(map[string]int64)
				}

				if _, ok := success[msg.DataSource][msg.DataType]; !ok {
					success[msg.DataSource][msg.DataType] = 0
				}

				success[msg.DataSource][msg.DataType]++
				successLock.Unlock()
			} else {
				failsLock.Lock()
				if _, ok := fails[msg.DataSource]; !ok {
					fails[msg.DataSource] = make(map[string]map[string]int64)
				}

				if _, ok := fails[msg.DataSource][msg.DataType]; !ok {
					fails[msg.DataSource][msg.DataType] = make(map[string]int64)
				}

				if _, ok := fails[msg.DataSource][msg.DataType][*msg.Cause]; !ok {
					fails[msg.DataSource][msg.DataType][*msg.Cause] = 0
				}

				fails[msg.DataSource][msg.DataType][*msg.Cause]++
				failsLock.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

type Success struct {
	DataSource string `json:"dataSource"`
	DataType   string `json:"dataType"`
	Count      int64  `json:"count"`
}

type Fail struct {
	DataSource string `json:"dataSource"`
	DataType   string `json:"dataType"`
	Cause      string `json:"cause"`
	Count      int64  `json:"count"`
}

func saveToDB(ctx context.Context, t string) {
	for {
		select {
		case <-time.After(1 * time.Minute):
			sendStatistic(t)
		case <-ctx.Done():
			return
		}
	}
}

func extractSuccess() []Success {
	successLock.Lock()
	defer successLock.Unlock()

	var result []Success

	for dataSource, dataTypes := range success {
		for dataType, count := range dataTypes {
			result = append(result, Success{
				DataSource: dataSource,
				DataType:   dataType,
				Count:      count,
			})
		}
	}

	success = make(map[string]map[string]int64)

	return result
}

func extractFails() []Fail {
	failsLock.Lock()
	defer failsLock.Unlock()

	var result []Fail

	for dataSource, dataTypes := range fails {
		for dataType, causes := range dataTypes {
			for cause, count := range causes {
				result = append(result, Fail{
					DataSource: dataSource,
					DataType:   dataType,
					Cause:      cause,
					Count:      count,
				})
			}
		}
	}

	fails = make(map[string]map[string]map[string]int64)

	return result
}

func sendStatistic(t string) {
	switch t {
	case "success":
		success := extractSuccess()
		for _, s := range success {
			saveToOpensearch(s, fmt.Sprintf("statistics-success-%s", time.Now().Format("2006.01")))
		}

	case "fails":
		fails := extractFails()
		for _, f := range fails {
			saveToOpensearch(f, fmt.Sprintf("statistics-fails-%s", time.Now().Format("2006.01")))
		}
	}
}

func saveToOpensearch[Data any](data Data, index string) {
	oCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := opensearch.IndexDoc(oCtx, data, index, uuid.NewString())
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
	}
}
