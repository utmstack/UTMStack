package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/factory"
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/plugins/events/internal/storecfg"
)

const (
	processName      = "plugin_com.utmstack.events"
	datasetLogs      = store.Dataset("logs")
	flushInterval    = 5 * time.Second
	flushThreshold   = 1000
	defaultLogsTable = "logs"
)

var logs = make(chan string, 100*runtime.NumCPU())

func addToQueue(l string) {
	if len(logs) >= 100*runtime.NumCPU() {
		_ = catcher.Error("cannot enqueue log", fmt.Errorf("queue is full"), map[string]any{
			"process": processName,
			"queue":   "logs",
		})

		return
	}

	logs <- l
}

func startQueue(ctx context.Context) <-chan struct{} {
	writer, built := connect()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer built.Close()
		runQueue(ctx, logs, writer)
	}()
	return done
}

func runQueue(ctx context.Context, queue <-chan string, writer store.BulkWriter) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	pending := 0
	flush := func(flushCtx context.Context) {
		if pending == 0 {
			return
		}
		if err := writer.Flush(flushCtx); err != nil {
			_ = catcher.Error("failed to write logs", err, map[string]any{"process": processName, "rows": pending})
		}
		pending = 0
	}
	for {
		select {
		case <-ctx.Done():
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			flush(closeCtx)
			_ = writer.Close(closeCtx)
			cancel()
			return
		case l := <-queue:
			tenant := gjson.Get(l, "tenantId").String()
			if tenant == "" {
				_ = catcher.Error("dropping log with no tenant", nil, map[string]any{"process": processName, "id": gjson.Get(l, "id").String()})
				continue
			}
			scope := store.Scope{Tenant: tenant, Dataset: datasetLogs}
			if err := writer.Write(scope, []byte(l)); err != nil {
				_ = catcher.Error("failed to queue log", err, map[string]any{"process": processName})
				continue
			}
			if pending++; pending >= flushThreshold {
				flush(ctx)
			}
		case <-ticker.C:
			flush(ctx)
		}
	}
}

func connect() (store.BulkWriter, *factory.Built) {
	cfg, err := storecfg.New(plugins.PluginCfg("store"), plugins.PluginCfg("clickhouse"), map[string]storecfg.Table{
		string(datasetLogs): {ConfigKey: "logsTable", Default: defaultLogsTable},
	})
	if err == nil {
		cfg, err = storecfg.ClickHousePrimary(cfg)
	}
	if err != nil {
		panic(err)
	}
	built, err := factory.New(cfg, factory.Options{
		OnReject: func(row []byte, err error) {
			_ = catcher.Error("dropped a log ClickHouse will not accept", err, map[string]any{
				"process": processName,
				"id":      gjson.GetBytes(row, "id").String(),
				"row":     string(row),
			})
		},
	})
	if err != nil {
		_ = catcher.Error("cannot build the ClickHouse store", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		panic(err)
	}

	writer, err := built.Store.BulkWriter(datasetLogs)
	if err != nil {
		_ = built.Close()
		_ = catcher.Error("cannot open the log writer", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		panic(err)
	}
	return writer, built
}
