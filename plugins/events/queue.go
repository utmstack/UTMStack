package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	ch "github.com/threatwinds/go-sdk/store/clickhouse"
	"github.com/tidwall/gjson"
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

func startQueue(ctx context.Context) {
	writer := connect()

	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		pending := 0
		flush := func() {
			if pending == 0 {
				return
			}
			if err := writer.Flush(ctx); err != nil {
				_ = catcher.Error("failed to write logs", err, map[string]any{
					"process": processName,
					"rows":    pending,
				})
			}
			pending = 0
		}

		for {
			select {
			case <-ctx.Done():
				flush()
				_ = writer.Close(context.WithoutCancel(ctx))
				return

			case l := <-logs:
				tenant := gjson.Get(l, "tenantId").String()
				if tenant == "" {
					_ = catcher.Error("dropping log with no tenant", nil, map[string]any{
						"process": processName,
						"id":      gjson.Get(l, "id").String(),
					})
					continue
				}

				scope := store.Scope{Tenant: tenant, Dataset: datasetLogs}
				if err := writer.Write(scope, []byte(l)); err != nil {
					_ = catcher.Error("failed to queue log", err, map[string]any{"process": processName})
					continue
				}

				if pending++; pending >= flushThreshold {
					flush()
				}

			case <-ticker.C:
				flush()
			}
		}
	}()
}

func connect() store.BulkWriter {
	cfg := plugins.PluginCfg("clickhouse")

	table := cfg.Get("logsTable").String()
	if table == "" {
		table = defaultLogsTable
	}

	driver, err := ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetLogs: cfg.Get("logsTable").String(),
		},
		TenantColumn:   "tenantId",
		TimeColumn:     "@timestamp",
		DataTypeColumn: "dataType",
		// A row ClickHouse still refuses after BulkWriter's own retries and
		// bisection is one this batch was never going to write — logged with
		// its own content and the real error, instead of vanishing into the
		// "rows: N" count a whole failed Flush used to be the only trace of.
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

	writer, err := driver.BulkWriter(datasetLogs)
	if err != nil {
		_ = catcher.Error("cannot open the log writer", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		panic(err)
	}
	return writer
}
